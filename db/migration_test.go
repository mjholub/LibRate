package db

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/member"
)

type testCase struct {
	name   string
	inputs interface{}
}

// TestRunMigrations tests the runMigrations function
func TestRunMigrations(t *testing.T) {
	path := flag.String("path", "migrations", "Path to migrations")
	flag.Parse()
	testCases := []testCase{
		{
			name: "HappyPath", // DB initialized, all migrations run
			inputs: func(t *testing.T) {
				config := cfg.TestConfig

				defer func(config *cfg.Config) {
					if os.Getenv("CLEANUP_TEST_DB") == "0" || os.Getenv("CLEANUP_HAPPY_PATH") == "0" {
						return
					}
					fmt.Println("Cleaning up")
					err := DBTearDown(config)
					require.NoError(t, err)
				}(&config)
				err := InitDB(&config, true, false)
				require.NoError(t, err)
				err = Migrate(&config, *path)
				require.NoError(t, err)
				cwd, err := os.Getwd()
				require.NoErrorf(t, err, "Error getting current path: %v", err)
				require.NoErrorf(t, err, "Error running migrations: %v. Current path is: %s", err, cwd)
			},
		},
		{
			name: "MigrationsOnEmpty", // Test if migrations will run even if the db is empty
			inputs: func(t *testing.T) {
				config := cfg.TestConfig
				// only clean the test db after all tests have run
				err := Migrate(&config, *path)
				require.Errorf(t, err, "Error running migrations: %v", err)
			},
		},
		{
			name: "FailOnMissingFile", // must return an error if auto-migrate flag is not set
			inputs: func(t *testing.T) {
				config := cfg.TestConfig
				config.AutoMigrate = false
				err := Migrate(&config, "aaaaa.sql")
				require.Error(t, err)
			},
		},
		{
			name: "ApplySingleMigration",
			inputs: func(t *testing.T) {
				config := cfg.TestConfig
				// only clean the test db after all tests have run
				var mu sync.Mutex
				defer func(config *cfg.Config) {
					if os.Getenv("CLEANUP_TEST_DB") == "0" || os.Getenv("CLEANUP_SINGLE") == "0" {
						return
					}
					fmt.Println("Cleaning up")
					err := DBTearDown(config)
					require.NoError(t, err)
				}(&config)
				mu.Lock()
				err := InitDB(&config, true, false)
				require.NoError(t, err)
				mu.Unlock()
				err = flag.Set("path", "migrations/000001-fix-missing-timestamps/reviews.sql")
				defer flag.Set("path", "migrations")
				require.NoError(t, err)

				cwd, err := os.Getwd()
				require.NoErrorf(t, err, "Error getting current path: %v", err)
				err = Migrate(&config, *path)
				require.NoErrorf(t, err, "Error running migrations: %v. Current path is: %s", err, cwd)
				conn, err := Connect(&config, true)
				require.NoErrorf(t, err, "Error connecting to database: %v", err)
				defer conn.Close()
				log := zerolog.New(os.Stdout).With().Timestamp().Logger()

				// create a test member so that fkey constraints are satisfied
				ms := member.NewSQLStorage(conn, &log, &config)
				member := member.Member{
					UUID:         uuid.Must(uuid.NewV4()).String(),
					MemberName:   "test",
					Email:        "test@test.com",
					PassHash:     "test",
					RegTimestamp: time.Now(),
				}
				err = ms.Save(context.Background(), &member)
				require.NoErrorf(t, err, "Error saving test member: %v", err)
				id, err := ms.GetID(context.Background(), member.Email)
				require.NoErrorf(t, err, "Error getting test member ID: %v", err)
				result, err := conn.Exec("INSERT INTO reviews.ratings (stars, id, user_id, created_at) VALUES (1, 1, $1, NOW())", id)
				require.NoErrorf(t, err, "Error inserting test rating: %v", err)
				require.NotNil(t, result)
			},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			testCases[i].inputs.(func(t *testing.T))(t)
		})
	}
}
