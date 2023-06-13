package models

type Place struct {
	ID      int     `json:"id" db:"id"`
	Name    string  `json:"name" db:"name"`
	Lat     float64 `json:"lat" db:"lat"`
	Lng     float64 `json:"lng" db:"lng"`
	Country string  `json:"country" db:"country"`
}
