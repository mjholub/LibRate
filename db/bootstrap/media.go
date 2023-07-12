package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func MediaCore(ctx context.Context, connection *sqlx.DB) (err error) {
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
		CREATE TYPE media.kind AS ENUM ('album', 'track', 'film', 'tv_show', 'book', 'anime', 'manga', 'comic', 'game');
		CREATE TABLE IF NOT EXISTS media.media (
			id UUID PRIMARY KEY DEFAULT uuid_time_nextval(),
			title VARCHAR(255) NOT NULL,
			kind media.kind NOT NULL,
			created TIMESTAMP DEFAULT NOW() NOT NULL,
			creators serial4 NOT NULL references people.person(id);
		);`)
		if err != nil {
			return fmt.Errorf("failed to create media table: %w", err)
		}
		_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.genres (
			id SMALLSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			desc_short TEXT,
			desc_long TEXT,
			keywords VARCHAR(255)[],
			parent int2,
			children int2[]
			);`)
		if err != nil {
			return fmt.Errorf("failed to create media genres table: %w", err)
		}
		_, err = connection.ExecContext(ctx, `
			ALTER TABLE media.genres
				ADD CONSTRAINT genres_parent_fkey FOREIGN KEY (parent) REFERENCES media.genres(id),
		`)
		if err != nil {
			return fmt.Errorf("failed to add foreign key constraints to media genres table: %w", err)
		}

		return nil
	}
}

func Media(ctx context.Context, connection *sqlx.DB) (err error) {
	// TODO: use foreign keys to link media to artists and
	// create a graph-like structure
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		/*
		 * Languages
		 */
		_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.languages (
			id SMALLSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			iso_code VARCHAR(3) NOT NULL,
			iso_639_1 VARCHAR(2) NOT NULL
			)`)
		if err != nil {
			return fmt.Errorf("failed to create media languages table: %w", err)
		}
		/*
		 * Albums
		 */
		if err = bootstrapAlbums(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap albums tables: %w", err)
		}
		/*
		 * Tracks
		 */
		if err = bootstrapTracks(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap tracks tables: %w", err)
		}
		/*
		 * Films
		 */
		if err = bootstrapFilms(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap films tables: %w", err)
		}
		/*
		 * TV Shows
		 */
		if err = bootstrapTV(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap tv shows tables: %w", err)
		}

		/*
		* Books
		 */
		if err = bootstrapBooks(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap books tables: %w", err)
		}

		/*
		* Media images
		* NOTE: might probably be a better idea to create this before the media tables?
		 */
		if err = bootstrapMediaImages(ctx, connection); err != nil {
			return fmt.Errorf("failed to bootstrap media images tables: %w", err)
		}

		err = bootstrapKeywords(ctx, connection)
		if err != nil {
			return err
		}
		return nil
	}
}

// media.media_images should just be a junction table with additional column for the primary image
func bootstrapMediaImages(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.media_images (
			media_id UUID NOT NULL REFERENCES media.media(id),
			image_id BIGINT NOT NULL REFERENCES cdn.images(id),
			is_main BOOLEAN NOT NULL DEFAULT FALSE,
			PRIMARY KEY (media_id, image_id)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media images table: %w", err)
	}
	return nil
}

func bootstrapAlbums(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.albums (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			release_date TIMESTAMP NOT NULL,
			duration TIME NOT NULL
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media albums table: %w", err)
	}
	// junction table for keywords
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.album_keywords (
			album UUID NOT NULL REFERENCES media.albums(media_id),
			keyword_id SMALLINT NOT NULL REFERENCES media.keywords(id),
			PRIMARY KEY (album, keyword_id)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media album keywords table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.album_langs (
			album UUID NOT NULL REFERENCES media.albums(media_id),
			lang SMALLINT NOT NULL REFERENCES media.languages(id),
			PRIMARY KEY (album, lang)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media album languages table: %w", err)
	}

	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.album_artists (
			album UUID NOT NULL REFERENCES media.albums(media_id),
			artist SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (album, artist)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media album artists table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.album_genres (
			album UUID NOT NULL REFERENCES media.albums(media_id),
			genre SMALLSERIAL NOT NULL REFERENCES media.genres(id),
			PRIMARY KEY (album, genre)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media album genres table: %w", err)
	}

	return nil
}

func bootstrapTracks(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tracks (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			album UUID NOT NULL REFERENCES media.albums(media_id),
			duration TIME NOT NULL,
			lyrics TEXT
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tracks table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.album_tracks (
			album UUID NOT NULL REFERENCES media.albums(media_id),
			track UUID NOT NULL REFERENCES media.tracks(media_id),
			track_number SMALLINT NOT NULL,
			PRIMARY KEY (album, track)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media album tracks table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.track_artists (
			track UUID NOT NULL REFERENCES media.tracks(media_id),
			artist SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (track, artist)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media track artists table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.track_langs (
			track UUID NOT NULL REFERENCES media.tracks(media_id),
			lang SMALLINT NOT NULL REFERENCES media.languages(id),
			PRIMARY KEY (track, lang)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media track languages table: %w", err)
	}
	return nil
}

func bootstrapFilms(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.films (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL
);`)
	if err != nil {
		return fmt.Errorf("failed to create media films table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.film_cast (
			film UUID NOT NULL REFERENCES media.films(media_id),
			person SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (film, person)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media film cast table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.film_countries (
			film UUID NOT NULL REFERENCES media.films(media_id),
			country SMALLINT NOT NULL REFERENCES places.country(id),
			PRIMARY KEY (film, country)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media film countries table: %w", err)
	}
	return nil
}

func bootstrapTV(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_shows (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv shows table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_show_countries (
			tv_show UUID NOT NULL REFERENCES media.tv_shows(media_id),
			country SMALLINT NOT NULL REFERENCES places.country(id),
			PRIMARY KEY (tv_show, country)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv show countries table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_show_cast (
			tv_show UUID NOT NULL REFERENCES media.tv_shows(media_id),
			person SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (tv_show, person)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv show cast table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_show_langs (
			tv_show UUID NOT NULL REFERENCES media.tv_shows(media_id),
			lang SMALLINT NOT NULL REFERENCES media.languages(id),
			PRIMARY KEY (tv_show, lang)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv show languages table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_show_seasons (
			media_id UUID REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			tv_show UUID NOT NULL REFERENCES media.tv_shows(media_id),
			season SMALLINT NOT NULL,
			PRIMARY KEY (tv_show, season)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv show seasons table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.tv_show_episodes (
			media_id UUID REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			tv_show UUID NOT NULL,
			season SMALLINT NOT NULL,
			episode SMALLINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			PRIMARY KEY (tv_show, season, episode),
			FOREIGN KEY (tv_show, season) REFERENCES media.tv_show_seasons(tv_show, season)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media tv show episodes table: %w", err)
	}
	return nil
}

func bootstrapBooks(ctx context.Context, connection *sqlx.DB) (err error) {
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.books (
			media_id UUID PRIMARY KEY REFERENCES media.media(id) DEFAULT uuid_generate_v4(),
			edition VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			publisher SERIAL REFERENCES people.group(id),
			publication_date TIMESTAMP,
			keywords TEXT[],
			pages SMALLINT,
			ISBN VARCHAR(255),
			ASIN VARCHAR(255),
			cover TEXT,
			summary TEXT
		);
		`)
	if err != nil {
		return fmt.Errorf("failed to create media tables: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.book_authors (
			book UUID NOT NULL REFERENCES media.books(media_id),
			person SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (book, person)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media book authors table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.book_languages (
			book UUID NOT NULL REFERENCES media.books(media_id),
			lang SMALLINT NOT NULL REFERENCES media.languages(id),
			PRIMARY KEY (book, lang)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media book languages table: %w", err)
	}
	_, err = connection.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media.book_genres (
			book UUID NOT NULL REFERENCES media.books(media_id),
			genre SMALLSERIAL NOT NULL REFERENCES media.genres(id),
			PRIMARY KEY (book, genre)
		);`)
	if err != nil {
		return fmt.Errorf("failed to create media book genres table: %w", err)
	}

	return nil
}

func bootstrapKeywords(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS media.keywords (
  id serial PRIMARY KEY,
  keyword varchar NOT NULL UNIQUE,
  media_id uuid references media.media(id),
  total_stars integer not null default 0,
  vote_count integer not null default 0
  );`)
	if err != nil {
		return fmt.Errorf("failed to create keywords table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS keywords_media_id_idx ON media.keywords(media_id);`)
	if err != nil {
		return fmt.Errorf("failed to create keywords index: %w", err)
	}
	return nil
}
