package db

import (
	"context"
	"fmt"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, conf.Engine, data)
	if err != nil {
		return nil, err
	}
	return db, nil
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
	// set up the public schema
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto" SCHEMA public;`)
	if err != nil {
		return fmt.Errorf("failed to create pgcrypto extension: %w", err)
	}
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;`)
	if err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "pg_trgm" SCHEMA public;`)
	if err != nil {
		return fmt.Errorf("failed to create pg_trgm extension: %w", err)
	}
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
