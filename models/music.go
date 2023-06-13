package models

import (
	"time"

	"github.com/samber/mo"
)

type Album struct {
	ID          int                          `json:"id" db:"id"`
	Name        string                       `json:"name" db:"name"`
	Artists     mo.Either[[]Person, []Group] `json:"artists" db:"artists"`
	ReleaseDate time.Time                    `json:"release_date" db:"release_date"`
	Genres      []Genre                      `json:"genres,omitempty" db:"genres"`
	Studio      Studio                       `json:"studio,omitempty" db:"studio"`
	Keywords    []string                     `json:"keywords,omitempty" db:"keywords"`
	Duration    time.Duration                `json:"duration" db:"duration"`
	Tracks      []Track                      `json:"tracks" db:"tracks"`
	Languages   []string                     `json:"languages" db:"languages,omitempty" db:"languages"`
}

type Track struct {
	ID        int                          `json:"id" db:"id"`
	Name      string                       `json:"name" db:"name"`
	Album     Album                        `json:"album" db:"album"`
	Artists   mo.Either[[]Person, []Group] `json:"artists" db:"artists"`
	Duration  time.Duration                `json:"duration" db:"duration"`
	Lyrics    string                       `json:"lyrics,omitempty" db:"lyrics"`
	Languages []string                     `json:"languages,omitempty" db:"languages"`
}
