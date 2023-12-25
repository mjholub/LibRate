package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
)

// Migrate runs the migrations located in the migrations folder
// paths is a variadic argument, meaning that if no specific path(s)
// are provided, the program will try to run all of them
// Otherwise the arguments to path should only include the containing
// directory name for each migration, e.g.
// "000001-fix-missing-timestamps"
func Migrate(conf *cfg.Config, paths ...string) error {
	dsn := CreateDsn(&conf.DBConfig)
	if paths == nil {
		m, err := migrate.New(
			"file:///migrations",
			dsn,
		)
		if err != nil {
			return fmt.Errorf("error preparing migrations: %v")
		}
		err = m.Up()
		if err != nil {
			return fmt.Errorf("error running migrations: %v", err)
		}
		return nil
	} else {
		for i := range paths {
			count, err := countFiles(paths[i])
			if err != nil {
				return err
			}
			m, err := migrate.New(
				fmt.Sprintf("file:///migrations/%s", paths[i]),
				dsn,
			)
			if err != nil {
				return fmt.Errorf("error running migration for directory: %s: %v", paths[i], err)
			}
			m.Steps(int(count))
		}
		return nil
	}
}

func countFiles(path string) (count uint8, err error) {
	dir, err := os.ReadDir(filepath.Join(".", "migrations", path))
	if err != nil {
		return 0, fmt.Errorf("error reading filesystem: %v", err)
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
