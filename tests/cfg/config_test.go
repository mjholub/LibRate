package cfg_test

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/mjh/LibRate/cfg"
)

type test struct {
	name string
	env  map[string]string
	want cfg.Config
}

func TestLoadConfig(t *testing.T) {
	tests := []test{
		{
			name: "default",
			env:  map[string]string{},
			want: cfg.Config{
				DBConfig: cfg.DBConfig{
					Host:     "localhost",
					Database: "librate",
					Port:     5432,
					User:     "postgres",
					Password: "librate",
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, envVar := range tc.env {
				splitEnvVar := strings.SplitN(envVar, "=", 2)
				if len(splitEnvVar) != 2 {
					t.Errorf("Environment variable is invalid: %s", envVar)
				}
				os.Setenv(splitEnvVar[0], splitEnvVar[1])
			}
		})
	}
}
