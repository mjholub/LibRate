package models

import "github.com/gofrs/uuid/v5"

type Place struct {
	UUID    uuid.UUID `json:"uuid" db:"uuid,pk"`
	Kind    string    `json:"kind" db:"kind"`
	Name    string    `json:"name" db:"name"`
	Lat     float64   `json:"lat" db:"lat"`
	Lng     float64   `json:"lng" db:"lng"`
	Country *Country  `json:"country" db:"country"`
}

type Country struct {
	ID   int16  `json:"id" db:"id,pk"`
	Name string `json:"name" db:"name"`
	Code string `json:"code" db:"code"`
}

type City struct {
	UUID    uuid.UUID `json:"uuid" db:"uuid,pk"`
	Name    string    `json:"name" db:"name"`
	Lat     float64   `json:"lat" db:"lat"`
	Lng     float64   `json:"lng" db:"lng"`
	Country *Country  `json:"country" db:"country"`
}

type Venue struct {
	UUID    uuid.UUID `json:"uuid" db:"uuid,pk"`
	Name    string    `json:"name" db:"name"`
	Active  bool      `json:"active" db:"active"`
	Street  string    `json:"street" db:"street"`
	Zip     string    `json:"zip" db:"zip"`
	Unit    string    `json:"unit" db:"unit"`
	City    *City     `json:"city" db:"city"`
	Country *Country  `json:"country" db:"country"`
}
