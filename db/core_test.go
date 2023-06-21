package db_test

import (
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
)

type TestCase struct {
	Name   string
	Inputs interface{}
	Want   []interface{}
}

func TestCreateDsn(t *testing.T) {
	testCases := []TestCase{
		{
			Name: "PostgresNoSSL",
			Inputs: cfg.DBConfig{
				Engine:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "librate_test",
				User:     "postgres",
				Password: "postgres",
				SSL:      "disable",
			},
			Want: []interface{}{("postgres://postgres:postgres@localhost:5432/librate_test?sslmode=disable")},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got := db.CreateDsn(tc.Inputs.(*cfg.DBConfig))
			assert.Equal(t, got, tc.Want)
		})
	}
}

func TestConnect(t *testing.T) {
	testCases := []TestCase{
		{
			Name: "HappyPath",
			Inputs: cfg.DBConfig{
				Engine:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "librate_test",
				User:     "postgres",
				Password: "postgres",
				SSL:      "disable",
			},
			Want: []interface{}{(&sqlx.DB{}), nil},
		},
		{
			Name: "BadEngine",
			Inputs: cfg.DBConfig{
				Engine:   "badengine",
				Host:     "localhost",
				Port:     5432,
				Database: "librate_test",
				User:     "postgres",
				Password: "postgres",
				SSL:      "disable",
			},
			Want: []interface{}{nil, errors.New("sql: unknown driver \"badengine\" (forgotten import?)")},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := db.Connect(tc.Inputs.(*cfg.Config))
			assert.Equal(t, got, tc.Want[0])
			assert.Equal(t, err, tc.Want[1])
		})
	}
}
