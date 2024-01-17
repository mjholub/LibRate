package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		ID         uuid.UUID      `json:"id,omitempty" db:"id,pk,unique"`
		FirstName  string         `json:"first_name" db:"first_name"`
		OtherNames pq.StringArray `json:"other_names,omitempty" db:"other_names"`
		LastName   string         `json:"last_name" db:"last_name"`
		NickNames  pq.StringArray `json:"nick_names,omitempty" db:"nick_names"`
		Roles      pq.StringArray `json:"roles,omitempty" db:"roles"`
		Works      []*uuid.UUID   `json:"works,omitempty" db:"works"`
		Birth      sql.NullTime   `json:"birth,omitempty" db:"birth"` // DOB can also be unknown
		Death      sql.NullTime   `json:"death,omitempty" db:"death"`
		Website    sql.NullString `json:"website,omitempty" db:"website"`
		Bio        sql.NullString `json:"bio,omitempty" db:"bio"`
		Photos     pq.StringArray `json:"photos,omitempty" db:"photos"`
		Hometown   Place          `json:"hometown,omitempty" db:"hometown"`
		Residence  Place          `json:"residence,omitempty" db:"residence"`
		Added      time.Time      `json:"added,omitempty" db:"added"`
		Modified   sql.NullTime   `json:"modified,omitempty" db:"modified"`
	}

	Group struct {
		ID              uuid.UUID      `json:"id,omitempty" db:"id"`
		Locations       []Place        `json:"locations,omitempty" db:"locations"`
		Name            string         `json:"name" db:"name"`
		Active          bool           `json:"active,omitempty" db:"active"`
		Formed          sql.NullTime   `json:"formed,omitempty" db:"formed"`
		Disbanded       sql.NullTime   `json:"disbanded,omitempty" db:"disbanded"`
		Website         sql.NullString `json:"website,omitempty" db:"website"`
		Photos          []string       `json:"photos,omitempty" db:"photos"`
		Works           []*uuid.UUID   `json:"works,omitempty" db:"works"`
		Members         []Person       `json:"members,omitempty" db:"members"`
		PrimaryGenre    Genre          `json:"primary_genre,omitempty" db:"primary_genre_id"`
		SecondaryGenres []Genre        `json:"genres,omitempty" db:"genres"`
		Kind            string         `json:"kind,omitempty" db:"kind"` // Orchestra, Choir, Ensemble, Collective, etc.
		Added           time.Time      `json:"added" db:"added"`
		Modified        sql.NullTime   `json:"modified,omitempty" db:"modified"`
		Wikipedia       sql.NullString `json:"wikipedia,omitempty" db:"wikipedia"`
		Bandcamp        sql.NullString `json:"bandcamp,omitempty" db:"bandcamp"`
		Soundcloud      sql.NullString `json:"soundcloud,omitempty" db:"soundcloud"`
		Bio             sql.NullString `json:"bio,omitempty" db:"bio"`
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
		newDBConn *pgxpool.Pool
		// legacy
		dbConn *sqlx.DB
		logger *zerolog.Logger
	}
)

func NewPeopleStorage(newConn *pgxpool.Pool, dbConn *sqlx.DB, logger *zerolog.Logger) *PeopleStorage {
	return &PeopleStorage{
		newDBConn: newConn,
		dbConn:    dbConn,
		logger:    logger,
	}
}

func (p *PeopleStorage) GetPersonNames(ctx context.Context, id int32) (Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return Person{}, ctx.Err()
	default:
		err := p.dbConn.Get(&person, "SELECT first_name, last_name, other_names, nick_names FROM people.person WHERE id = $1", id)
		if err != nil {
			return Person{}, err
		}
		return person, nil
	}
}

func (p *PeopleStorage) GetPerson(ctx context.Context, id int64) (Person, error) {
	var person Person
	select {
	case <-ctx.Done():
		return Person{}, ctx.Err()
	default:
		err := p.dbConn.Get(&person, "SELECT * FROM people.person WHERE id = $1", id)
		if err != nil {
			return Person{}, err
		}
		return person, nil
	}
}

func (p *PeopleStorage) GetGroup(ctx context.Context, id int32) (Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return Group{}, ctx.Err()
	default:
		err := p.dbConn.Get(&group, "SELECT * FROM people.group WHERE id = $1", id)
		if err != nil {
			return Group{}, err
		}
		return group, nil
	}
}

// FIXME: "scannable dest type slice with >1 columns (12) in result"
func (p *PeopleStorage) GetArtistsByName(ctx context.Context, name string) (persons []Person, groups []Group, err error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		rows, err := p.newDBConn.Query(ctx, `SELECT *
FROM people.person
WHERE (first_name LIKE $1 OR last_name LIKE $1) OR $1 LIKE ANY(nick_names)`, name)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var person Person
			if err := rows.Scan(&person.ID, &person.FirstName, &person.OtherNames, &person.LastName,
				&person.NickNames, &person.Roles, &person.Works, &person.Birth, &person.Death,
				&person.Website, &person.Bio, &person.Photos, &person.Hometown, &person.Residence,
				&person.Added, &person.Modified); err != nil {
				return nil, nil, err
			}
			persons = append(persons, person)
		}

		if err = rows.Err(); err != nil {
			return nil, nil, err
		}

		// Query for groups
		rows, err = p.newDBConn.Query(ctx, "SELECT * FROM people.group WHERE name LIKE $1", name)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var group Group
			if err := rows.Scan(&group.ID, &group.Locations, &group.Name, &group.Active,
				&group.Formed, &group.Disbanded, &group.Website, &group.Photos, &group.Works,
				&group.Members, &group.PrimaryGenre, &group.SecondaryGenres, &group.Kind,
				&group.Added, &group.Modified, &group.Wikipedia, &group.Bandcamp, &group.Soundcloud, &group.Bio); err != nil {
				return nil, nil, err
			}
			groups = append(groups, group)
		}

		if err := rows.Err(); err != nil {
			return nil, nil, err
		}

		return persons, groups, nil
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

func (p *PeopleStorage) GetGroupName(ctx context.Context, id int32) (Group, error) {
	var group Group
	select {
	case <-ctx.Done():
		return Group{}, ctx.Err()
	default:
		err := p.dbConn.Get(&group, "SELECT name FROM people.group WHERE id = $1", id)
		if err != nil {
			return Group{}, err
		}
		return group, nil
	}
}

func (g *Group) Validate() error {
	GroupKinds := []string{
		"Orchestra",
		"Choir",
		"Ensemble",
		"Collective",
		"Band",
		"Troupe",
		"Other",
	}
	if lo.Contains(GroupKinds, g.Kind) {
		return nil
	}
	return fmt.Errorf("invalid group kind: %s, must be one of %s", g.Kind, strings.Join(GroupKinds, ", "))
}

func (p *PeopleStorage) GetID(ctx context.Context, name, kind string) (id int32, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		switch kind {
		case "group":
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM people.group WHERE name = $1 AND kind = $2 LIMIT 1",
				name, kind)
			if err != nil {
				return 0, err
			}
			return id, nil
		case "person":
			firstName := strings.Split(name, " ")[0]
			lastName := strings.Split(name, " ")[1]
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM people.person WHERE first_name = $1 AND last_name = $2 LIMIT 1",
				firstName, lastName)
			if err != nil {
				return 0, err
			}
			return id, nil
		case "studio":
			err := p.dbConn.GetContext(ctx, &id,
				"SELECT id FROM people.studio WHERE name = $1 LIMIT 1",
				name)
			if err != nil {
				return 0, err
			}
			return id, nil
		default:
			return 0, fmt.Errorf("invalid kind: %s", kind)
		}
	}
}
