package cfg

import (
	"os"
)

type Config struct {
	ArangoDB ArangoDBConfig
	Fiber    FiberConfig
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
	}
}
