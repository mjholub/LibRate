package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type (
	Entity interface {
		GetID() int
	}

	Person struct {
		ID         int32          `json:"id" db:"id,pk,unique,autoincrement"`
		FirstName  string         `json:"first_name" db:"first_name"`
		OtherNames pq.StringArray `json:"other_names,omitempty" db:"other_names"`
		LastName   string         `json:"last_name" db:"last_name"`
		NickNames  pq.StringArray `json:"nick_names,omitempty" db:"nick_names"`
		Roles      pq.StringArray `json:"roles,omitempty" db:"roles"`
		Works      []*uuid.UUID   `json:"works" db:"works"`
		Birth      sql.NullTime   `json:"birth,omitempty" db:"birth"` // DOB can also be unknown
		Death      sql.NullTime   `json:"death,omitempty" db:"death"`
		Website    string         `json:"website,omitempty" db:"website"`
		Bio        string         `json:"bio,omitempty" db:"bio"`
		Photos     pq.StringArray `json:"photos,omitempty" db:"photos"`
		Hometown   Place          `json:"hometown,omitempty" db:"hometown"`
		Residence  Place          `json:"residence,omitempty" db:"residence"`
		Added      time.Time      `json:"added" db:"added"`
		Modified   sql.NullTime   `json:"modified,omitempty" db:"modified"`
	}

	Group struct {
		ID              int32        `json:"id" db:"id"`
		Locations       []Place      `json:"locations,omitempty" db:"locations"`
		Name            string       `json:"name" db:"name"`
		Active          bool         `json:"active" db:"active"`
		Formed          sql.NullTime `json:"formed,omitempty" db:"formed"`
		Disbanded       sql.NullTime `json:"disbanded,omitempty" db:"disbanded"`
		Website         string       `json:"website,omitempty" db:"website"`
		Photos          []string     `json:"photos,omitempty" db:"photos"`
		Works           []*uuid.UUID `json:"works,omitempty" db:"works"`
		Members         []Person     `json:"members,omitempty" db:"members"`
		PrimaryGenre    Genre        `json:"primary_genre,omitempty" db:"primary_genre_id"`
		SecondaryGenres []Genre      `json:"genres,omitempty" db:"genres"`
		Kind            string       `json:"kind,omitempty" db:"kind"` // Orchestra, Choir, Ensemble, Collective, etc.
		Added           time.Time    `json:"added" db:"added"`
		Modified        sql.NullTime `json:"modified,omitempty" db:"modified"`
		Wikipedia       string       `json:"wikipedia,omitempty" db:"wikipedia"`
		Bandcamp        string       `json:"bandcamp,omitempty" db:"bandcamp"`
		Soundcloud      string       `json:"soundcloud,omitempty" db:"soundcloud"`
		Bio             string       `json:"bio,omitempty" db:"bio"`
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

	PeopleStorage struct {
		dbConn *sqlx.DB
		logger *zerolog.Logger
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

func NewPeopleStorage(dbConn *sqlx.DB, logger *zerolog.Logger) *PeopleStorage {
	return &PeopleStorage{
		dbConn: dbConn,
		logger: logger,
	}
}

func (p *PeopleStorage) GetPersonNames(ctx context.Context, id int32) (*Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := p.dbConn.Get(&person, "SELECT first_name, last_name, other_names, nick_names FROM people.person WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &person, nil
	}
}

func (p *PeopleStorage) GetPerson(ctx context.Context, id int32) (*Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := p.dbConn.Get(&person, "SELECT * FROM people.person WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &person, nil
	}
}

func (p *PeopleStorage) GetGroup(ctx context.Context, id int32) (*Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := p.dbConn.Get(&group, "SELECT * FROM people.group WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &group, nil
	}
}

func (p *PeopleStorage) GetStudio(ctx context.Context, id int32) (*Studio, error) {
	var studio Studio
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := p.dbConn.Get(&studio, "SELECT * FROM people.studio WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &studio, nil
	}
}

func (p *PeopleStorage) GetGroupName(ctx context.Context, id int32) (*Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := p.dbConn.Get(&group, "SELECT name FROM people.group WHERE id = $1", id)
		if err != nil {
			return nil, err
		}
		return &group, nil
	}
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
