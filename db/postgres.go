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
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
	`)
	if err != nil {
		return fmt.Errorf("failed to create members table: %w", err)
	}
	// TODO: use foreign keys to link media to artists and
	// create a graph-like structure
	_, err = db.Exec(`
		CREATE SCHEMA IF NOT EXISTS media;`,
	)
	if err != nil {
		return fmt.Errorf("failed to create media schema: %w", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS media.albums (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			arists VARCHAR(255) NOT NULL,
			release_date TIMESTAMP NOT NULL,
			genres VARCHAR(255) NOT NULL,
			keywords VARCHAR(255) NOT NULL,
			duration INTERVAL NOT NULL,
			tracks VARCHAR(255) NOT NULL,
			languages VARCHAR(255) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS media.tracks (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			artists ARRAY NOT NULL,
			album VARCHAR(255) NOT NULL,
			duration INTERVAL NOT NULL,
			languages ARRAY NOT NULL,
			lyrics TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS media.films (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			cast ARRAY NOT NULL, 
);
		CREATE TABLE IF NOT EXISTS media.tv_shows (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			cast ARRAY NOT NULL,
			seasons ARRAY NOT NULL,	
		);
		CREATE TABLE IF NOT EXISTS media.books (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			authors ARRAY NOT NULL,
			publisher VARCHAR(255) NOT NULL,
			publication_date TIMESTAMP NOT NULL,
			genres ARRAY NOT NULL,
			keywords ARRAY NOT NULL,
			languages ARRAY NOT NULL,
			pages INTEGER NOT NULL,
			ISBN VARCHAR(255) NOT NULL,
			ASIN VARCHAR(255) NOT NULL,
			cover TEXT NOT NULL,
			summary TEXT NOT NULL
		);
		`)
	if err != nil {
		return fmt.Errorf("failed to create media tables: %w", err)
	}
	return nil
}
