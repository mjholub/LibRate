package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Media(ctx context.Context, connection *sqlx.DB) (err error) {
	// TODO: use foreign keys to link media to artists and
	// create a graph-like structure
	defer connection.Close()
	_, err = connection.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS media;`,
	)
	if err != nil {
		return fmt.Errorf("failed to create media schema: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.albums (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			arists INTEGER[] NOT NULL,
			release_date TIMESTAMP NOT NULL,
			genres VARCHAR(255),
			keywords VARCHAR(255),
			duration INTERVAL NOT NULL,
			tracks UUID[] NOT NULL,
			languages SMALLINT[],
		);
		CREATE TABLE IF NOT EXISTS media.tracks (
			id SERIAL PRIMARY KEY,
			uuid UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			artists INTEGER[] NOT NULL,
			album INTEGER NOT NULL,
			duration INTERVAL NOT NULL,
			languages SMALLINT[],
			lyrics TEXT,
		);
		CREATE TABLE IF NOT EXISTS media.films (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			cast INTEGER[],  -- person.id
);
		CREATE TABLE IF NOT EXISTS media.tv_shows (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			cast INTEGER[],  -- person.id
			seasons integer[],	
		);
		CREATE TABLE IF NOT EXISTS media.books (
			id SERIAL PRIMARY KEY,
			edition VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			authors ARRAY NOT NULL,
			publisher VARCHAR(255),
			publication_date TIMESTAMP,
			genres TEXT[],
			keywords TEXT[],
			languages TEXT[],
			pages INTEGER,
			ISBN VARCHAR(255),
			ASIN VARCHAR(255),
			cover TEXT,
			summary TEXT,
		);
		`)
	if err != nil {
		return fmt.Errorf("failed to create media tables: %w", err)
	}
}
