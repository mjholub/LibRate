package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

// Migrate runs the migrations located in the migrations folder
// paths is a variadic argument, meaning that if no specific path(s)
// are provided, the program will try to run all of them
// Otherwise the arguments to path should only include the containing
// directory name for each migration, e.g.
// "000001-fix-missing-timestamps"
func Migrate(ctx context.Context, log *zerolog.Logger, conf *cfg.Config, paths ...string) error {
	dsn := CreateDsn(&conf.DBConfig)
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer conn.Close()

	if paths == nil {
		// list all directories in migrations folder
		// then loop through them and run the migrations
		dirsWithFiles, err := getDir("", conf.MigrationsPath)
		if err != nil {
			return err
		}
		log.Info().Msgf("found %d migrations", len(dirsWithFiles))
		// iterate through the files in the directory
		for dir, files := range dirsWithFiles {
			log.Info().Msgf("running migration %s", dir.Name())
			dirPath := dir.Name()
			migrationNames := getMigrationNames(files)
			if err := migrateUp(ctx, log, conn, conf.MigrationsPath, dirPath, migrationNames); err != nil {
				return err
			}
		}
		return nil
	} else {
		for i := range paths {
			log.Info().Msgf("running migration %s", paths[i])
			files, err := os.ReadDir(filepath.Join(conf.MigrationsPath, paths[i]))
			if err != nil {
				return fmt.Errorf("error reading filesystem: %v", err)
			}
			f := getMigrationNames(files)
			if err := migrateUp(ctx, log, conn, conf.MigrationsPath, paths[i], f); err != nil {
				return err
			}
		}
		return nil
	}
}

// getDir is an easy way to stat a path that should
// work in both container (absolute path /app/data/migrations/...)
// and with relative paths (./migrations/)
func getDir(basePath, migrationsPath string) (dirsWithFiles map[os.DirEntry][]os.DirEntry, err error) {
	// nolint: gocritic // must use absolute path
	dirs, err := os.ReadDir(filepath.Join(migrationsPath, basePath))
	if err != nil {
		return nil, fmt.Errorf("error reading filesystem: %v", err)
	}

	dirsWithFiles = make(map[os.DirEntry][]os.DirEntry)

	for i := range dirs {
		if dirs[i].IsDir() {
			// map the directory name to it's contents
			dirsWithFiles[dirs[i]], err = os.ReadDir(
				filepath.Join(migrationsPath, basePath, dirs[i].Name()))
			if err != nil {
				return nil, fmt.Errorf("error reading filesystem: %v", err)
			}
		}
	}
	return dirsWithFiles, nil
}

func getMigrationNames(files []os.DirEntry) (migrationNames []string) {
	for i := range files {
		if files[i].IsDir() {
			continue
		}
		migrationChunks := strings.Split(files[i].Name(), ".")
		migrationNames = append(migrationNames, migrationChunks[0])
	}

	return migrationNames
}

func migrateUp(
	ctx context.Context, log *zerolog.Logger, conn *pgxpool.Pool,
	migrationsPath, dirPath string, migrationNames []string,
) error {
	for i := range migrationNames {
		f, err := os.ReadFile(filepath.Join(migrationsPath, dirPath, migrationNames[i]+".up.sql"))
		if err != nil {
			return fmt.Errorf("error reading migration file %s/%s: %v", dirPath, migrationNames[i]+".up.sql", err)
		}
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}

		// nolint: errcheck // we don't care about the error here
		defer tx.Rollback(ctx)

		_, err = conn.Exec(ctx, string(f))
		log.Info().Msgf("running query: %s", string(f))
		if err != nil {
			log.Warn().Msgf("%s: rolling back migration %s: %v", dirPath, migrationNames[i], err)
			if e := migrateDown(ctx, conn, migrationsPath, dirPath, migrationNames[i]); e != nil {
				return e
			}
			return fmt.Errorf("error running migration %s: %v", migrationNames[i], err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("error committing transaction: %v", err)
		}
	}
	return nil
}

func migrateDown(ctx context.Context, conn *pgxpool.Pool,
	migrationsPath, dirPath, migrationName string,
) error {
	downFile, err := os.ReadFile(filepath.Join(migrationsPath, dirPath, migrationName+".down.sql"))
	if err != nil {
		return fmt.Errorf("error reading migration file %s/%s: %v", dirPath, migrationName+".down.sql", err)
	}
	_, err = conn.Exec(ctx, string(downFile))
	if err != nil {
		return fmt.Errorf("error rolling back migration %s: %v", migrationName, err)
	}
	return nil
}
