package cfg

import (
	"os"
)

type Config struct {
	ArangoDB    ArangoDBConfig
	Dgraph      DgraphConfig
	Fiber       FiberConfig
	SiginingKey string
}

type ArangoDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type FiberConfig struct {
	Host string
	Port string
}

type DgraphConfig struct {
	Host           string
	GRPCPort       string
	HTTPPort       string
	AlphaBadger    string
	AlphaBlockRate string
	AlphaTrace     string
	AlphaTLS       string
	AlphaSecurity  string
}

func LoadConfig() Config {
	return Config{
		ArangoDB: ArangoDBConfig{
			Host:     os.Getenv("ARANGODB_HOST"),
			Port:     os.Getenv("ARANGODB_PORT"),
			User:     os.Getenv("ARANGODB_USER"),
			Password: os.Getenv("ARANGODB_PASSWORD"),
			Database: os.Getenv("ARANGODB_DATABASE"),
		},
		Fiber: FiberConfig{
			Host: os.Getenv("FIBER_HOST"),
			Port: os.Getenv("FIBER_PORT"),
		},
		SiginingKey: os.Getenv("SIGNING_KEY"),
		Dgraph: DgraphConfig{
			Host:     os.Getenv("DGRAPH_HOST"),
			GRPCPort: os.Getenv("DGRAPH_GRPC_PORT"),
			HTTPPort: os.Getenv("DGRAPH_HTTP_PORT"),
		},
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
