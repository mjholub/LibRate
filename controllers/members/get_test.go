package members

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/cmd"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/middleware/session"
	"codeberg.org/mjh/LibRate/models/member"
)

func PrepareTest(ctx context.Context, testDBConn *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		err := createTestSchema(ctx, testDBConn)
		if err != nil {
			return err
		}
		err = createTestRole(ctx, testDBConn)
		if err != nil {
			return err
		}
		tx, err := testDBConn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() error {
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil
		}()
		_, err = testDBConn.Exec(`
	CREATE TABLE public.members (
	id serial4 NOT NULL,
	"uuid" uuid NOT NULL,
	nick varchar(255) NOT NULL,
	email varchar(255) NOT NULL,
	passhash varchar(255) NOT NULL,
	reg_timestamp timestamp NOT NULL DEFAULT now(),
	profilepic_id int8 NULL,
	display_name varchar NULL,
	homepage varchar NULL,
	irc varchar NULL,
	xmpp varchar NULL,
	matrix varchar NULL,
	bio text NULL,
	active bool NOT NULL DEFAULT false,
	roles public."_role" NOT NULL,
	"following" jsonb NOT NULL DEFAULT '[]'::jsonb,
	visibility text NOT NULL DEFAULT 'private'::text,
	CONSTRAINT members_pkey PRIMARY KEY (id)
);`)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
}

func createTestSchema(ctx context.Context, conn *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		_, err = conn.Exec("CREATE SCHEMA IF NOT EXISTS public; SET search_path TO public;")
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
}

func createTestRole(ctx context.Context, conn *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		_, err = conn.Exec(`
	CREATE TYPE public."_role" AS ENUM (
	'member',
	'admin'
);`)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
}

func CleanupTest(testDBConn *sqlx.DB) error {
	_, err := testDBConn.Exec("DROP TABLE IF EXISTS members")
	if err != nil {
		return err
	}

	_, err = testDBConn.Exec("DROP TYPE IF EXISTS public.\"_role\"")
	if err != nil {
		return err
	}
	return nil
}

func TestGetMember(t *testing.T) {
	app := cmd.CreateApp(&cfg.TestConfig)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	sess, err := session.Setup(&cfg.TestConfig)
	require.NoError(t, err)
	app.Use(sess)
	middlewares := cmd.SetupMiddlewares(&cfg.TestConfig, &logger)
	for i := range middlewares {
		app.Use(middlewares[i])
	}
	logger.Debug().Msg("middlewares setup")
	nameUUID, err := uuid.NewV4()
	require.NoError(t, err)
	name := nameUUID.String()
	nameParts := strings.Split(name, "-")
	emailName := nameParts[0]
	emailDomain := nameParts[1]
	email := fmt.Sprintf("%s@%s.com", emailName, emailDomain)

	conn, err := db.Connect(&cfg.TestConfig)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()
	logger.Debug().Msgf("connected to database with DSN: %s", db.CreateDsn(&cfg.TestConfig.DBConfig))
	/*
		var mu sync.Mutex
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mu.Lock()
		err = PrepareTest(ctx, conn)
		logger.Debug().Msg("test schema prepared")
		mu.Unlock()
		require.NoError(t, err)
		defer require.NoError(t, CleanupTest(conn))
		// get all schemas and tables in test database
		var dbContents []struct {
			TableCatalog              string         `db:"table_catalog"`
			TableSchema               string         `db:"table_schema"`
			TableName                 string         `db:"table_name"`
			TableType                 string         `db:"table_type"`
			SelfReferencingColumnName sql.NullString `db:"self_referencing_column_name"`
			ReferenceGeneration       sql.NullString `db:"reference_generation"`
			UserDefinedTypeCatalog    sql.NullString `db:"user_defined_type_catalog"`
			UserDefinedTypeSchema     sql.NullString `db:"user_defined_type_schema"`
			UserDefinedTypeName       sql.NullString `db:"user_defined_type_name"`
			IsInsertableInto          string         `db:"is_insertable_into"`
			IsTyped                   string         `db:"is_typed"`
			CommitAction              sql.NullString `db:"commit_action"`
		}
		err = conn.SelectContext(context.Background(), &dbContents, "SELECT * FROM information_schema.tables WHERE table_catalog = 'librate_test' AND table_schema = 'public'")
		require.NoError(t, err)
		assert.NotNilf(t, dbContents, "test database is empty")
		log.Tracef("test database contents: %v", dbContents)
	*/
	testUser := &member.Member{
		UUID:         uuid.Must(uuid.NewV4()),
		PassHash:     "testhash",
		MemberName:   name,
		Email:        email,
		RegTimestamp: time.Now(),
		Roles:        []string{"member"},
	}

	storage := member.NewSQLStorage(conn, &logger, &cfg.TestConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = storage.Save(ctx, testUser)
	if err != nil && strings.Contains(err.Error(), "already taken") {
		var mu sync.Mutex
		mu.Lock()
		err = storage.Delete(ctx, testUser)
		require.NoErrorf(t, err, "failed to delete test user %s", name)
		mu.Unlock()
		err = storage.Save(ctx, testUser)
		require.NoErrorf(t, err, "failed to save test user %s", name)
	}
	require.NoErrorf(t, err, "failed to save test user %s", name)
	defer func() {
		err = storage.Delete(ctx, testUser)
		require.NoErrorf(t, err, "failed to delete test user %s", name)
	}()
	service := NewController(storage, &logger, &cfg.TestConfig)

	app.Get("/api/members/:email_or_username/info", service.GetMemberByNickOrEmail)

	host := fmt.Sprintf("http://%s:%d/api/members/%s/info", cfg.TestConfig.Host, cfg.TestConfig.Port, name)
	req := httptest.NewRequest("GET", host, nil)

	resp, err := app.Test(req)
	require.NoErrorf(t, err, "failed to make request to %s", host)
	require.Equalf(t, 200, resp.StatusCode, "unexpected status code %d", resp.StatusCode)
}
