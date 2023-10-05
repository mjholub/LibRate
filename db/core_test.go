package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
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
			got := CreateDsn(tc.Inputs.(*cfg.DBConfig))
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
			got, err := Connect(tc.Inputs, true)
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
	config := cfg.TestConfig

	require.Equal(t, config.Database, "librate_test")

	defer func(config *cfg.Config) {
		err := DBTearDown(config)
		require.NoError(t, err)
	}(&config)
	err := InitDB(&config, true, false)
	require.NoError(t, err)
}

// not passing *testing.T as parameter to avoid this helper function being treated as a test
func DBTearDown(conf *cfg.Config) error {
	if os.Getenv("CLEANUP_TEST_DB") == "0" {
		return nil
	}
	database := conf.DBConfig
	dsn := CreateDsn(&database)
	var cleanTables *sqlx.DB
	var err error
	cleanTables, err = sqlx.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer cleanTables.Close()
	_, err = cleanTables.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS cdn CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS places CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS media CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS people CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS reviews CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP SCHEMA IF EXISTS members CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	// delete the extensions
	_, err = cleanTables.Exec("DROP EXTENSION IF EXISTS pgcrypto CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP EXTENSION IF EXISTS \"uuid-ossp\" CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP EXTENSION IF EXISTS pg_trgm CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP EXTENSION IF EXISTS sequential_uuids CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	// cleanup custom types
	_, err = cleanTables.Exec("DROP TYPE IF EXISTS places.place_kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP TYPE IF EXISTS media.kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP TYPE IF EXISTS people.role CASCADE;")
	if err != nil {
		return teardownErr(err)
	}
	_, err = cleanTables.Exec("DROP TYPE IF EXISTS people.group_kind CASCADE;")
	if err != nil {
		return teardownErr(err)
	}

	return nil
}

func teardownErr(err error) error {
	return fmt.Errorf("failed to clean up database: %w", err)
}
