package cfg

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromFile(t *testing.T) {
	ExampleConfig := Config{
		DBConfig: DBConfig{
			Engine:             "postgres",
			Host:               "localhost",
			Port:               5432,
			Database:           "librate",
			User:               "postgres",
			Password:           "postgres",
			PGConfig:           "/usr/bin/pg_config",
			StartCmd:           "sudo systemctl start postgresql",
			AutoMigrate:        true,
			ExitAfterMigration: false,
		},
		Fiber: FiberConfig{
			Host:           "localhost",
			Port:           3000,
			Prefork:        true,
			ReduceMemUsage: false,
			StaticDir:      "./static",
			PowInterval:    300,
			PowDifficulty:  30000,
		},
		Secret:     "librate-secret-key",
		LibrateEnv: "production",
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Username: "",
			Password: "",
			CacheDB:  0,
		},
	}

	conf, err := LoadFromFile(filepath.Join("..", "example_config.yml"))
	assert.Equal(t, conf, &ExampleConfig)
	assert.Nil(t, err)

	// test whether the cli flags are parsed correctly
	configFile := flag.String("config", "../example_config.yml", "Path to config file")
	flag.Parse()

	conf, err = LoadFromFile(*configFile)
	assert.Equal(t, conf, &ExampleConfig)
	assert.Nil(t, err)
}

func TestParseRaw(t *testing.T) {
	configLocation := filepath.Join("..", "config_enc.yml")
	conf, err := parseRaw(configLocation)
	assert.Nil(t, err)
	assert.Equal(t, conf.DBConfig.Engine, "postgres")
}
