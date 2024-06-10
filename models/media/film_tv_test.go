package media

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFilmTables(ctx context.Context, conn *pgxpool.Pool) error {
	if _, err := conn.Exec(ctx, `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE media.media (
		id uuid DEFAULT uuid_generate_v4() NOT NULL,
	);`); err != nil {
		return fmt.Errorf("error creating media.media table: %w", err)
	}

	_, err := conn.Exec(ctx, `
		CREATE TABLE media.films (
	media_id uuid DEFAULT uuid_generate_v4() NOT NULL,
	title varchar(255) NOT NULL,
	duration time NULL,
	release_date date NULL,
	synopsis text NULL,
	CONSTRAINT films_pkey PRIMARY KEY (media_id)
);
-- media.films foreign keys
ALTER TABLE media.films ADD CONSTRAINT films_media_id_fkey 
FOREIGN KEY (media_id) REFERENCES media.media(id) ON DELETE CASCADE;
`)
	if err != nil {
		return fmt.Errorf("error creating media.films table: %w", err)
	}

	return nil
}

func cleanTestFilmTables(ctx context.Context, conn *pgxpool.Pool) error {
	if _, err := conn.Exec(ctx, `DROP TABLE media.films`); err != nil {
		return fmt.Errorf("error dropping media.films table: %w", err)
	}
	if _, err := conn.Exec(ctx, `DROP TABLE media.media`); err != nil {
		return fmt.Errorf("error dropping media.media table: %w", err)
	}

	return nil
}

// since AddFilm is coupled to Cast this test also covers AddCreators
func TestAddFilm(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, db.CreateDsn(&cfg.TestConfig.DBConfig))
	require.NoError(t, err)
	require.NotNil(t, conn)
	log := zerolog.Nop()
	ms := NewStorage(conn, &log)

	defer func(ctx context.Context, conn *pgxpool.Pool) {
		err := cleanTestFilmTables(ctx, conn)
		require.NoErrorf(t, err, "failed to clean test tables: %v", err)
	}(ctx, conn)

	err = createTestFilmTables(ctx, conn)
	require.NoErrorf(t, err, "failed to create test tables: %v", err)

	testCases := []struct {
		name      string
		film      *Film
		wantError bool
	}{
		{
			name: "valid film",
			film: &Film{
				Title: "The Matrix",
				ReleaseDate: sql.NullTime{
					Time:  time.Date(1999, time.March, 31, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				Duration: sql.NullTime{
					Time:  time.Date(1970, time.January, 1, 2, 16, 0, 0, time.UTC),
					Valid: true,
				},
				Synopsis: sql.NullString{
					String: `A computer hacker learns from mysterious rebels about the true
				nature of his reality and his role in the war against its controllers.`,
					Valid: true,
				},
				// TODO: add cast via st.ps.CreatePerson
			},

			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ms.AddFilm(ctx, tc.film)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestAddCast(t *testing.T) {

}
