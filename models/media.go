package models

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
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
		UUID     uuid.UUID `json:"uuid" db:"uuid,pk,unique"`
		Kind     string    `json:"kind" db:"kind"`
		Name     string    `json:"name" db:"name"`
		Genres   []Genre   `json:"genres,omitempty" db:"genres"`
		Keywords []string  `json:"keywords,omitempty" db:"keywords"` // WARN: should this really be nullable?
		LangIDs  []int16   `json:"lang_ids,omitempty" db:"lang_ids"`
		Creators []Person  `json:"creators,omitempty" db:"creators"`
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
		db *sqlx.DB
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

func NewMediaStorage(db *sqlx.DB) *MediaStorage {
	return &MediaStorage{}
}

func (ms *MediaStorage) Get(ctx context.Context, id uuid.UUID) (media any, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		stmt, err := ms.db.PrepareContext(ctx, "SELECT kind FROM media WHERE uuid = ?")
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		row := stmt.QueryRowContext(ctx, id)
		var kind string
		if err := row.Scan(&kind); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		switch kind {
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
}

func (ms *MediaStorage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

func (ms *MediaStorage) GetRandom(ctx context.Context, count int) (media []*Media, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		stmt, err := ms.db.PreparexContext(ctx, "SELECT * FROM media ORDER BY RANDOM() LIMIT ?")
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		rows, err := stmt.QueryxContext(ctx, count)
		if err != nil {
			return nil, fmt.Errorf("error querying rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var m Media
			if err := rows.StructScan(&m); err != nil {
				return nil, fmt.Errorf("error scanning row: %w", err)
			}
			media = append(media, &m)
		}
		return media, nil
	}
}

func (ms *MediaStorage) Add(ctx context.Context, db *sqlx.DB, media MediaService, props Media) error {
	switch m := media.(type) {
	case *Book:
		return addBook(ctx, db, BookKeys[:], *m)
	case *Album:
		return addAlbum(ctx, db, *m)
	case *Track:
		return addTrack(ctx, db, *m)
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
