package cfg

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteConfig(t *testing.T) {
	type args struct {
		configPath string
		c          *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NoConfigFileSpecified", args{"", &Config{}}, true},
		{"ValidConfigFile", args{"config_tmp.yml", &Config{
			DBConfig: DBConfig{
				Engine:    "postgres",
				Host:      "localhost",
				Port:      uint16(5432),
				Database:  "write_test",
				TestDB:    "write_test_test",
				User:      "test_user",
				Password:  "test_password",
				SSL:       "unknown",
				PG_Config: "/usr/bin/pg_config",
			},
			Fiber: FiberConfig{
				Host: "localhost",
				Port: 3000,
			},
			SigningKey: "test_signing_key",
			Secret:     "test_secret",
			LibrateEnv: "test",
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeConfig(tt.args.configPath, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("WriteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCorrectWrite tests that the config file is written correctly
func TestCorrectWrite(t *testing.T) {
	suite := require.New(t)
	// test for existence of the old .env file, then back it up
	if _, err := os.Stat(".env"); err == nil {
		oldEnvFile, err := os.Open(".env")
		suite.NoError(err)
		// create a new file
		newEnvFile, err := os.Create(".env.bak")
		suite.NoError(err)
		// copy the old file to the new file
		_, err = io.Copy(newEnvFile, oldEnvFile)
		suite.NoError(err)
		// close the files
		err = oldEnvFile.Close()
		suite.NoError(err)
		err = newEnvFile.Close()
		suite.NoError(err)
		// restore the old file and delete the new file
		defer func() {
			oldEnvFile, err := os.Open(".env.bak")
			suite.NoError(err)
			newEnvFile, err := os.Create(".env")
			suite.NoError(err)
			_, err = io.Copy(newEnvFile, oldEnvFile)
			suite.NoError(err)
			err = oldEnvFile.Close()
			suite.NoError(err)
			err = newEnvFile.Close()
			suite.NoError(err)
			err = os.Remove(".env.bak")
			suite.NoError(err)
		}()
	}
	configPath := "config_tmp.yml"
	c := &Config{
		DBConfig: DBConfig{
			Engine:    "postgres",
			Host:      "localhost",
			Port:      uint16(5432),
			Database:  "write_test",
			TestDB:    "write_test_test",
			User:      "test_user",
			Password:  "test_password",
			SSL:       "unknown",
			PG_Config: "/usr/bin/pg_config",
		},
		Fiber: FiberConfig{
			Host: "localhost",
			Port: 3000,
		},
		SigningKey: "test_signing_key",
		Secret:     "test_secret",
		LibrateEnv: "test",
	}
	err := writeConfig(configPath, c)
	suite.NoError(err)
	yamlFile, err := os.ReadFile(configPath)
	suite.NoError(err)
	// NOTE: must use spaces (2 per tab) instead of tabs
	suite.Equal(`database:
    engine: postgres
    host: localhost
    port: 5432
    database: write_test
    test_db: write_test_test
    user: test_user
    password: test_password
    ssl: unknown
    pg_config: /usr/bin/pg_config
fiber:
    host: localhost
    port: 3000
signing_key: test_signing_key
secret: test_secret
librate_env: test
`, string(yamlFile))
	os.Remove(configPath)
}
