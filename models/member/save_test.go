package member

import (
	"context"
	"fmt"
	"testing"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestData(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, `
	CREATE SEQUENCE IF NOT EXISTS members_id_seq START 1;

	CREATE TYPE public."role" AS ENUM (
		'admin',
		'member'
	);

	CREATE TABLE IF NOT EXISTS cdn.images (
		id bigserial NOT NULL,
		"source" varchar(255) NOT NULL,
		thumbnail varchar(255) NULL,
		alt varchar(255) NULL,
		uploader varchar NOT NULL,
		sha256sum bpchar(64) NULL,
		CONSTRAINT images_pkey PRIMARY KEY (id)
	);

	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS public.members (
		id_numeric int4 DEFAULT nextval('members_id_seq'::regclass) NOT NULL,
		id uuid DEFAULT uuid_generate_v4() NOT NULL,
		nick varchar(255) NOT NULL,
		email varchar(255) NOT NULL,
		passhash varchar(255) NOT NULL,
		reg_timestamp timestamp DEFAULT now() NOT NULL,
		profilepic_id int8 NULL,
		display_name varchar NULL,
		bio text NULL,
		active bool DEFAULT false NOT NULL,
		roles public."_role" NOT NULL,
		following_uri text DEFAULT ''::text NOT NULL,
		visibility text DEFAULT 'public'::text NOT NULL,
		session_timeout int8 NULL,
		public_key_pem text DEFAULT ''::text NOT NULL,
		webfinger varchar NOT NULL,
		custom_fields jsonb NULL,
		modified int8 NULL,
		added int8 NULL,
		doc jsonb NULL,
		doc_id text NULL,
		CONSTRAINT members_pkey PRIMARY KEY (id_numeric),
		CONSTRAINT members_un UNIQUE (nick, email),
		CONSTRAINT members_unique UNIQUE (id),
		CONSTRAINT members_unique_1 UNIQUE (webfinger)
	);
	CREATE INDEX members_uuid_idx ON public.members USING btree (id);

	CREATE TABLE IF NOT EXISTS member_prefs (
		member_id int4 PRIMARY KEY,
		CONSTRAINT member_prefs_fk FOREIGN KEY (member_id) 
		REFERENCES public.members(id_numeric) 
		ON DELETE CASCADE
	);
	`)

	if err != nil {
		return fmt.Errorf("failed to create test data: %w", err)
	}

	return nil
}

func cleanTestData(ctx context.Context, conn *pgxpool.Pool) error {

	_, err := conn.Exec(ctx, `DROP TABLE IF EXISTS members CASCADE;
	DROP TABLE IF EXISTS cdn.images CASCADE;
	DROP SEQUENCE IF EXISTS members_id_seq CASCADE;
	DROP TABLE IF EXISTS member_prefs CASCADE;
	DROP TYPE public."role" CASCADE;`)

	if err != nil {
		return fmt.Errorf("failed to clean test data: %w", err)
	}

	return nil
}

func createTestStorage(conn *pgxpool.Pool, conf *cfg.Config, vp *validator.Validate) *PgMemberStorage {
	log := zerolog.Nop()
	return NewSQLStorage(conn, &log, conf, vp)
}

func TestSave(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	conf := cfg.TestConfig

	conn, err := pgxpool.New(ctx, db.CreateDsn(&conf.DBConfig))
	require.NoErrorf(t, err, "failed to connect to database: %v", err)
	defer conn.Close()
	fmt.Printf("Connected to database %s\n", conf.DBConfig.Database)

	vp := validator.New()
	s := createTestStorage(conn, &conf, vp)
	require.NotNil(t, s)
	fmt.Println("Created storage")

	err = createTestData(ctx, conn)
	require.NoErrorf(t, err, "failed to create test data: %v", err)
	fmt.Println("Created test data")

	defer func() {
		err = cleanTestData(ctx, conn)
		require.NoErrorf(t, err, "failed to clean test data: %v. PLEASE CLEAN UP THE DATABASE MANUALLY IF RUNNING LOCALLY ON A PERSISTENT TEST DB!", err)
	}()

	testCases := []struct {
		name    string
		input   *Member
		wantErr bool
	}{{
		name: "valid member data",
		input: &Member{
			PassHash:     "123456",
			MemberName:   "testValid",
			Webfinger:    "test-valid@example.com",
			Email:        "test-valid@example.com",
			RegTimestamp: time.Now(),
			Active:       true,
			Roles:        []string{"member"},
		},
		wantErr: false,
	}, {
		name: "missing values for non-nullable fields",
		input: &Member{
			PassHash:   "1234567",
			MemberName: "test",
		},
		wantErr: true,
	},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("Running test case %q\n", tc.name)
			err := s.Save(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			}
			if !tc.wantErr {
				assert.NoError(t, err)
			}
		})
	}

}
