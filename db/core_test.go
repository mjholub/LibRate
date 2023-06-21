package db_test

import (
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
			Inputs: &cfg.DBConfig{
				//				DBConfig: cfg.DBConfig{
				Engine:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "librate_test",
				User:     "postgres",
				Password: "postgres",
				SSL:      "disable",
			},
			//			},
			Want: []interface{}{("postgres://postgres:postgres@localhost:5432/librate_test?sslmode=disable")},
		},
	}
	for i, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got := db.CreateDsn(tc.Inputs.(*cfg.DBConfig))
			assert.Equal(t, tc.Want[i], got)
		})
	}
}

func TestConnect(t *testing.T) {
	testCases := []struct {
		Name    string
		Inputs  *cfg.Config
		WantErr bool
	}{
		{
			Name: "HappyPath",
			Inputs: &cfg.Config{
				DBConfig: cfg.DBConfig{
					Engine:   "postgres",
					Host:     "localhost",
					Port:     5432,
					Database: "librate_test",
					User:     "postgres",
					Password: "postgres",
					SSL:      "disable",
				},
			},
			WantErr: false,
		},
		{
			Name: "BadEngine",
			Inputs: &cfg.Config{
				DBConfig: cfg.DBConfig{
					Engine:   "badengine",
					Host:     "localhost",
					Port:     5432,
					Database: "librate_test",
					User:     "postgres",
					Password: "postgres",
					SSL:      "disable",
				},
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := db.Connect(tc.Inputs)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			assert.IsType(t, &sqlx.DB{}, got)
			assert.NoError(t, err)
		})
	}
}
