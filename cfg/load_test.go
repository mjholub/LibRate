package cfg

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Name   string
	Env    *map[string]string
	Inputs interface{}
}

func TestLoadFromFile(t *testing.T) {
	ExampleConfig := Config{
		DBConfig: DBConfig{
			Engine:             "postgres",
			Host:               "localhost",
			Port:               5432,
			Database:           "librate",
			User:               "postgres",
			Password:           "postgres",
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
	testCases := []testCase{
		{
			Name:   "encrypted config",
			Inputs: filepath.Join("..", "config_enc.yml"),
		}, {
			Name:   "plain text",
			Env:    &map[string]string{"USE_SOPS": "false"},
			Inputs: filepath.Join("..", "example_config.yml"),
		},
	}
	for _, tc := range testCases {
		if tc.Env != nil {
			os.Setenv(lo.Keys(*tc.Env)[0], lo.Values(*tc.Env)[0])
		}

		conf, err := parseRaw(tc.Inputs.(string))
		assert.Nil(t, err)
		assert.Equal(t, conf.DBConfig.Engine, "postgres")
	}
}
