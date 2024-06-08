package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTearDown is a helper function to clean up the test database
func DBTearDown(conf *cfg.Config) error {
	if os.Getenv("CLEANUP_TEST_DB") == "0" {
		return nil
	}
	database := conf.DBConfig
	dsn := CreateDsn(&database)
	var cleanTables *pgxpool.Pool
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cleanTables, err = pgxpool.New(ctx, dsn)
	defer cleanTables.Close()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer cleanTables.Close()
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS cdn CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS places CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS media CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS people CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS reviews CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP SCHEMA IF EXISTS members CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	// delete the extensions
	_, err = cleanTables.Exec(ctx, "DROP EXTENSION IF EXISTS pgcrypto CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP EXTENSION IF EXISTS \"uuid-ossp\" CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP EXTENSION IF EXISTS pg_trgm CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP EXTENSION IF EXISTS sequential_uuids CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	// cleanup custom types
	_, err = cleanTables.Exec(ctx, "DROP TYPE IF EXISTS places.place_kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP TYPE IF EXISTS media.kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP TYPE IF EXISTS people.role CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec(ctx, "DROP TYPE IF EXISTS people.group_kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	return nil
}

func teardownErr(err error) error {
	return fmt.Errorf("failed to clean up database: %w", err)
}
