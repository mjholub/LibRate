package places

import "github.com/gofrs/uuid/v5"

type (
	Place struct {
		ID      uint64   `json:"id" db:"id,pk" fake:"{number:1,9001}"`
		Kind    string   `json:"kind" db:"kind" fake:"{randomstring:[city,venue]}" swaggertype:"string" example:"city" enums:"city,venue"`
		Name    string   `json:"name" db:"name" fake:"{city}" swaggertype:"string" example:"Tokyo"`
		Lat     float64  `json:"lat" db:"lat" fake:"{latitude}" example:"35.6895" swaggertype:"number" format:"float"`
		Lng     float64  `json:"lng" db:"lng" fake:"{longitude}" example:"139.6917" swaggertype:"number" format:"float"`
		Country *Country `json:"country" db:"country" fake:"{country}" example:"Japan"`
	}

	CityName struct {
		ID       uuid.UUID `json:"id" db:"id" fake:"{uuid}" swaggertype:"string" example:"12345678-90ab-cdef-9876-543210fedcba"`
		Name     string    `json:"name" db:"name" fake:"{city}" swaggertype:"string" example:"東京"`
		Language string    `json:"language" db:"lang" fake:"{language}" swaggertype:"string" example:"ja-JA"`
	}

	Country struct {
		ID   int16  `json:"id" db:"id,pk" fake:"{number:1,9001}"`
		Name string `json:"name" db:"name" fake:"{country}" swaggertype:"string" example:"Japan"`
		Code string `json:"code" db:"code" fake:"{countrycode}" swaggertype:"string" example:"JP"`
	}

	City struct {
		UUID    uuid.UUID `json:"uuid" db:"uuid,pk" fake:"{uuid}" swaggertype:"string" example:"12345678-90ab-cdef-9876-543210fedcba"`
		Name    string    `json:"name" db:"name" fake:"{city}" swaggertype:"string" example:"Tokyo"`
		Lat     float64   `json:"lat" db:"lat" fake:"{latitude}" swaggertype:"number" example:"35.6895" format:"float"`
		Lng     float64   `json:"lng" db:"lng" fake:"{longitude}" swaggertype:"number" example:"139.6917" format:"float"`
		Country *Country  `json:"country" db:"country" fake:"{country}" example:"Japan"`
	}

	Venue struct {
		UUID    uuid.UUID `json:"uuid" db:"uuid,pk" fake:"{uuid}" swaggertype:"string" example:"12345678-90ab-cdef-9876-543210fedcba"`
		Name    string    `json:"name" db:"name" fake:"{city}" swaggertype:"string" example:"Tokyo"`
		Active  bool      `json:"active" db:"active" fake:"{bool}" swaggertype:"boolean" example:"true"`
		Street  string    `json:"street" db:"street" fake:"{street}" swaggertype:"string" example:"1-1-1"`
		Zip     string    `json:"zip" db:"zip" fake:"{zip}" swaggertype:"string" example:"100-0001"`
		Unit    string    `json:"unit" db:"unit" fake:"{building}" swaggertype:"string" example:"Shinjuku Sumitomo Building"`
		City    *City     `json:"city" db:"city" fake:"{city}" example:"Tokyo"`
		Country *Country  `json:"country" db:"country" fake:"{country}" example:"Japan"`
	}
)
