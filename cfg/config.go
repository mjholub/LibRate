package cfg

import (
	"os"
	"strings"

	"codeberg.org/mjh/LibRate/cfg/parser"
	"codeberg.org/mjh/LibRate/internal/logging"

	config "github.com/gookit/config/v2"
	"github.com/samber/lo"
)

// nolint:gochecknoglobals
var log = logging.Init()

func LoadConfig() Config {
	configRaw, err := parser.Parse("config.yml")
	if err != nil {
		log.Panic().Err(err).Msgf("error parsing config: %s", err.Error())
	}
	log.Info().Msgf("got config: %v", configRaw)

	keys := lo.Keys[string, interface{}](configRaw)
	dbKeys := lo.Map(keys, func(key string, index int) bool {
		return strings.Contains(key, "database")
	})

	dbConfig := &DBConfig{}
	dbConf := lo.ForEach(dbKeys, func(key string, index int) {
		config.MapStruct(dbKeys[index], dbConfig)
	})

	envChan := make(chan string, 1)
	defer close(envChan)
	getEnvOrDefault := func(envVar, defaultValue string) string {
		value := os.Getenv(envVar)
		if value == "" {
			os.Setenv(envVar, defaultValue)
			value = defaultValue
		}
		envChan <- value
		return value
	}

	if err != nil {
		log.Panic().Err(err).Msgf("error parsing config: %s", err.Error())
	}
	return Config{

		SiginingKey: os.Getenv("SIGNING_KEY"),
		DBPass:      os.Getenv("DB_PASS"),
	}
}

func LoadDgraph() *DgraphConfig {
	var (
		dghost        string
		dgport        string
		dghttp        string
		dgAlphaBadger string
		dgAlphaBRate  string
		dgAlphaTrace  string
		dgAlphaTLS    string
		dgAlphaSec    string
	)

	envChan := make(chan string, 1)
	defer close(envChan)

	getEnvOrDefault := func(envVar, defaultValue string) string {
		value := os.Getenv(envVar)
		if value == "" {
			os.Setenv(envVar, defaultValue)
			value = defaultValue
		}
		envChan <- value
		return value
	}
	go func() {
		dghost = getEnvOrDefault("DGRAPH_HOST", "0.0.0.0")
		dgport = getEnvOrDefault("DGRAPH_GRPC_PORT", "5080")
		dghttp = getEnvOrDefault("DGRAPH_HTTP_PORT", "6080")
		dgAlphaBadger = getEnvOrDefault("DGRAPH_ALPHA_BADGER", "compression=zstd;cache_size=1G;cache_ttl=1h;max_table_size=1G;level_size=128MB")
		dgAlphaBRate = getEnvOrDefault("DGRAPH_ALPHA_BLOCK_RATE", "20")
		dgAlphaTrace = getEnvOrDefault("DGRAPH_ALPHA_TRACE", "prometheus=localhost:9090")
		dgAlphaTLS = getEnvOrDefault("DGRAPH_ALPHA_TLS", "false")
		dgAlphaSec = getEnvOrDefault("DGRAPH_ALPHA_SECURITY", `whitelist=
		10.0.0.0/8,
		172.0.0.0/8,
		192.168.0.0/16,
		`+dghost+`
		`)
	}()

	// Retrieve the values from the channel
	dghost = <-envChan
	dgport = <-envChan
	dghttp = <-envChan
	dgAlphaBadger = <-envChan
	dgAlphaBRate = <-envChan
	dgAlphaTrace = <-envChan
	dgAlphaTLS = <-envChan
	dgAlphaSec = <-envChan

	return &DgraphConfig{
		Host:           dghost,
		GRPCPort:       dgport,
		HTTPPort:       dghttp,
		AlphaBadger:    dgAlphaBadger,
		AlphaBlockRate: dgAlphaBRate,
		AlphaTrace:     dgAlphaTrace,
		AlphaTLS:       dgAlphaTLS,
		AlphaSecurity:  dgAlphaSec,
	}
}
