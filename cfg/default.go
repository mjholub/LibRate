// this file contains an exported global file with default config as a fallback if loading the
// proper config fails.

package cfg

import (
	"codeberg.org/mjh/LibRate/internal/logging"
	"github.com/gofrs/uuid/v5"
)

var (
	// nolint:gochecknoglobals
	// DefaultConfig is the default config for the application.
	// It is used as a fallback if loading the proper config fails.
	// Loading it should always be accompanied by a warning.
	DefaultConfig = Config{
		DBConfig: DBConfig{
			Engine:         "postgres",
			Host:           "localhost",
			Port:           uint16(5432),
			Database:       "librate",
			User:           "postgres",
			Password:       "postgres",
			SSL:            "unknown",
			MigrationsPath: "/app/data/migrations",
		},
		Fiber: FiberConfig{
			Host:    "localhost",
			Port:    3000,
			Prefork: false,
		},
		Secret:     uuid.Must(uuid.NewV7()).String(),
		LibrateEnv: "production",
	}
	// nolint:gochecknoglobals
	// TestConfig is a convenience config for testing, so that the test functions are terser, avoiding unnecessary repetition.
	TestConfig = Config{
		DBConfig: DBConfig{
			Engine:             "postgres",
			Host:               "0.0.0.0",
			Port:               uint16(5432),
			Database:           "librate_test",
			User:               "postgres",
			Password:           "postgres",
			SSL:                "disable",
			ExitAfterMigration: false,
			MigrationsPath:     "./migrations",
		},
		Logging: logging.Config{
			Level:  "debug",
			Target: "stdout",
			Format: "json",
			Caller: true,
			Timestamp: logging.TimestampConfig{
				Enabled: false,
				Format:  "2006-01-0215:04:05.000Z07:00",
			},
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Username: "",
			Password: "",
			CacheDB:  8,
			PowDB:    9,
			CsrfDB:   11,
		},
		CouchDB: Search{
			Host:     "0.0.0.0",
			Port:     5984,
			User:     "librate",
			Password: "librate",
		},
		Fiber: FiberConfig{
			Host:           "0.0.0.0",
			Port:           3001,
			Prefork:        false,
			ReduceMemUsage: false,
			StaticDir:      "./static",
			PowInterval:    1,
			PowDifficulty:  1,
		},
		Secret:     "secret",
		LibrateEnv: "test",
	}
)
