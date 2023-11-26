// this file contaisn an exported global file with defalut config as a fallback if loading the
// proper config fails.

package cfg

import "github.com/gofrs/uuid/v5"

var (
	// nolint:gochecknoglobals
	// DefaultConfig is the default config for the application.
	// It is used as a fallback if loading the proper config fails.
	// Loading it should always be accompanied by a warning.
	DefaultConfig = Config{
		DBConfig: DBConfig{
			Engine:   "postgres",
			Host:     "localhost",
			Port:     uint16(5432),
			Database: "librate",
			User:     "postgres",
			Password: "postgres",
			SSL:      "unknown",
			PGConfig: "/usr/bin/pg_config",
			StartCmd: "sudo service postgresql start",
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
			Host:               "localhost",
			Port:               uint16(5432),
			Database:           "librate_test",
			User:               "postgres",
			Password:           "postgres",
			SSL:                "disable",
			PGConfig:           "/usr/bin/pg_config",
			StartCmd:           "skip",
			AutoMigrate:        true,
			ExitAfterMigration: false,
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Username: "",
			Password: "",
			Database: 5,
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
