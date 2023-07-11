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
			CREATE SCHEMA IF NOT EXISTS people;
			SET search_path TO people, public;
		`)
		if err != nil {
			return fmt.Errorf("failed to create people schema: %w", err)
		}
		_, err = db.Exec(`
		CREATE TYPE people.role AS ENUM ('actor', 'director', 'producer', 'writer',
			'composer', 'artist', 'author', 'publisher', 'editor', 'photographer',
			'illustrator', 'narrator', 'performer', 'host', 'guest', 'other');`)
		if err != nil {
			return fmt.Errorf("failed to create people role enum: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.person (
			id SERIAL PRIMARY KEY,
			first_name VARCHAR(255) NOT NULL,
			other_names VARCHAR(255)[],
			last_name VARCHAR(255) NOT NULL,	
			nick_names VARCHAR(255)[],
			roles people.role[],
			birth DATE,
			death DATE,
			website VARCHAR(255),
			bio TEXT,
			added TIMESTAMP DEFAULT NOW() NOT NULL,
			modified TIMESTAMP DEFAULT NOW()
		);
	`)
		if err != nil {
			return fmt.Errorf("failed to create people table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.person_photos (
				person_id SERIAL REFERENCES people.person(id),
				image_id BIGINT REFERENCES cdn.images(id),
				PRIMARY KEY (person_id, image_id)
	);`)
		if err != nil {
			return fmt.Errorf("failed to create people photos table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.person_works (
				person_id SERIAL REFERENCES people.person(id),
				media_id UUID REFERENCES media.media(id),
				PRIMARY KEY (person_id, media_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people works table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TYPE people.group_kind AS ENUM (
		'band', 'orchestra', 'choir', 'ensemble', 'troupe', 'collective', 'other');`)
		if err != nil {
			return fmt.Errorf("failed to create people group kind enum: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.group (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			formed DATE,
			disbanded DATE,
			website VARCHAR(255),
			kind people.group_kind,
			added TIMESTAMP DEFAULT NOW() NOT NULL,
			modified TIMESTAMP DEFAULT NOW()
			wikipedia VARCHAR(255),
			bandcamp VARCHAR(255),
			soundcloud VARCHAR(255),
			bio TEXT
		);`)
		if err != nil {
			return fmt.Errorf("failed to create people group table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.group_locations (
				group_id SERIAL REFERENCES people.group(id),
				location_id UUID REFERENCES places.place(uuid),
				PRIMARY KEY (group_id, location_id)
	);`)
		if err != nil {
			return fmt.Errorf("failed to create people group locations table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.group_photos (
				group_id SERIAL REFERENCES people.group(id),
				image_id BIGINT REFERENCES cdn.images(id),
				PRIMARY KEY (group_id, image_id)
	);`)
		if err != nil {
			return fmt.Errorf("failed to create people group photos table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.group_members (
				group_id SERIAL REFERENCES people.group(id),
				person_id SERIAL REFERENCES people.person(id),
				PRIMARY KEY (group_id, person_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people group members table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.group_genres (
				group_id SERIAL REFERENCES people.group(id),
				primary_genre_id SMALLINT REFERENCES media.genres(id),
				secondary_genres SMALLINT[] REFERENCES media.genres(id),
				PRIMARY KEY (group_id, genre_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people group genres table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.group_works (
				group_id SERIAL REFERENCES people.group(id),
				media_id UUID REFERENCES media.media(id),
				PRIMARY KEY (group_id, media_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people group works table: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS people.studio (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			city UUID REFERENCES places.city(uuid),	
			is_film BOOLEAN DEFAULT FALSE,
			is_music BOOLEAN DEFAULT FALSE,
			is_tv BOOLEAN DEFAULT FALSE,
			is_publishing BOOLEAN DEFAULT FALSE,
			is_game BOOLEAN DEFAULT FALSE
);`)
		if err != nil {
			return fmt.Errorf("failed to create people studio table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.studio_artists (
				studio_id SERIAL REFERENCES people.studio(id),
				person_id SERIAL REFERENCES people.person(id),
				PRIMARY KEY (studio_id, person_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people studio artists table: %w", err)
		}
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people.studio_works (
				studio_id SERIAL REFERENCES people.studio(id),
				media_id UUID REFERENCES media.media(id),
				PRIMARY KEY (studio_id, media_id)
			);`)
		if err != nil {
			return fmt.Errorf("failed to create people studio works table: %w", err)
		}

		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS people.cast (
  id BIGSERIAL PRIMARY KEY,
  media_id uuid NOT NULL REFERENCES media.media(id),
  actors INTEGER[] NOT NULL,
  directors INTEGER[] NOT NULL
);
);`)
		if err != nil {
			return fmt.Errorf("failed to create people cast table: %w", err)
		}
		errChan := make(chan error)
		// don't defer closing the channel, let the GC handle it
		defer func() {
			err := createCastTrigger(db)
			if err != nil {
				errChan <- err
			}
		}()

		return nil
	}
}

func createCastTrigger(db *sqlx.DB) error {
	_, err := db.Exec(`
CREATE OR REPLACE FUNCTION check_actor_director_roles()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT (NEW.actors <@ (SELECT ARRAY(SELECT id FROM people.person WHERE roles = 'actor'))) THEN
        RAISE EXCEPTION 'Invalid actor(s) provided.';
    END IF;
    
    IF NOT (NEW.directors <@ (SELECT ARRAY(SELECT id FROM people.person WHERE roles = 'director'))) THEN
        RAISE EXCEPTION 'Invalid director(s) provided.';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_roles_before_insert_or_update
BEFORE INSERT OR UPDATE ON people.cast
FOR EACH ROW EXECUTE FUNCTION check_actor_director_roles();
`)
	if err != nil {
		return fmt.Errorf("failed to create cast trigger: %w", err)
	}

	return nil
}
