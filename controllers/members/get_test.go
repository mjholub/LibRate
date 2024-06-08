package members

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func PrepareTest(ctx context.Context, testDBConn *pgxpool.Pool) error {
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
		tx, err := testDBConn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			return err
		}
		defer func() error {
			err = tx.Rollback(ctx)
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return nil
		}()
		_, err = testDBConn.Exec(ctx, `
	CREATE TABLE public.members (
	id serial4 NOT NULL,
	"uuid" uuid NOT NULL,
	nick varchar(255) NOT NULL,
	email varchar(255) NOT NULL,
	passhash varchar(255) NOT NULL,
	reg_timestamp timestamp NOT NULL DEFAULT now(),
	profilepic_id int8 NULL,
	display_name varchar NULL,
	bio text NULL,
	active bool NOT NULL DEFAULT false,
	roles public."_role" NOT NULL,
	"following" jsonb NOT NULL DEFAULT '[]'::jsonb,
	visibility text NOT NULL DEFAULT 'private'::text,
	custom_fields jsonb NULL,
	CONSTRAINT members_pkey PRIMARY KEY (id)
);`)
		if err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}
}

func createTestSchema(ctx context.Context, conn *pgxpool.Pool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		_, err = conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS public; SET search_path TO public;")
		if err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}
}

func createTestRole(ctx context.Context, conn *pgxpool.Pool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		_, err = conn.Exec(ctx, `
	CREATE TYPE public."_role" AS ENUM (
	'member',
	'admin'
);`)
		if err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}
}

func CleanupTest(testDBConn *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := testDBConn.Exec(ctx, "DROP TABLE IF EXISTS members")
	if err != nil {
		return err
	}

	_, err = testDBConn.Exec(ctx, "DROP TYPE IF EXISTS public.\"_role\"")
	if err != nil {
		return err
	}
	return nil
}

// FIXME: upstream pgxmock issue with interface incompqatibility with pgxpool and
// I'm reluctant to do E2E tests
/*
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

	conn, err := db.Connect(
		cfg.TestConfig.Engine,
		db.CreateDsn(&cfg.TestConfig.DBConfig),
		cfg.TestConfig.RetryAttempts,
	)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()
	logger.Debug().Msgf("connected to database with DSN: %s", db.CreateDsn(&cfg.TestConfig.DBConfig))
	testUser := &member.Member{
		UUID:         uuid.Must(uuid.NewV4()),
		PassHash:     "testhash",
		MemberName:   name,
		Email:        email,
		RegTimestamp: time.Now(),
		Roles:        []string{"member"},
	}

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := member.NewSQLStorage(conn, &mock, &logger, &cfg.TestConfig)
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

	service := NewController(storage, conn, sess, &logger, &cfg.TestConfig)

	app.Get("/api/members/:email_or_username/info", service.GetMemberByNickOrEmail)

	host := fmt.Sprintf("http://%s:%d/api/members/%s/info", cfg.TestConfig.Host, cfg.TestConfig.Port, name)
	req := httptest.NewRequest("GET", host, nil)

	resp, err := app.Test(req)
	require.NoErrorf(t, err, "failed to make request to %s", host)
	require.Equalf(t, 200, resp.StatusCode, "unexpected status code %d", resp.StatusCode)
}
*/

func TestJsonbToStringMap(t *testing.T) {
	testCases := []struct {
		name  string
		input pgtype.JSONB
		want  map[string]string
	}{
		{
			name: "empty jsonb",
			input: pgtype.JSONB{
				Bytes:  []byte{},
				Status: pgtype.Null,
			},
			want: nil,
		},
		{
			name: "jsonb with one key-value pair",
			input: pgtype.JSONB{
				Bytes:  []byte(`{"key": "value"}`),
				Status: pgtype.Present,
			},
			want: map[string]string{"key": "value"},
		},
		{
			name: "multi-key jsonb",
			input: pgtype.JSONB{
				Bytes:  []byte(`{"key1": "value1", "key2": "value2"}`),
				Status: pgtype.Present,
			},
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	for _, tc := range testCases {
		out, err := jsonbToStringMap(tc.input)
		assert.Nil(t, err)
		assert.Equalf(t, tc.want, out, "unexpected output for test case %s", tc.name)
	}
}

func BenchmarkJsonbToStringMap(b *testing.B) {
	input := pgtype.JSONB{
		Bytes: []byte(`
		{"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
		"key5": "value5",
		"key6": "value6",
		"key7": "value7",
		"key8": "value8"
	}`),
		Status: pgtype.Present,
	}
	for i := 0; i < b.N; i++ {
		_, err := jsonbToStringMap(input)
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
	}
}
