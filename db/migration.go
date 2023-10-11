package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
)

func Migrate(conf *cfg.Config, path string) error {
	if conf.AutoMigrate || !conf.AutoMigrate {
		if err := runMigrations(conf, path); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		if conf.ExitAfterMigration {
			// nolint:revive
			os.Exit(0)
		}
	}
	return nil
}

// TODO: find use for the auto parameter or remove it
func runMigrations(conf *cfg.Config, path string) error {
	dbConn := CreateDsn(&conf.DBConfig)

	// connect to database
	conn, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	// map of migration files and their queries
	faultyQueries := make([]map[string]string, 0)
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			migrationSQL, err := os.ReadFile(path)
			if err != nil {
				faultyQueries = append(faultyQueries, map[string]string{info.Name(): err.Error()})
				return fmt.Errorf("failed to read migration file %s: %w", info.Name(), err)
			}
			if _, err := conn.Exec(string(migrationSQL)); err != nil {
				faultyQueries = append(faultyQueries, map[string]string{info.Name(): err.Error()})
				return fmt.Errorf("failed to run migration %s: %w", info.Name(), err)
			}
		}
		if len(faultyQueries) > 0 {
			return fmt.Errorf("failed to run migrations: %v", faultyQueries)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
