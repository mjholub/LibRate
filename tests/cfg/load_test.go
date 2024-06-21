package cfg_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
)

type testCase struct {
	name string
	want interface{}
}

func TestLoadConfig(t *testing.T) {
	// defaultConfig denotes the default configuration for the application
	// as defined in the example_config.yml file.
	defaultConfig := cfg.Config{
		DBConfig: cfg.DBConfig{
			Host:     "localhost",
			Port:     uint16(5432),
			Database: "librate",
			User:     "postgres",
			Password: "postgres",
			SSL:      "unknown",
		},
		Fiber: cfg.FiberConfig{
			Host: "localhost",
			Port: 3000,
		},
		Secret: "librate-secret-key",
	}

	suite := require.New(t)
	if _, err := os.Stat("config.yml"); err == nil {
		t.Run("ExampleConfigFileExists", func(t *testing.T) {
			result, err := cfg.LoadConfig().Get()
			suite.NoError(err)
			suite.NotNil(result)
		})
	} else {
		t.Run("NoConfigFile", func(t *testing.T) {
			result, err := cfg.LoadConfig().Get()
			suite.Error(err)
			suite.Nil(&result, t)
		})
	}

	t.Run("NonExampleConfigFileExists", func(t *testing.T) {
		result, err := cfg.LoadConfig().Get()
		suite.NoError(err)
		suite.NotEqual(result, defaultConfig, t)
	})
}
