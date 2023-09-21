package db

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
)

func Migrate(conf *cfg.Config) (err error) {
	autoMigrate := flag.Bool("auto-migrate", false, "Automatically run migrations")
	exit := flag.Bool("exit", false, "Exit after running migrations")
	flag.Parse()

	if *autoMigrate {
		// run migrations	in db/migrations
		// NOTE: paral
		if err := runMigrations(conf, true); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		if *exit {
			os.Exit(0)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func runMigrations(conf *cfg.Config, auto bool) error {
	dbConn := CreateDsn(&conf.DBConfig)
	path := flag.String("path", "db/migrations", "Path to migrations")
	dsn := flag.String("database", dbConn, "Database connection string")
	flag.Parse()

	if *dsn != "" {
		dbConn = *dsn
	}

	// connect to database
	conn, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	// map of migration files and their queries
	faultyQueries := make([]map[string]string, 0)
	err = filepath.Walk(*path, func(path string, info os.FileInfo, err error) error {
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
