package media

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/brianvoe/gofakeit/v7"

	"codeberg.org/mjh/LibRate/cfg"
	dblib "codeberg.org/mjh/LibRate/db"

	"codeberg.org/mjh/LibRate/internal/testhelpers"
)

func createTestPTables(ctx context.Context, conn *pgxpool.Pool) error {
	if _, err := conn.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS people;`); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	if _, err := conn.Exec(ctx, `CREATE TYPE people."role" AS ENUM (
	'actor',
	'director',
	'producer',
	'writer',
	'composer',
	'artist',
	'author',
	'publisher',
	'editor',
	'photographer',
	'illustrator',
	'narrator',
	'performer',
	'host',
	'guest',
	'other');`); err != nil {
		return fmt.Errorf("failed to create role type: %w", err)
	}

	if _, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	_, err := conn.Exec(ctx, `CREATE TABLE people.person (
	id uuid DEFAULT uuid_generate_v4() NOT NULL,
	first_name varchar(255) NOT NULL,
	other_names _varchar NULL,
	last_name varchar(255) NOT NULL,
	nick_names _varchar NULL,
	roles people."_role" NULL,
	birth date NULL,
	death date NULL,
	website varchar(255) NULL,
	bio text NULL,
	modified int8 NULL,
	added int8 NULL,
	doc jsonb NULL,
	from_pg bool DEFAULT true NOT NULL,
	doc_id text NULL,
	CONSTRAINT person_pkey PRIMARY KEY (id)
);`)
	if err != nil {
		return err
	}
	return nil
}

func dropTestPTables(ctx context.Context, conn *pgxpool.Pool) error {
	if _, err := conn.Exec(ctx, `DROP TABLE IF EXISTS people.person;`); err != nil {
		return fmt.Errorf("failed to drop person table: %w", err)
	}
	if _, err := conn.Exec(ctx, `DROP TYPE people."role";`); err != nil {
		return fmt.Errorf("failed to drop role type: %w", err)
	}

	return nil
}

func TestCreatePerson(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dblib.CreateDsn(&cfg.TestConfig.DBConfig))
	require.NoErrorf(t, err, "failed to create db connection: %v", err)
	require.NotNilf(t, db, "db connection is nil")

	err = createTestPTables(ctx, db)
	require.NoErrorf(t, err, "failed to create test tables: %v", err)
	defer func(ctx context.Context, db *pgxpool.Pool) {
		err := dropTestPTables(ctx, db)
		require.NoErrorf(t, err, "failed to drop test tables: %v", err)
	}(ctx, db)

	l := zerolog.Nop()

	ps := NewPeopleStorage(db, &l)

	testCases := []testhelpers.TestCase[*Person, map[string]bool]{{
		Name: "Full struct",
		Input: func() *Person {
			var p Person
			err := gofakeit.Struct(&p)
			require.NoErrorf(t, err, "error generating test data: %v", err)

			return &p
		},
		Output: map[string]bool{"uuid valid?": true, "has errors?": false}},
		{
			Name: "Empty struct",
			Input: func() *Person {
				return &Person{}
			},
			Output: map[string]bool{"uuid valid?": false, "has errors?": true}},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := tc.Input()

			id, err := ps.CreatePerson(ctx, p)
			if tc.Output["has errors?"] {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.Output["uuid valid?"] {
				require.Truef(t, id != nil, "id is nil")
			} else {
				require.Truef(t, id == nil, "id is not nil")
			}
		})
	}
}
