package models

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type (
	MediaService interface {
		IsMedia() bool // dummy placeholder so that we can have somewhat idiomatic parametric polymorphism
	}

	MediaStorer[T any] interface {
		Get(ctx context.Context, key string) (T, error)
		GetAll() ([]T, error)
		Add(ctx context.Context, key, value T) error
		Update(ctx context.Context, key string, value T) error
		Delete(ctx context.Context, key string) error
	}

	Media struct {
		ID      uuid.UUID `json:"id" db:"id,pk,unique"`
		Title   string    `json:"title" db:"title"`
		Kind    string    `json:"kind" db:"kind"`
		Created time.Time `json:"keywords,omitempty" db:"created"`
		Creator int32     `json:"creator,omitempty" db:"creator"`
	}

	MediaObject interface {
		Book | Album | Track | TVShow | Season | Episode
	}

	// Genre does not hage a UUID due to parent-child relationships
	Genre struct {
		ID          int16    `json:"id" db:"id,pk,autoinc"`
		Name        string   `json:"name" db:"name"`
		DescShort   string   `json:"desc_short" db:"desc_short"`
		DescLong    string   `json:"desc_long" db:"desc_long"`
		Keywords    []string `json:"keywords" db:"keywords"`
		ParentGenre *Genre   `json:"parent_genre omitempty" db:"parent"`
		Children    []Genre  `json:"children omitempty" db:"children"`
	}

	MediaStorage struct {
		db  *sqlx.DB
		Log *zerolog.Logger
		ks  *KeywordStorage
	}
)

// strictly necessary (not nil keys for each media type)
var (
	BookKeys = []string{
		"media_id", "title", "authors",
		"genres", "edition", "languages",
	}
	AlbumKeys = [7]string{
		"media_id", "title", "artists", "genres", "keywords", "languages", "cover",
	}
	TrackKeys = [7]string{
		"media_id", "title", "artists", "genres", "keywords", "languages", "cover",
	}
	GenreKeys = [5]string{
		"id", "name", "desc_short", "desc_long", "keywords",
	}
)

func NewMediaStorage(db *sqlx.DB, l *zerolog.Logger) *MediaStorage {
	ks := NewKeywordStorage(db, l)
	return &MediaStorage{db: db, Log: l, ks: ks}
}

// Get scans into a complete Media struct
// In most cases though, all we need is an intermediate, partial instance with the UUID and Kind fields
// to be passed to GetMediaDetails
func (ms *MediaStorage) Get(ctx context.Context, id uuid.UUID) (media Media, err error) {
	select {
	case <-ctx.Done():
		return Media{}, ctx.Err()
	default:
		stmt, err := ms.db.PrepareContext(ctx, "SELECT * FROM media WHERE uuid = ?")
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return Media{}, fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		row := stmt.QueryRowContext(ctx, id)
		err = row.Scan(
			&media.ID, &media.Title, &media.Kind, &media.Created, &media.Creator)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error scanning row")
			return Media{}, fmt.Errorf("error scanning row: %w", err)
		}
		return media, nil
	}
}

func (ms *MediaStorage) GetKind(ctx context.Context, id uuid.UUID) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		stmt, err := ms.db.PrepareContext(ctx, "SELECT kind FROM media WHERE uuid = ?")
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return "", fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		var kind string
		row := stmt.QueryRowContext(ctx, id)
		err = row.Scan(&kind)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error scanning row")
			return "", fmt.Errorf("error scanning row: %w", err)
		}
		return kind, nil
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

func (ms *MediaStorage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

// mwks - media IDs with their corresponding kind
func (ms *MediaStorage) GetRandom(ctx context.Context, count int) (
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
			ORDER BY RANDOM()
			LIMIT $1`)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error preparing statement")
			return nil, fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		// query
		rows, err := stmt.QueryxContext(ctx, count)
		if err != nil {
			ms.Log.Error().Err(err).Msg("error querying rows")
			return nil, fmt.Errorf("error querying rows: %w", err)
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
				return nil, fmt.Errorf("error scanning row: %w", err)
			}
			mwks[id] = kind
		}
		return mwks, nil
	}
}

func (ms *MediaStorage) Add(ctx context.Context, db *sqlx.DB, media MediaService, props Media) error {
	switch m := media.(type) {
	case *Book:
		return addBook(ctx, db, BookKeys, *m)
	case *Album:
		return addAlbum(ctx, db, *m)
	case *Track:
		return addTrack(ctx, db, m)
	default:
		return fmt.Errorf("unknown media type")
	}
}

func (ms *MediaStorage) Update(ctx context.Context, key, value interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Delete(ctx context.Context, key interface{}, objType interface{}) error {
	return nil
}

func (b *Book) GetMedia(db *sqlx.DB) (m *Media, err error) {
	if b.MediaID == nil {
		return nil, fmt.Errorf("book has no media id")
	}
	err = db.Get(m, "SELECT * FROM media WHERE uuid = $1", b.MediaID)
	if err != nil {
		return nil, err
	}
	return m, nil
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (b Book) IsMedia() bool {
	return true
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (a Album) IsMedia() bool {
	return true
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (t Track) IsMedia() bool {
	return true
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (g Genre) IsMedia() bool {
	return false
}
