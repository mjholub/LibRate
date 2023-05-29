package cfg

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ReadPostgres() (*PostgresConfig, error) {
	var (
		host     string
		port     string
		user     string
		password string
		database string
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
		host = getEnvOrDefault("POSTGRES_HOST", "localhost")
		port = getEnvOrDefault("POSTGRES_PORT", "5432")
		user = getEnvOrDefault("POSTGRES_USER", "postgres")
		password = getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
		database = getEnvOrDefault("POSTGRES_DATABASE", "librate")
	}()

	return &PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}, nil
}

func createIfNotExists(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS members (
			id SERIAL PRIMARY KEY,
			uuid UUID NOT NULL DEFAULT uuid_generate_v4(),
			nickname VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			reg_timestamp TIMESTAMP DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}
	return nil
}
