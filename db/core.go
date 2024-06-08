package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

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
	err = bootstrap.CDN(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create cdn tables: %w", err)
	}
	log.Info().Msg("Created cdn tables")
	err = bootstrap.Places(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created places tables")
	err = bootstrap.MediaCore(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Creating media tables: 1/2...")
	err = bootstrap.People(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created people tables")
	err = bootstrap.Roles(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created roles tables")
	err = bootstrap.MediaCreators(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created media_creators tables")
	err = bootstrap.PeopleMeta(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created people_meta tables")
	err = bootstrap.Media(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Creating media tables complete")
	err = bootstrap.CreatorGroups(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created creator_groups tables")
	err = bootstrap.AlbumArtists(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created album_artists tables")
	err = bootstrap.Studio(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created studio tables")
	err = bootstrap.Books(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created books tables")
	err = bootstrap.Cast(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created cast tables")
	err = bootstrap.Members(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created members tables")
	err = bootstrap.MembersProfilePic(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created members profilepic tables")
	err = bootstrap.Review(ctx, db)
	if err != nil {
		return err
	}
	log.Info().Msg("Created review tables")

	return nil
}

func TxErr(action string, input interface{}, err error) error {
	return fmt.Errorf("failed to init transaction to %s, with input: %+v: %w",
		action, input, err)
}

func SerializableParametrizedTx(
	ctx context.Context,
	conn *pgxpool.Pool,
	qName, sql string,
	errorHandlerInput any,
	params ...any) (dest interface{ any | []any }, err error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return nil, TxErr(qName, errorHandlerInput, err)
	}

	// nolint:errcheck // we don't care about the error here
	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, qName, sql)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := conn.Query(ctx, qName, params)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	if err = rows.Scan(&dest); err != nil {
		return nil, fmt.Errorf("error scanning rows: %w", err)
	}

	return dest, nil
}
