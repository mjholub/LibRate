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
	Host     string
	GRPCPort string
	HTTPPort string
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
	dghost := os.Getenv("DGRAPH_HOST")
	if dghost == "" {
		dghost = "0.0.0.0"
	}
	dgport := os.Getenv("DGRAPH_GRPC_PORT")
	if dgport == "" {
		dgport = "5080"
	}
	dghttp := os.Getenv("DGRAPH_HTTP_PORT")
	if dghttp == "" {
		dghttp = "6080"
	}
	return &DgraphConfig{
		Host:     dghost,
		GRPCPort: dgport,
		HTTPPort: dghttp,
	}
}
