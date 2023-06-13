package models

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type Entity interface {
	GetID() int
}

func (p Person) GetID() int {
	return p.ID
}

func (g Group) GetID() int {
	return g.ID
}

func (s Studio) GetID() int {
	return s.ID
}

type Person struct {
	ID         int      `json:"id" db:"id"`
	UUID       string   `json:"uuid" db:"uuid"`
	FirstName  string   `json:"first_name" db:"first_name"`
	OtherNames string   `json:"other_names,omitempty" db:"other_names"`
	LastName   string   `json:"last_name" db:"last_name"`
	NickName   string   `json:"nick_name,omitempty" db:"nick_name"`
	Roles      []string `json:"roles,omitempty" db:"roles"`
	Works      []string `json:"works" db:"works"`
	Birth      string   `json:"birth" db:"birth"`
	Death      string   `json:"death,omitempty" db:"death"`
	Website    string   `json:"website,omitempty" db:"website"`
	Photos     []string `json:"photos,omitempty" db:"photos"`
	Hometown   Place    `json:"hometown,omitempty" db:"hometown"`
	Residence  Place    `json:"residence,omitempty" db:"residence"`
}

type Group struct {
	ID        int      `json:"id" db:"id"`
	UUID      string   `json:"uuid" db:"uuid"`
	Locations []Place  `json:"locations,omitempty" db:"locations"`
	Name      string   `json:"name" db:"name"`
	Active    bool     `json:"active" db:"active"`
	Formed    string   `json:"formed,omitempty" db:"formed"`
	Disbanded string   `json:"disbanded,omitempty" db:"disbanded"`
	Website   string   `json:"website,omitempty" db:"website"`
	Photos    []string `json:"photos,omitempty" db:"photos"`
	Works     []Album  `json:"works,omitempty" db:"works"`
	Members   []Person `json:"members,omitempty" db:"members"`
	Genres    []Genre  `json:"genres,omitempty" db:"genres"`
	Countries []string `json:"countries,omitempty" db:"countries"`
	Plays     []string `json:"plays,omitempty" db:"plays"`
	Kind      string   `json:"kind,omitempty" db:"kind"` // Orchestra, Choir, Ensemble, Collective, etc.
}

// FIXME: find a better workaround for go's shitty immutability support
var (
	GroupKinds = []string{
		"Orchestra",
		"Choir",
		"Ensemble",
		"Collective",
		"Band",
		"Troupe",
		"Other",
	}
)

func (g Group) Validate() error {
	if lo.Contains(GroupKinds, g.Kind) {
		return nil
	}
	return fmt.Errorf("invalid group kind: %s, must be one of %s", g.Kind, strings.Join(GroupKinds, ", "))
}

type Studio struct {
	ID           int      `json:"id" db:"id"`
	UUID         string   `json:"uuid" db:"uuid"`
	Name         string   `json:"name" db:"name"`
	Active       bool     `json:"active" db:"active"`
	City         string   `json:"city,omitempty" db:"city"`
	Artists      []Person `json:"artists,omitempty" db:"artists"`
	Works        Media    `json:"works,omitempty" db:"works"`
	IsFilm       bool     `json:"is_film" db:"is_film"`
	IsMusic      bool     `json:"is_music" db:"is_music"`
	IsTV         bool     `json:"is_tv" db:"is_tv"`
	IsPublishing bool     `json:"is_publishing" db:"is_publishing"`
}
