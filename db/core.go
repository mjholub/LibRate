package db

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db/bootstrap"
)

func CreateDsn(dsn *cfg.DBConfig) string {
	switch dsn.SSL {
	case "require", "verify-ca", "verify-full", "disable":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, dsn.SSL)
		fmt.Println(data)
		return data
	case "prefer":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, "require")
		fmt.Println(data)
		return data
	case "unknown":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		fmt.Println(data)
		return data
	default:
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		fmt.Println(data)
		return data
	}
}

func Connect(conf *cfg.Config) (*sqlx.DB, error) {
	data := CreateDsn(&conf.DBConfig)
	var db *sqlx.DB

	err := retry.Do(
		func() error {
			var err error
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			db, err = sqlx.ConnectContext(ctx, conf.Engine, data)
			return err
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second), // Delay between retries
		retry.OnRetry(func(n uint, _ error) {
			fmt.Printf("Attempt %d failed; retrying...", n)
		}),
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createUniversalExtension(db *sqlx.DB, extNames ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `SELECT schema_name
	FROM information_schema.schemata
	WHERE schema_name NOT LIKE 'pg_%' AND schema_name != 'information_schema';`)
	if err != nil {
		return fmt.Errorf("failed to query schema names: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		err = rows.Scan(&schemaName)
		if err != nil {
			return fmt.Errorf("failed to scan schema name: %w", err)
		}
		for i := range extNames {
			_, err = db.ExecContext(ctx,
				fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA "%s";`, extNames[i], schemaName))
			if err != nil {
				return fmt.Errorf("failed to create extension %s in schema %s: %w", extNames[i], schemaName, err)
			}
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over schema names: %w", err)
	}

	return nil
}

func InitDB() error {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	db, err := Connect(&conf)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS public;`)
	if err != nil {
		return fmt.Errorf("failed to create public schema: %w", err)
	}
	// set up the extensions
	if err = createUniversalExtension(db, "pgcrypto", "uuid-ossp", "pg_trgm"); err != nil {
		return fmt.Errorf("failed to create database extensions: %w", err)
	}
	/* TODO: verify whether use sequential UUIDs or just ints
	* if err = createUniversalExtension(db, "pgsequentialuuid"); err != nil {
	* return fmt.Errorf("failed to create database extensions: %w", err)
	* }
	 */
	/* postgres 15 no longer supports pg_atoi
	_, err = db.ExecContext(ctx, "CREATE EXTENSION uint;")
	if err != nil {
		return fmt.Errorf("failed to create uint extension: %w", err)
	}
	*/
	err = bootstrap.CDN(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create cdn tables: %w", err)
	}
	err = bootstrap.Places(ctx, db)
	if err != nil {
		return err
	}
	err = bootstrap.MediaCore(ctx, db)
	if err != nil {
		return err
	}
	err = bootstrap.People(ctx, db)
	if err != nil {
		return err
	}
	err = bootstrap.Media(ctx, db)
	if err != nil {
		return err
	}
	err = bootstrap.Members(ctx, db)
	if err != nil {
		return err
	}
	err = bootstrap.Review(ctx, db)
	if err != nil {
		return err
	}

	return nil
}
