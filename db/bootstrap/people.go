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
		CREATE ENUM IF NOT EXISTS people.role AS ('actor', 'director', 'producer', 'writer',
			'composer', 'artist', 'author', 'publisher', 'editor', 'photographer',
			'illustrator', 'narrator', 'performer', 'host', 'guest', 'other');`)
		if err != nil {
			return fmt.Errorf("failed to create people role enum: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.person (
			id INT PRIMARY KEY AUTO_INCREMENT,
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
			photos BIGINT[] REFERENCES cdn.images(id),
			added TIMESTAMP DEFAULT NOW() NOT NULL,
			modified TIMESTAMP DEFAULT NOW()
		);
	`)
		if err != nil {
			return fmt.Errorf("failed to create people table: %w", err)
		}
		_, err = db.Exec(`
			CREATE ENUM IF NOT EXISTS people.group_kind AS (
		'band', 'orchestra', 'choir', 'ensemble', 'troupe', 'collective', 'other');`)
		if err != nil {
			return fmt.Errorf("failed to create people group kind enum: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.group (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			locations UUID[] REFERENCES places.place(uuid),
			active BOOLEAN DEFAULT TRUE,
			formed DATE,
			disbanded DATE,
			website VARCHAR(255),
			photos BIGINT[] REFERENCES cdn.images(id),
			works UUID[] REFERENCES media.media(id),
			members INT[] REFERENCES people.person(id),
			genres SMALLINT[] REFERENCES media.genres(id),
			kind people.group_kind,
			added TIMESTAMP DEFAULT NOW() NOT NULL,
			modified TIMESTAMP DEFAULT NOW()
		);`)
		if err != nil {
			return fmt.Errorf("failed to create people group table: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.studio (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			city UUID REFERENCES places.city(uuid),	
			artists INT[] REFERENCES people.person(id),
			worts UUID[] REFERENCES media.media(id),
			is_film BOOLEAN DEFAULT FALSE,
			is_music BOOLEAN DEFAULT FALSE,
			is_tv BOOLEAN DEFAULT FALSE,
			is_publishing BOOLEAN DEFAULT FALSE,
			is_game BOOLEAN DEFAULT FALSE,
);`)
		if err != nil {
			return fmt.Errorf("failed to create people studio table: %w", err)
		}
		return nil
	}
}
