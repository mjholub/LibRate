package cfg

import (
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
				Engine: "postgres",
				Host:   "localhost",
			},
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
	configPath := "config_tmp.yml"
	c := &Config{
		DBConfig: DBConfig{
			Engine: "postgres",
			Host:   "localhost",
		},
	}
	err := writeConfig(configPath, c)
	suite.NoError(err)
	yamlFile, err := os.ReadFile(configPath)
	suite.NoError(err)
	suite.Equal(`database:
	engine: postgres
	host: localhost
`, string(yamlFile))
}
