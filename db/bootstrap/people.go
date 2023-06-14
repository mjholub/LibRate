package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func People(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(`
		CREATE ENUM people.role AS ('actor', 'director', 'producer', 'writer',
			'composer', 'artist', 'author', 'publisher', 'editor', 'photographer',
			'illustrator', 'narrator', 'performer', 'host', 'guest', 'other');
		CREATE TABLE IF NOT EXISTS people.person (
			id UNSIGNED INT PRIMARY KEY AUTO_INCREMENT,
			first_name VARCHAR(255) NOT NULL,
			other_names VARCHAR(255)[],
			last_name VARCHAR(255) NOT NULL,	
			nick_name VARCHAR(255),
			roles people.role[],
			works UUID[] REFERENCES media.media(id),
			birth DATE,
			death DATE,
			website VARCHAR(255),
			bio TEXT,
			photos UNSIGNED INT[] REFERENCES cdn.images(id),
			email VARCHAR(255) NOT NULL,
			passhash VARCHAR(255) NOT NULL,
			reg_timestamp TIMESTAMP DEFAULT NOW() NOT NULL 
		);
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
	`)
		if err != nil {
			return fmt.Errorf("failed to create people table: %w", err)
		}
		return nil
	}
}
