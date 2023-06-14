package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Media(ctx context.Context, connection *sqlx.DB) (err error) {
	// TODO: use foreign keys to link media to artists and
	// create a graph-like structure
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err = connection.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS media;`,
		)
		if err != nil {
			return fmt.Errorf("failed to create media schema: %w", err)
		}
		_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.genres (
			id  SMALLINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			desc_short VARCHAR(255),
			desc_long TEXT,
			keywords VARCHAR(255)[],
			parent UUID REFERENCES media.genres(media_id),
			children UUID REFERENCES media.genres(media_id),
		);
		CREATE TABLE IF NOT EXISTS media.albums (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			arists INTEGER[] NOT NULL REFERENCES media.person(id),
			release_date TIMESTAMP NOT NULL,
			genres VARCHAR(255)[] NREFERENCES media.genres(media_id),
			keywords VARCHAR(255)[],
			duration INTERVAL NOT NULL,
			tracks UUID[] NOT NULL REFERENCES media.tracks(media_id),
			languages SMALLINT[] REFERENCES media.languages(id),
		);
		CREATE TABLE IF NOT EXISTS media.tracks (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			artists INTEGER[] NOT NULL REFERENCES media.person(id),
			album UUID NOT NULL REFERENCES media.albums(media_id),
			duration INTERVAL NOT NULL,
			languages SMALLINT[] REFERENCES media.languages(id),
			lyrics TEXT,
		);
		CREATE TABLE IF NOT EXISTS media.films (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			cast INTEGER[] REFERENCES media.person(id),
);
		CREATE TABLE IF NOT EXISTS media.tv_shows (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			cast INTEGER[] REFERENCES media.person(id),
			seasons UUID[] REFERENCES media.seasons(media_id),	
		);
		CREATE TABLE IF NOT EXISTS media.books (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			edition VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			authors INTEGER[] NOT NULL REFERENCES media.person(id),
			publisher VARCHAR(255) REFERENCES media.group(id),
			publication_date TIMESTAMP,
			genres UUID[] REFERENCES media.genres(media_id),
			keywords TEXT[],
			languages INTEGER[] REFERENCES media.languages(id),
			pages SMALLINT,
			ISBN VARCHAR(255),
			ASIN VARCHAR(255),
			cover TEXT,
			summary TEXT,
		);
		CREATE ENUM media.kind AS ('album', 'track', 'film', 'tv_show', 'book');
		CREATE TABLE IF NOT EXISTS media.media (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			kind media.kind NOT NULL,
			created TIMESTAMP DEFAULT NOW() NOT NULL,
			genres UUID[] REFERENCES media.genres(media_id),
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create media tables: %w", err)
		}
		return nil
	}
}
