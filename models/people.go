package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"
)

type (
	Entity interface {
		GetID() int
	}

	Person struct {
		ID         int32        `json:"id" db:"id,pk,unique,autoincrement"`
		FirstName  string       `json:"first_name" db:"first_name"`
		OtherNames []string     `json:"other_names,omitempty" db:"other_names"`
		LastName   string       `json:"last_name" db:"last_name"`
		NickNames  []string     `json:"nick_name,omitempty" db:"nick_name"`
		Roles      []string     `json:"roles,omitempty" db:"roles"`
		Works      []*uuid.UUID `json:"works" db:"works"`
		Birth      sql.NullTime `json:"birth,omitempty" db:"birth"` // DOB can also be unknown
		Death      sql.NullTime `json:"death,omitempty" db:"death"`
		Website    string       `json:"website,omitempty" db:"website"`
		Bio        string       `json:"bio,omitempty" db:"bio"`
		Photos     []string     `json:"photos,omitempty" db:"photos"`
		Hometown   Place        `json:"hometown,omitempty" db:"hometown"`
		Residence  Place        `json:"residence,omitempty" db:"residence"`
		Added      time.Time    `json:"added" db:"added"`
		Modified   sql.NullTime `json:"modified,omitempty" db:"modified"`
	}

	Group struct {
		ID        int32        `json:"id" db:"id"`
		Locations []Place      `json:"locations,omitempty" db:"locations"`
		Name      string       `json:"name" db:"name"`
		Active    bool         `json:"active" db:"active"`
		Formed    sql.NullTime `json:"formed,omitempty" db:"formed"`
		Disbanded sql.NullTime `json:"disbanded,omitempty" db:"disbanded"`
		Website   string       `json:"website,omitempty" db:"website"`
		Photos    []string     `json:"photos,omitempty" db:"photos"`
		Works     []*uuid.UUID `json:"works,omitempty" db:"works"`
		Members   []Person     `json:"members,omitempty" db:"members"`
		Genres    []Genre      `json:"genres,omitempty" db:"genres"`
		Kind      string       `json:"kind,omitempty" db:"kind"` // Orchestra, Choir, Ensemble, Collective, etc.
		Added     time.Time    `json:"added" db:"added"`
		Modified  sql.NullTime `json:"modified,omitempty" db:"modified"`
	}

	Studio struct {
		ID           int32    `json:"id" db:"id,pk,serial,unique"`
		Name         string   `json:"name" db:"name"`
		Active       bool     `json:"active" db:"active"`
		City         *City    `json:"city,omitempty" db:"city"`
		Artists      []Person `json:"artists,omitempty" db:"artists"`
		Works        Media    `json:"works,omitempty" db:"works"`
		IsFilm       bool     `json:"is_film" db:"is_film"`
		IsMusic      bool     `json:"is_music" db:"is_music"`
		IsTV         bool     `json:"is_tv" db:"is_tv"`
		IsPublishing bool     `json:"is_publishing" db:"is_publishing"`
		IsGame       bool     `json:"is_game" db:"is_game"`
	}
)

var GroupKinds = []string{
	"Orchestra",
	"Choir",
	"Ensemble",
	"Collective",
	"Band",
	"Troupe",
	"Other",
}

func (g *Group) Validate() error {
	if lo.Contains(GroupKinds, g.Kind) {
		return nil
	}
	return fmt.Errorf("invalid group kind: %s, must be one of %s", g.Kind, strings.Join(GroupKinds, ", "))
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (p Person) GetID() int32 {
	return p.ID
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (g Group) GetID() int32 {
	return g.ID
}

//nolint:gocritic // we can't use pointer receivers to implement interfaces
func (s Studio) GetID() int32 {
	return s.ID
}
