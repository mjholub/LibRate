package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
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
					Engine:        "postgres",
					Host:          "localhost",
					Port:          5432,
					Database:      "librate_test",
					User:          "postgres",
					Password:      "postgres",
					SSL:           "disable",
					RetryAttempts: 2,
				},
			},
			WantErr: false,
		},
		{
			Name: "BadEngine",
			Inputs: &cfg.Config{
				DBConfig: cfg.DBConfig{
					Engine:        "badengine",
					Host:          "localhost",
					Port:          5432,
					Database:      "librate_test",
					User:          "postgres",
					Password:      "postgres",
					SSL:           "disable",
					RetryAttempts: 1,
				},
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			dsn := CreateDsn(&tc.Inputs.DBConfig)
			ctx, cancel := context.WithTimeout(context.Background(), 5)
			defer cancel()
			got, err := Connect(ctx, dsn, tc.Inputs.RetryAttempts)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			assert.IsType(t, &pgxpool.Pool{}, got)
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
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	err := InitDB(&config.DBConfig, &log)
	require.NoError(t, err)
}

func TestCreateExtension(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	conn, err := pgxpool.New(ctx, CreateDsn(&cfg.TestConfig.DBConfig))
	require.NotNil(t, conn)
	require.NoError(t, err)
	err = createExtension(conn, "sequential_uuids")
	assert.NoError(t, err)
}

func createTestData(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS members (
			id SERIAL PRIMARY KEY,
			email text NOT NULL
		);`)
	if err != nil {
		return fmt.Errorf("failed to create test table: %w", err)
	}

	_, err = conn.Exec(ctx, `
	INSERT INTO members (email) VALUES ('test@foo.com');`)

	if err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}

	return nil
}

func cleanTestData(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, "DROP TABLE IF EXISTS members")
	if err != nil {
		return fmt.Errorf("failed to drop test table: %w", err)
	}
	return nil
}

func TestSerializableParametrizedTx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, CreateDsn(&cfg.TestConfig.DBConfig))
	require.NotNil(t, conn)
	require.NoError(t, err)

	defer func(ctx context.Context, conn *pgxpool.Pool) {
		if err := cleanTestData(ctx, conn); err != nil {
			t.Fatalf("failed to clean test data: %v", err)
		}
	}(ctx, conn)

	if err = createTestData(ctx, conn); err != nil {
		t.Fatalf("failed to create test data: %v", err)
	}

	var result []int
	result, err = SerializableParametrizedTx[int](ctx, conn,
		"test-query", // query name
		"SELECT id FROM members WHERE email = $1", // sql query
		"test",         // error handler input
		"test@foo.com", // parameter
	)
	require.Nilf(t, err, "failed to run serializable transaction: %v", err)
	assert.IsType(t, []int{0}, result)

	assert.NotEmpty(t, result)
	assert.Greater(t, result[0], 0)
}
