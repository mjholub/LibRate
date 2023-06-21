package db_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// TestInitDB bootstraps, then cleans up on the test database
func TestInitDB(t *testing.T) {
	os.Setenv("LIBRATE_ENV", "test")
	cfg, err := cfg.LoadConfig().Get()
	assert.NoError(t, err)
	require.Equal(t, cfg.Database, "librate_test")
	err = db.InitDB()
	require.NoError(t, err)
	db := cfg.Database
	cleanTables, err := sqlx.Open("postgres", db)
	assert.NoError(t, err)
	defer cleanTables.Close()
	_, err = cleanTables.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	assert.NoError(t, err)
}
