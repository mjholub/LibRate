package places

import "github.com/gofrs/uuid/v5"

type (
	Place struct {
		ID      uint64   `json:"id" db:"id,pk"`
		Kind    string   `json:"kind" db:"kind"`
		Name    string   `json:"name" db:"name"`
		Lat     float64  `json:"lat" db:"lat"`
		Lng     float64  `json:"lng" db:"lng"`
		Country *Country `json:"country" db:"country"`
	}

	Country struct {
		ID   int16  `json:"id" db:"id,pk"`
		Name string `json:"name" db:"name"`
		Code string `json:"code" db:"code"`
	}

	City struct {
		UUID    uuid.UUID `json:"uuid" db:"uuid,pk"`
		Name    string    `json:"name" db:"name"`
		Lat     float64   `json:"lat" db:"lat"`
		Lng     float64   `json:"lng" db:"lng"`
		Country *Country  `json:"country" db:"country"`
	}

	Venue struct {
		UUID    uuid.UUID `json:"uuid" db:"uuid,pk"`
		Name    string    `json:"name" db:"name"`
		Active  bool      `json:"active" db:"active"`
		Street  string    `json:"street" db:"street"`
		Zip     string    `json:"zip" db:"zip"`
		Unit    string    `json:"unit" db:"unit"`
		City    *City     `json:"city" db:"city"`
		Country *Country  `json:"country" db:"country"`
	}
)
