package bootstrap

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
)

func People(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people", err)
		}
		defer tx.Rollback()

		_, err = db.Exec(`
			CREATE SCHEMA IF NOT EXISTS people;
			SET search_path TO people, public;
		`)
		if err != nil {
			return fmt.Errorf("failed to create people schema: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}
}

// Roles creates the people.roles enum type and the people.person table
// Supported roles are:
// actor, director, producer, writer, composer, artist, author, publisher, editor, photographer, illustrator, narrator, performer, host, guest, other
func Roles(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.roles", err)
		}
		defer tx.Rollback()
		var mu sync.Mutex
		mu.Lock()
		err = roleTypes(ctx, db)
		if err != nil {
			return err
		}
		mu.Unlock()
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

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}
}

func roleTypes(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.role_types", err)
		}
		defer tx.Rollback()
		peopleRoles := []string{
			"actor", "director", "producer", "writer",
			"composer", "artist", "author", "publisher", "editor", "photographer",
			"illustrator", "narrator", "performer", "host", "guest", "other",
		}
		err = createEnumType(ctx, db, "role", "people", peopleRoles...)
		if err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}
}

// MediaCreators creates the media.media_creators table
func MediaCreators(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("media.media_creators", err)
		}
		defer tx.Rollback()
		// junction table for additional media creators
		_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.media_creators (
			media_id UUID NOT NULL references media.media(id),
			creator_id INTEGER NOT NULL references people.person(id),
			PRIMARY KEY (media_id, creator_id)
		);`)
		if err != nil {
			return fmt.Errorf("failed to create media creators table: %w", err)
		}

		_, err = db.ExecContext(ctx, `CREATE SEQUENCE media.media_creators_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;`)
		if err != nil {
			return fmt.Errorf("failed to create media creators sequence: %w", err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		err = mediaFkey(ctx, db)
		if err != nil {
			return err
		}

		return nil
	}
}

func mediaFkey(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.ExecContext(ctx, `
		ALTER TABLE media.media
			ADD COLUMN IF NOT EXISTS creator int4
			DEFAULT nextval('media.media_creators_seq'::regclass),
			ADD CONSTRAINT media_creator_fkey FOREIGN KEY (creator) REFERENCES people.person(id);
		`)
		if err != nil {
			return fmt.Errorf("failed to add foreign key constraints to media table: %w", err)
		}
		return nil
	}
}

// PeopleMeta creates the tables that store the artists' photos and works
func PeopleMeta(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.person", err)
		}
		defer tx.Rollback()

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

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	}
}

// CreatorGroups creates the people.group table and its associated tables
func CreatorGroups(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.group", err)
		}

		defer tx.Rollback()

		groupTypes := []string{"band", "orchestra", "choir", "ensemble", "troupe", "collective", "other"}
		err = createEnumType(ctx, db, "group_kind", "people", groupTypes...)
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
			modified TIMESTAMP DEFAULT NOW(),
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
				PRIMARY KEY (group_id, primary_genre_id)
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

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	}
}

// Studios creates the people.studio table and its associated tables
// By studio we mean an entity that produces media, such as a film studio, a record label, a publishing house, etc.

func Studio(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.studio", err)
		}
		defer tx.Rollback()

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

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}
}

// Cast creates the people.cast table and its associated tables
func Cast(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return txErr("people.cast", err)
		}
		defer tx.Rollback()

		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS people.cast (
  id BIGSERIAL PRIMARY KEY,
  media_id uuid NOT NULL REFERENCES media.media(id),
  actors INTEGER[] NOT NULL,
  directors INTEGER[] NOT NULL
);`)
		if err != nil {
			return fmt.Errorf("failed to create people cast table: %w", err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
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

func txErr(table string, err error) error {
	return fmt.Errorf("failed to create %s schema/table: %w", table, err)
}
