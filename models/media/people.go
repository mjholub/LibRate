package media

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/models/places"
)

type studioKind string

const (
	FilmStudio     studioKind = "film"
	Music          studioKind = "music"
	Game           studioKind = "game"
	TV             studioKind = "tv"
	Publishing     studioKind = "publishing"
	VisualArtOther studioKind = "visual_art_other"
	Unknown        studioKind = "unknown"
)

type (
	StudioKind interface {
		// WARN: flashy background image on linked site
		// see https://vladimir.varank.in/notes/2023/09/compile-time-safety-for-enumerations-in-go/
		valid() bool
	}

	GroupedArtists struct {
		Individual []Person `json:"individual,omitempty" db:"individual"`
		Group      []Group  `json:"group,omitempty" db:"group"`
	}

	Person struct {
		ID         uuid.UUID      `json:"id,omitempty" db:"id,pk,unique" swaggertype:"string" example:"12345678-90ab-cdef-9876-543210fedcba"`
		Name       string         `json:"name,omitempty" db:"-"` // helper field for complete name
		FirstName  string         `json:"first_name" db:"first_name" example:"Karol"`
		OtherNames pq.StringArray `json:"other_names,omitempty" db:"other_names" example:"['Jan Paweł II']"`
		LastName   string         `json:"last_name" db:"last_name" example:"Wojtyła"`
		NickNames  pq.StringArray `json:"nick_names,omitempty" db:"nick_names" example:"['pawlacz', 'jan pawulon']"`
		Roles      pq.StringArray `json:"roles,omitempty" db:"roles"`
		Works      []*uuid.UUID   `json:"works,omitempty" db:"works"`
		Birth      sql.NullTime   `json:"birth,omitempty" db:"birth"` // DOB can also be unknown
		Death      sql.NullTime   `json:"death,omitempty" db:"death" example:"2005-04-02T21:37:00Z"`
		Website    sql.NullString `json:"website,omitempty" db:"website" example:"https://www.vatican.va/content/john-paul-ii/en.html"`
		Bio        sql.NullString `json:"bio,omitempty" db:"bio" example:"wojtyła disco dance"`
		Photos     pq.StringArray `json:"photos,omitempty" db:"photos"`
		Hometown   places.Place   `json:"hometown,omitempty" db:"hometown"`
		Residence  places.Place   `json:"residence,omitempty" db:"residence"`
		Added      time.Time      `json:"added,omitempty" db:"added"`
		Modified   sql.NullTime   `json:"modified,omitempty" db:"modified"`
	}

	Group struct {
		ID              uuid.UUID      `json:"id,omitempty" db:"id"`
		Locations       []places.Place `json:"locations,omitempty" db:"locations"`
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
		ID      int32        `json:"id" db:"id,pk,serial,unique"`
		Name    string       `json:"name" db:"name"`
		Active  bool         `json:"active" db:"active"`
		City    *places.City `json:"city,omitempty" db:"city"`
		Artists []Person     `json:"artists,omitempty" db:"artists"`
		Works   Media        `json:"works,omitempty" db:"works"`
		Kinds   []studioKind `json:"kinds,omitempty" db:"kinds"`
	}

	PeopleStorage struct {
		newDBConn *pgxpool.Pool
		// legacy
		dbConn *sqlx.DB
		logger *zerolog.Logger
	}
)

// see the top of the type block
func (s studioKind) valid() bool {
	return lo.Contains([]studioKind{
		FilmStudio,
		Music,
		Game,
		TV,
		Publishing,
		VisualArtOther,
		Unknown,
	}, s,
	)
}

func NewPeopleStorage(newConn *pgxpool.Pool, dbConn *sqlx.DB, logger *zerolog.Logger) *PeopleStorage {
	return &PeopleStorage{
		newDBConn: newConn,
		dbConn:    dbConn,
		logger:    logger,
	}
}
