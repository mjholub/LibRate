package cfg_test

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/samber/lo"
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
				ArangoDB: cfg.ArangoDBConfig{
					Host:     "http://localhost:8529",
					Database: "librate",
					Port:     "8529",
					User:     "librate",
					Password: "librate",
				},
				Dgraph: cfg.DgraphConfig{
					Host:           "0.0.0.0",
					GRPCPort:       "5080",
					HTTPPort:       "6080",
					AlphaBadger:    "compression=zstd;cache_size=1G;cache_ttl=1h;max_table_size=1G;level_size=128MB",
					AlphaBlockRate: "20",
					AlphaTrace:     "prometheus=localhost:9090",
					AlphaTLS:       "false",
					AlphaSecurity: `whitelist=
								10.0.0.0/8,
								172.0.0.0/8,
								192.168.0.0/16,
								0.0.0.0,
				`,
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
