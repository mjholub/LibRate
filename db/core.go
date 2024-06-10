package db

import (
	"context"
	"database/sql/driver"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	lop "github.com/samber/lo/parallel"

	"github.com/avast/retry-go/v4"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db/bootstrap"
)

func CreateDsn(dsn *cfg.DBConfig) string {
	switch dsn.SSL {
	case "require", "verify-ca", "verify-full", "disable":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, dsn.SSL)
		return data
	case "prefer":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, "require")
		return data
	case "unknown":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		return data
	default:
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		return data
	}
}

// Currently only postgres is supported, as we try to
// stick to YAGNI and convetion over configuration.
func Connect(ctx context.Context, dsn string, attempts int32) (*pgxpool.Pool, error) {
	var db *pgxpool.Pool

	if attempts < 0 {
		attempts = 2147483647
	}
	// ensure default value is actually loaded
	if attempts == 0 {
		attempts = 15
	}

	err := retry.Do(
		func() error {
			var err error
			db, err = pgxpool.New(ctx, dsn)
			return err
		},
		retry.Attempts(uint(attempts)),
		retry.Delay(1*time.Second), // Delay between retries
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf("Attempt %d to connect to database failed: %v; retrying...",
				n, err)
		},
		),
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createExtension(db *pgxpool.Pool, extName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// nolint:gocritic // postgres doesn't parse "%q" properly
	_, err := db.Exec(ctx,
		fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, extName))
	if err != nil {
		return fmt.Errorf("failed to create extension %s: %w", extName, err)
	}

	return nil
}

func InitDB(conf *cfg.DBConfig, log *zerolog.Logger) error {
	dsn := CreateDsn(conf)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := Connect(ctx, dsn, conf.RetryAttempts)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS public; SET search_path TO public;`)
	if err != nil {
		return fmt.Errorf("failed to create public schema: %w", err)
	}
	err = bootstrap.Schemas(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created schemas")

	// set up the extensions
	var mu sync.Mutex
	errChan := make(chan error)
	mu.Lock()
	extNames := []string{"pgcrypto", "uuid-ossp", "pg_trgm"}
	log.Info().Msg("Creating extensions...")
	for i := range extNames {
		go func(i int) {
			errChan <- createExtension(db, extNames[i])
		}(i)
	}
	for i := 0; i < len(extNames); i++ {
		err = <-errChan
		if err != nil {
			return err
		}
	}
	close(errChan)
	mu.Unlock()
	log.Info().Msg("Created extensions")
	type seedFn struct {
		fn   func(context.Context, *pgxpool.Pool) error
		name string
	}

	seedFns := []func(context.Context, *pgxpool.Pool) error{
		bootstrap.CDN, bootstrap.Places, bootstrap.MediaCore, bootstrap.People,
		bootstrap.Roles, bootstrap.MediaCreators, bootstrap.PeopleMeta, bootstrap.Media,
		bootstrap.CreatorGroups, bootstrap.AlbumArtists, bootstrap.Studio, bootstrap.Books,
		bootstrap.Cast, bootstrap.Members, bootstrap.MembersProfilePic, bootstrap.Review}

	seedFnNames := []string{
		"cdn", "places", "media_core", "people", "roles", "media_creators", "people_meta",
		"media", "creator_groups", "album_artists", "studio", "books", "cast", "members",
		"members_profilepic", "review"}

	for i := range seedFns {
		err = seedFns[i](ctx, db)
		if err != nil {
			return fmt.Errorf("failed to create %s tables: %w", seedFnNames[i], err)
		}
		log.Info().Msgf("Created %s tables", seedFnNames[i])
	}

	return nil
}

func TxErr(action string, input interface{}, err error) error {
	return fmt.Errorf("failed to init transaction to %s, with input: %+v: %w",
		action, input, err)
}

func convertParam(param interface{}) (driver.Value, error) {
	switch v := param.(type) {
	case string:
		return v, nil
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return v, nil
	case bool:
		return v, nil
	case []byte:
		return v, nil
	case time.Time:
		return v, nil
	case uuid.UUID:
		return v, nil
	case []string:
		return pgtype.FlatArray[string](param.([]string)), nil
	case []int:
		return pgtype.FlatArray[int](param.([]int)), nil
	default:
		return nil, fmt.Errorf("unsupported type %T for conversion", v)
	}
}

func convertParams(params ...any) ([]driver.Value, error) {
	var args []driver.Value

	for _, p := range params {
		v, err := convertParam(p)
		if err != nil {
			return nil, fmt.Errorf("error converting parameter: %w", err)
		}
		args = append(args, v)
	}

	return args, nil
}

// like SerializableParametrizedTx, but doesn't scan into anything,
// instead simply returning an error
func SerialParametrizedUnaryTx(
	ctx context.Context,
	conn *pgxpool.Pool,
	qName, sql string,
	errorInput any,
	params ...any,
) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return TxErr(qName, errorInput, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, qName, sql)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	args, err := convertParams(params...)

	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	defer close(errCh)

	lop.ForEach(args, func(v driver.Value, _ int) {
		rows, err := tx.Query(ctx, qName, v)
		if err != nil {
			errCh <- fmt.Errorf("error executing query: %w", err)
			return
		}
		defer rows.Close()
	})

	select {
	case err = <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		if err = tx.Commit(ctx); err != nil {
			return TxErr(qName, errorInput, err)
		}
		return nil
	}
}

func SerializableParametrizedTx[T any](
	ctx context.Context,
	conn *pgxpool.Pool,
	qName, sql string,
	errorHandlerInput any,
	params ...any) (dest []T, err error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})

	if err != nil {
		return dest, TxErr(qName, errorHandlerInput, err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, qName, sql)
	if err != nil {
		return dest, fmt.Errorf("error preparing statement: %w", err)
	}

	args, err := convertParams(params...)
	if err != nil {
		return dest, err
	}

	errCh := make(chan error, 1)
	defer close(errCh)

	lop.ForEach(args, func(v driver.Value, _ int) {
		rows, err := tx.Query(ctx, qName, v)
		if err != nil {
			errCh <- fmt.Errorf("error executing query: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var d T
			if err = rows.Scan(&d); err != nil {
				errCh <- fmt.Errorf("error scanning row: %w", err)
				return
			}
			dest = append(dest, d)
		}
	})

	select {
	case err = <-errCh:
		return dest, err
	case <-ctx.Done():
		return dest, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		if err = tx.Commit(ctx); err != nil {
			return dest, TxErr(qName, errorHandlerInput, err)
		}
		return dest, nil
	}
}
