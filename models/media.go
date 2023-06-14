package models

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

type (
	MediaService interface {
		IsMedia() bool // dummy placeholder so that we can have somewhat idiomatic parametric polymorphism
	}

	MediaStorer interface {
		Get(ctx context.Context, key string, kind interface{}) (any, error)
		GetAll() ([]*interface{}, error)
		Add(ctx context.Context, key, value interface{}, obj interface{}) error
		Update(ctx context.Context, key, value interface{}, obj interface{}) error
		Delete(ctx context.Context, key interface{}, obj interface{}) error
	}

	Media struct {
		UUID     uuid.UUID `json:"uuid" db:"uuid,pk,unique"`
		Kind     string    `json:"kind" db:"kind"`
		Name     string    `json:"name" db:"name"`
		Genres   []Genre   `json:"genres,omitempty" db:"genres"`
		Keywords []string  `json:"keywords,omitempty" db:"keywords"` // WARN: should this really be nullable?
		LangIDs  []int16   `json:"lang_ids,omitempty" db:"lang_ids"`
	}

	Book struct {
		MediaID         *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		Title           string     `json:"title" db:"title"`
		Authors         []Person   `json:"author" db:"author"`
		Publisher       string     `json:"publisher" db:"publisher"`
		PublicationDate time.Time  `json:"publication_date" db:"publication_date"`
		Genres          []string   `json:"genres" db:"genres"`
		Keywords        []string   `json:"keywords,omitempty" db:"keywords,omitempty"`
		Languages       []string   `json:"languages" db:"languages"`
		Pages           int16      `json:"pages" db:"pages"`
		ISBN            string     `json:"isbn,omitempty" db:"isbn,unique,omitempty"`
		ASIN            string     `json:"asin,omitempty" db:"asin,unique,omitempty"`
		Cover           string     `json:"cover,omitempty" db:"cover,omitempty"`
		Summary         string     `json:"summary" db:"summary"`
	}

	Genre struct {
		ID          int16    `json:"id" db:"id,pk,autoinc"`
		Name        string   `json:"name" db:"name"`
		DescShort   string   `json:"desc_short" db:"desc_short"`
		DescLong    string   `json:"desc_long" db:"desc_long"`
		Keywords    []string `json:"keywords" db:"keywords"`
		ParentGenre *Genre   `json:"parent_genre omitempty" db:"parent_genre"`
		Children    []Genre  `json:"children omitempty" db:"children"`
	}

	MediaStorage struct{}
)

func NewMediaStorage() *MediaStorage {
	return &MediaStorage{}
}

func (ms *MediaStorage) Get(ctx context.Context, key string, kind interface{}) (any, error) {
	return nil, nil
}

func (ms *MediaStorage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

func (ms *MediaStorage) Add(ctx context.Context, media Media) error {
	return nil
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
