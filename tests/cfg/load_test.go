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
			Engine:    "postgres",
			Host:      "localhost",
			Port:      uint16(5432),
			Database:  "librate",
			TestDB:    "librate_test",
			User:      "postgres",
			Password:  "postgres",
			SSL:       "unknown",
			PG_Config: "/usr/bin/pg_config",
		},
		Fiber: cfg.FiberConfig{
			Host: "localhost",
			Port: "3000",
		},
		SigningKey: "eyJhbGciOiJIUzUxMiJ9.eyJSb2xlIjoiQWRtaW4iLCJJc3N1ZXIiOiJMaWJSYXRlIGRldmVsb3BlcnMiLCJVc2VybmFtZSI6InVzZXIiLCJleHAiOjE3MjUzMDc2OTMsImlhdCI6MTY5MzY4NTI5M30.GXq5OBlI4xvIlY5EnotksThjbsgDclm8ZjPl2Ans54XkeUnDE35RA9OD477EfkHrjVch8QihNFJrjpLgoeFQhA",
		Secret:     "librate-secret-key",
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
			suite.Nil(result)
		})
	}

	t.Run("NonExampleConfigFileExists", func(t *testing.T) {
		result, err := cfg.LoadConfig().Get()
		suite.NoError(err)
		suite.NotEqual(result, defaultConfig)
	})
}
