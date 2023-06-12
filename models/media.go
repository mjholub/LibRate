package models

import (
	"context"
	"time"
)

type Film struct {
	ID    int    `json:"id" db:"id"`
	Title string `json:"title"`
	Cast  Cast
	Year  int `json:"year"`
}

type TVShow struct {
	ID      int      `json:"id" db:"id"`
	Title   string   `json:"title" db:"title"`
	Cast    Cast     `json:"cast" db:"cast"`
	Year    int      `json:"year" db:"year"`
	Active  bool     `json:"active" db:"active"`
	Seasons []Season `json:"seasons" db:"seasons"`
}

type Season struct {
	ID       int       `json:"id" db:"id"`
	ShowID   int       `json:"show_id" db:"show_id"`
	Number   uint16    `json:"number" db:"number"`
	Episodes []Episode `json:"episodes" db:"episodes"`
}

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

type Episode struct {
	ID        int           `json:"id" db:"id"`
	Title     string        `json:"title" db:"title"`
	Season    uint16        `json:"season" db:"season"`
	Episode   uint16        `json:"episode" db:"episode"`
	AirDate   time.Time     `json:"air_date" db:"air_date"`
	Duration  time.Duration `json:"duration" db:"duration"`
	Languages []string      `json:"languages" db:"languages"`
	Plot      string        `json:"plot" db:"plot"`
}

type Cast struct {
	Actors    []Person `json:"actors"`
	Directors []Person `json:"directors"`
}

type Album struct {
	ID          int           `json:"id" db:"id"`
	Name        string        `json:"name"`
	Artists     []Person      `json:"artists"`
	ReleaseDate time.Time     `json:"release_date"`
	Genres      []string      `json:"genres"`
	Keywords    []string      `json:"keywords"`
	Duration    time.Duration `json:"duration"`
	Tracks      []Track       `json:"tracks"`
	Languages   []string      `json:"languages" db:"languages"`
}

type Track struct {
	ID        int           `json:"id" db:"id"`
	Name      string        `json:"name"`
	Abum      Album         `json:"album"`
	Artists   []Person      `json:"artists"`
	Duration  time.Duration `json:"duration"`
	Lyrics    string        `json:"lyrics"`
	Languages []string      `json:"languages"`
}

type Genre struct {
	ID          int      `json:"id" db:"id"`
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

func (ms *MediaStorage) Add(ctx context.Context, media interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Update(ctx context.Context, key, value interface{}, objType interface{}) error {
	return nil
}

func (ms *MediaStorage) Delete(ctx context.Context, key interface{}, objType interface{}) error {
	return nil
}
