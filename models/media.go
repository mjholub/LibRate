package models

import (
	"context"
	"time"
)

type Film struct {
	Title string `json:"title"`
	Cast  Cast
	Year  uint32 `json:"year"`
}

type Cast struct {
	Actors    []Person `json:"actors"`
	Directors []Person `json:"directors"`
}

type Album struct {
	Name        string        `json:"name"`
	Artists     []Person      `json:"artists"`
	ReleaseDate time.Time     `json:"release_date"`
	Genres      []string      `json:"genres"`
	Keywords    []string      `json:"keywords"`
	Duration    time.Duration `json:"duration"`
	Tracks      []Track       `json:"tracks"`
}

type Track struct {
	Name     string        `json:"name"`
	Artists  []Person      `json:"artists"`
	Duration time.Duration `json:"duration"`
	Lyrics   string        `json:"lyrics"`
}

type Genre struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
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

func (ms *MediaStorage) Add(ctx context.Context, key, value interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Update(ctx context.Context, key, value interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Delete(ctx context.Context, key interface{}, objType interface{}) error {
	return nil
}
