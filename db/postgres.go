// TODO: verify if this file is needed
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"codeberg.org/mjh/LibRate/cfg"
)

func CreateDsn(dsn *cfg.DBConfig) string {
	switch dsn.SSL {
	case "require", "verify-ca", "verify-full", "disable":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, dsn.SSL)
		fmt.Println(data)
		return data
	case "prefer":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database, "require")
		fmt.Println(data)
		return data
	case "unknown":
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		fmt.Println(data)
		return data
	default:
		data := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
			dsn.Engine, dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.Database)
		fmt.Println(data)
		return data
	}
}

func Connect(conf *cfg.Config) (*sqlx.DB, error) {
	data := CreateDsn(&conf.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, conf.Engine, data)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitDB() error {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	db, err := Connect(&conf)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS public.members (
			id SERIAL PRIMARY KEY,
			uuid UUID NOT NULL,
			nick VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			passhash VARCHAR(255) NOT NULL,
			reg_timestamp TIMESTAMP DEFAULT NOW() NOT NULL 
		);
	`)
	if err != nil {
		return err
	}

	return nil
}
