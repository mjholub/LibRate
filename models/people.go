package models

import "github.com/samber/mo"

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
	Hometown   string   `json:"hometown,omitempty" db:"hometown"`
	Residence  string   `json:"residence,omitempty" db:"residence"`
	MemberOf   []Group  `json:"member_of,omitempty" db:"member_of"`
}

type Group struct {
	ID        int      `json:"id" db:"id"`
	UUID      string   `json:"uuid" db:"uuid"`
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

type Studio struct {
	ID           int                                   `json:"id" db:"id"`
	UUID         string                                `json:"uuid" db:"uuid"`
	Name         string                                `json:"name" db:"name"`
	Active       bool                                  `json:"active" db:"active"`
	City         string                                `json:"city,omitempty" db:"city"`
	Artists      []Person                              `json:"artists,omitempty" db:"artists"`
	Works        mo.Either4[Book, Film, TVShow, Album] `json:"works,omitempty" db:"works"`
	IsFilm       bool                                  `json:"is_film" db:"is_film"`
	IsMusic      bool                                  `json:"is_music" db:"is_music"`
	IsTV         bool                                  `json:"is_tv" db:"is_tv"`
	IsPublishing bool                                  `json:"is_publishing" db:"is_publishing"`
}
