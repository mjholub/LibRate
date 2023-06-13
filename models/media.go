package models

import (
	"context"
	"time"
)

type Book struct {
	ID              int       `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Authors         []Person  `json:"author" db:"author"`
	Publisher       string    `json:"publisher" db:"publisher"`
	PublicationDate time.Time `json:"publication_date" db:"publication_date"`
	Genres          []string  `json:"genres" db:"genres"`
	Keywords        []string  `json:"keywords" db:"keywords"`
	Languages       []string  `json:"languages" db:"languages"`
	Pages           uint16    `json:"pages" db:"pages"`
	ISBN            string    `json:"isbn" db:"isbn"`
	ASIN            string    `json:"asin" db:"asin"`
	Cover           string    `json:"cover" db:"cover"`
	Summary         string    `json:"summary" db:"summary"`
}

type Genre struct {
	ID          int      `json:"id" db:"id"`
	Name        string   `json:"name"`
	DescShort   string   `json:"desc_short", db:"desc_short"`
	DescLong    string   `json:"desc_long", db:"desc_long"`
	Keywords    []string `json:"keywords", db:"keywords"`
	ParentGenre *Genre   `json:"parent_genre, omitempty", db:"parent_genre"`
	Children    []Genre  `json:"children, omitempty", db:"children"`
}

type MediaStorer interface {
	Get(ctx context.Context, key string, kind interface{}) (any, error)
	GetAll() ([]*interface{}, error)
	Add(ctx context.Context, key, value interface{}, objType interface{}) error
	Update(ctx context.Context, key, value interface{}, objType interface{}) error
	Delete(ctx context.Context, key interface{}, objType interface{}) error
}

type MediaStorage struct{}

func NewMediaStorage() *MediaStorage {
	return &MediaStorage{}
}

func (ms *MediaStorage) Get(ctx context.Context, key string, kind interface{}) (any, error) {
	return nil, nil
}

func (ms *MediaStorage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

func (ms *MediaStorage) Add(ctx context.Context, media interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Update(ctx context.Context, key, value interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Delete(ctx context.Context, key interface{}, objType interface{}) error {
	return nil
}
