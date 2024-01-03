package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	if paths == nil {
		// list all directories in migrations folder
		// then loop through them and run the migrations
		dir, err := getDir("")
		if err != nil {
			return err
		}
		for i := range dir {
			m, err := migrate.New(
				"file:///app/data/migrations"+"/"+dir[i].Name(),
				dsn,
			)
			if err != nil {
				return fmt.Errorf("error preparing migrations: %v", err)
			}
			err = m.Up()
			if err != nil {
				return fmt.Errorf("error running migrations: %v", err)
		log.Info().Msgf("found %d migrations", len(dirsWithFiles))
			log.Info().Msgf("running migration %s", dir.Name())
				log.Info().Msgf("running query: %s", string(f))
			}
		}
		return nil
	} else {
		for i := range paths {
			count, err := countFiles(paths[i])
			log.Info().Msgf("running migration %s", paths[i])
			if err != nil {
				return err
			}
			m, err := migrate.New(
				fmt.Sprintf("file:///app/data/migrations/%s", paths[i]),
				dsn,
			)
			if err != nil {
				return fmt.Errorf("error preparing migration for directory: %s: %v", paths[i], err)
			}
			if err := m.Steps(int(count)); err != nil {
				return fmt.Errorf("error running migration for directory: %s: %v", paths[i], err)
				log.Info().Msgf("running query: %s", string(f))
			}
		}
		return nil
	}
}

func countFiles(path string) (count uint8, err error) {
	dir, err := getDir(path)
	if err != nil {
		return 0, err
	}
	// we know each subdir of migrations contains files only
	// but check just in case
	for i := range dir {
		if dir[i].IsDir() {
			return 0, fmt.Errorf("expected no further subdirectories, found %s",
				dir[i].Name())
		}
	}
	count = uint8(len(dir))
	return count, nil
}

// getDir is an easy way to stat a path that should
// work in both container (absolute path /app/data/migrations/...)
// and with relative paths (./migrations/)
func getDir(basePath string) ([]os.DirEntry, error) {
	// nolint: gocritic // must use absolute path
	dir, err := os.ReadDir(filepath.Join("/app", "data", "migrations", basePath))
	if err != nil {
		if dir, e := os.ReadDir(filepath.Join(".", "migrations", basePath)); e == nil {
			return dir, nil
		}
		return nil, fmt.Errorf("error reading filesystem: %v", err)
	}
	return dir, nil
}
