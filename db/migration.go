package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
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
func Migrate(log *zerolog.Logger, conf *cfg.Config, paths ...string) error {
	dsn := CreateDsn(&conf.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer conn.Close(ctx)

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
			for i := range files {
				f, err := os.ReadFile(filepath.Join(conf.MigrationsPath, dir.Name(), files[i].Name()))
				if err != nil {
					return fmt.Errorf("error reading migration file: %v", err)
				}
				_, err = conn.Exec(ctx, string(f))
				log.Info().Msgf("running query: %s", string(f))
				if err != nil {
					return fmt.Errorf("error running migration: %v", err)
				}
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
			for i := range files {
				f, err := os.ReadFile(filepath.Join(conf.MigrationsPath, paths[i], files[i].Name()))
				if err != nil {
					return fmt.Errorf("error reading migration file: %v", err)
				}
				_, err = conn.Exec(ctx, string(f))
				log.Info().Msgf("running query: %s", string(f))
				if err != nil {
					return fmt.Errorf("error running migration: %v", err)
				}
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
