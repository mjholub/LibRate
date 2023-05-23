package models

import (
	"time"
)

type Film struct {
	Cast Cast
	Year uint32 `json:"year"`
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
