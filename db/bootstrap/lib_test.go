package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/require"
)

func TestCreateEnumtype(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(
		ctx, "postgres", "user=postgres dbname=librate_test sslmode=disable",
	)
	require.NoErrorf(t, err, "failed to connect to the test database: %v", err)

	defer db.Close()

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS test_schema;")
	require.NoErrorf(t, err, "failed to create test schema: %v", err)

	defer db.ExecContext(ctx, "DROP SCHEMA IF EXISTS test_schema CASCADE;")

	typeName := "test_enum"
	schema := "test_schema"
	values := []string{"foo", "bar"}

	err = createEnumType(ctx, db, typeName, schema, values...)
	require.NoErrorf(t, err, "failed to create enum: %v", err)
}

func TestFormatValues(t *testing.T) {
	mediaKinds := []string{"album", "track", "film"}
	fmted := formatValues(mediaKinds)

	require.Equal(t, fmted, "'album', 'track', 'film'")
}
