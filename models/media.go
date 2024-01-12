package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	MediaService interface {
		IsMedia() bool // dummy placeholder so that we can have somewhat idiomatic parametric polymorphism
	}

	MediaStorer[T any] interface {
		Get(ctx context.Context, key string) (T, error)
		GetAll() ([]T, error)
		Add(ctx context.Context, db *sqlx.DB, props Media) (uuid.UUID, error)
		Update(ctx context.Context, key string, value T) error
		Delete(ctx context.Context, key string) error
	}

	// nolint:musttag // false positive, can only annotate fields, not types
	Media struct {
		ID       uuid.UUID     `json:"id" db:"id,pk,unique"`
		Title    string        `json:"title" db:"title"`
		Kind     string        `json:"kind" db:"kind"`
		Created  time.Time     `json:"keywords,omitempty" db:"created"`
		Creator  sql.NullInt32 `json:"creator,omitempty" db:"creator"`
		Creators []Person      `json:"creators,omitempty"` // no db tag, we're using a junction table
		Added    time.Time     `json:"added,omitempty" db:"added"`
		Modified sql.NullTime  `json:"modified,omitempty" db:"modified"`
	}

	MediaDetails struct {
		Kind    string      `json:"kind" db:"kind"`
		Details interface{} `json:"details" db:"details"`
	}

	MediaObject interface {
		Book | Album | Track | TVShow | Season | Episode
	}

	// Genre does not hage a UUID due to parent-child relationships
	Genre struct {
		ID          int64              `json:"id" db:"id,pk,autoinc"`
		Kinds       pq.StringArray     `json:"kind" db:"kind" enum:"music,film,tv,book,game"`
		Name        string             `json:"name" db:"name"`
		Description []GenreDescription `json:"description,omitempty" db:"-"`
		//	DescLong    string   `json:"desc_long" db:"desc_long"`
		Characteristics []string `json:"keywords" db:"-"`
		ParentGenreID   *int64   `json:"parent_genre omitempty" db:"parent,omitempty"`
		Children        []int64  `json:"children,omitempty" db:"children,omitempty"`
	}

	GenreCharacteristics struct {
		ID         int64          `json:"id" db:"id,pk,autoinc"`
		Name       string         `json:"name" db:"name"`
		Descripion sql.NullString `json:"description,omitempty" db:"description"`
	}

	GenreDescription struct {
		GenreID     int64  `json:"genre_id" db:"genre_id"`
		Language    string `json:"language" db:"language"`
		Description string `json:"description" db:"description"`
	}

	MediaStorage struct {
		newDB *pgxpool.Pool
		db    *sqlx.DB // legacy
		Log   *zerolog.Logger
		ks    *KeywordStorage
		Ps    *PeopleStorage
	}
)

func NewMediaStorage(newDB *pgxpool.Pool, db *sqlx.DB, l *zerolog.Logger) *MediaStorage {
	ks := NewKeywordStorage(db, l)
	Ps := NewPeopleStorage(db, l)
	return &MediaStorage{newDB: newDB, db: db, Log: l, ks: ks, Ps: Ps}
}

// Get scans into a complete Media struct
// In most cases though, all we need is an intermediate, partial instance with the UUID and Kind fields
// to be passed to GetMediaDetails
func (ms *MediaStorage) Get(ctx context.Context, id uuid.UUID) (media Media, err error) {
	select {
	case <-ctx.Done():
		return Media{}, ctx.Err()
	default:
		stmt, err := ms.db.PrepareContext(ctx, "SELECT * FROM media.media WHERE id = $1")
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return Media{}, fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		row := stmt.QueryRowContext(ctx, id)
		err = row.Scan(
			&media.ID, &media.Title, &media.Kind, &media.Created, &media.Creator)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error scanning row")
			return Media{}, fmt.Errorf("error scanning row: %v", err)
		}
		return media, nil
	}
}

func (ms *MediaStorage) GetImagePath(ctx context.Context, id uuid.UUID) (path string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// TODO: add thumbnail paths
		err := ms.db.GetContext(ctx, &path, `SELECT i.source
			FROM media.media_images AS mi
			JOIN cdn.images AS i ON mi.image_id = i.id
			WHERE mi.media_id = $1
			LIMIT 1
			`, id)
		if err != nil {
			return "", err
		}

		return path, nil
	}
}

func (ms *MediaStorage) GetKind(ctx context.Context, id uuid.UUID) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		stmt, err := ms.db.PrepareContext(ctx, "SELECT kind FROM media.media WHERE id = $1")
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return "", fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		var kind string
		row := stmt.QueryRowContext(ctx, id)
		err = row.Scan(&kind)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error scanning row")
			return "", fmt.Errorf("error scanning row: %v", err)
		}
		return kind, nil
	}
}

// GetGenres returns all genres for specified media type.
// Generally to avoid overfetching, it's advisable to use GetGenreNames instead
// (accessed with optional query parameter ?names_only=true) (if this parameter is not provided,
// it uses true as a default value).
func (ms *MediaStorage) GetGenres(ctx context.Context, kind string) ([]Genre, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		stmt, err := ms.db.PreparexContext(ctx, `
		SELECT * FROM media.genres
		WHERE kind = $1
		`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		var genres []Genre
		err = stmt.SelectContext(ctx, &genres, kind)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error selecting rows")
			return nil, fmt.Errorf("error selecting rows: %v", err)
		}
		return genres, nil
	}
}

// TODO: add multilingual description support
func (ms *MediaStorage) GetGenre(ctx context.Context, kind, lang, name string) (genre *Genre, err error) {
	title := cases.Title(language.AmericanEnglish)
	name = title.String(strings.ReplaceAll(name, "_", " "))
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

		tx, err := ms.newDB.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("error beginning transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		genre := Genre{
			Kinds: pq.StringArray{kind},
		}
		rows, err := ms.newDB.Query(ctx, `
			SELECT name, parent, children FROM media.genres WHERE $1 = ANY(kinds) AND name = $2`,
			kind, name)
		ms.Log.Debug().Msgf("query: %v", rows)
		if err != nil {
			return nil, fmt.Errorf("error querying genre rows: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			if err = pgxscan.ScanRow(&genre, rows); err != nil {
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
		}
		dc, err := ms.newDB.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("error acquiring connection: %v", err)
		}
		defer dc.Release()

		var description string
		err = dc.QueryRow(ctx, `
			SELECT description FROM media.genre_descriptions	
			WHERE genre_id = (SELECT id FROM media.genres WHERE $1 = ANY(kinds) AND name = $2) 
			AND language = $3
					`, kind, name, lang).Scan(&description)
		if err != nil {
			return nil, fmt.Errorf("error querying rows for description: %v", err)
		}

		genre.Description = []GenreDescription{
			{
				Language:    lang,
				Description: description,
			},
		}

		ms.Log.Debug().Msgf("genre: %v", genre)

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %v", err)
		}
		return &genre, nil
	}
}

// GetGenreNames returns all genre names for specified media type, without any additional information.
func (ms *MediaStorage) GetGenreNames(ctx context.Context, kind string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		stmt, err := ms.db.PreparexContext(ctx, `
		SELECT name FROM media.genres
		WHERE kind = $1
		`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		var names []string
		err = stmt.SelectContext(ctx, &names, kind)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error selecting rows")
			return nil, fmt.Errorf("error selecting rows: %v", err)
		}
		return names, nil
	}
}

func (ms *MediaStorage) GetMediaDetails(
	ctx context.Context,
	mediaKind string,
	id uuid.UUID,
) (interface{}, error) {
	switch mediaKind {
	case "book":
		return ms.getBook(ctx, id)
	case "album":
		return ms.getAlbum(ctx, id)
	case "track":
		return ms.getTrack(ctx, id)
	case "film":
		return ms.getFilm(ctx, id)
	case "tv_show":
		return ms.getSeries(ctx, id)
	default:
		return nil, fmt.Errorf("unknown media kind")
	}
}

// mwks - media IDs with their corresponding kind
func (ms *MediaStorage) GetRandom(ctx context.Context, count int, blacklistKinds ...string) (
	mwks map[uuid.UUID]string, err error,
) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			ms.Log.Error().Msg("no database connection or nil pointer")
			return nil, fmt.Errorf("no database connection or nil pointer")
		}

		// prepare statement
		stmt, err := ms.db.PreparexContext(ctx,
			`SELECT id, kind
			FROM media.media 
			WHERE kind != ALL($1)
			ORDER BY RANDOM()
			LIMIT $2`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		// query
		rows, err := stmt.QueryxContext(ctx, pq.Array(blacklistKinds), count)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error querying rows")
			return nil, fmt.Errorf("error querying rows: %v", err)
		}
		defer rows.Close()

		// scan rows into map
		mwks = make(map[uuid.UUID]string)
		var (
			id   uuid.UUID
			kind string
		)
		for rows.Next() {
			if err := rows.Scan(&id, &kind); err != nil {
				ms.Log.Error().Err(err).Msg("error scanning row")
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
			mwks[id] = kind
		}
		return mwks, nil
	}
}

// Add is a generic method that adds an object to the media.media table. It needs to be run
// BEFORE the object is added to its respective table, since it needs the media ID to be
// generated first.
func (ms *MediaStorage) Add(ctx context.Context, props *Media) (mediaID *uuid.UUID, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			return nil, fmt.Errorf("no database connection or nil pointer")
		}
		stmt, err := ms.db.PreparexContext(ctx, `	
		INSERT INTO media.media (
			title, kind, created
		) VALUES (
			$1, $2, $3
		)
		RETURNING id
		`)
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		err = stmt.GetContext(ctx, mediaID, props.Title, props.Kind, props.Created)
		if err != nil {
			return nil, fmt.Errorf("error executing statement: %v", err)
		}
		err = ms.AddCreators(ctx, *mediaID, props.Creators)
		// handle the case in which the said person is not in the database
		if err == sql.ErrNoRows {
			ms.Log.Warn().Msg("no rows were affected")
			return nil, fmt.Errorf("no rows were affected: %v", err)
		}
		if err != nil {
			return nil, fmt.Errorf("error adding creators: %v", err)
		}
		return mediaID, nil
	}
}

func (ms *MediaStorage) AddCreators(ctx context.Context, uuid uuid.UUID, creators []Person) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			return fmt.Errorf("no database connection or nil pointer")
		}
		stmt, err := ms.db.PreparexContext(ctx, `	
		INSERT INTO media.media_creators (
			media_id, creator_id
		) VALUES (
			$1, $2
		)
		`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		for i := range creators {
			_, err = stmt.ExecContext(ctx, uuid, creators[i].ID)
			if err != nil {
				ms.Log.Error().Err(err).Msg("error executing statement")
				return fmt.Errorf("error executing statement: %v", err)
			}
		}
		return nil
	}
}

func (ms *MediaStorage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

func (ms *MediaStorage) Update(ctx context.Context, key string, value interface{}, mediaID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			return fmt.Errorf("no database connection or nil pointer")
		}
		// TODO: add a switch statement to handle different types of values
		stmt, err := ms.db.PreparexContext(ctx, `
		UPDATE media.media
		SET $1 = $2
		WHERE id = $3
		`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx, key, value, mediaID)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error executing statement")
			return fmt.Errorf("error executing statement: %v", err)
		}
		return nil
	}
}

func (ms *MediaStorage) Delete(ctx context.Context, mediaID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if ms.db == nil {
			return fmt.Errorf("no database connection or nil pointer")
		}
		stmt, err := ms.db.PreparexContext(ctx, `
		DELETE FROM media.media
		WHERE id = $1
		`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return fmt.Errorf("error preparing statement: %v", err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx, mediaID)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error executing statement")
			return fmt.Errorf("error executing statement: %v", err)
		}
		return nil
	}
}
