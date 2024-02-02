package member

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
)

func TestBan(t *testing.T) {
	logger := zerolog.Nop()
	// create connection to the test database

	conf := cfg.TestConfig
	dsn := db.CreateDsn(&conf.DBConfig)
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoErrorf(t, err, "failed to create connection pool: %v", err)

	defer pool.Close()

	storage := NewSQLStorage(nil, pool, &logger, &cfg.TestConfig)
	ends := time.Now().Add(24 * time.Hour)
	_, mask, err := net.ParseCIDR("192.0.2.0/24")
	require.NoError(t, err)

	uuid, err := uuid.NewV4()
	require.NoErrorf(t, err, "failed to generate uuid: %v", err)

	bi := &BanInput{
		Reason:    "spam",
		Ends:      ends,
		CanAppeal: false,
		Mask:      mask,
	}

	// create the bans table using the ddl from the original DB
	membersQuery := `CREATE TABLE public.members (
		uuid uuid NOT NULL,
		roles ENUM('user', 'moderator', 'admin') NOT NULL DEFAULT 'user',
	);`

	query := `CREATE TABLE public.bans (
	member_uuid uuid NOT NULL,
	reason text NOT NULL,
	ends timetz NOT NULL,
	can_appeal bool NOT NULL DEFAULT true,
	mask inet NULL,
	occurrence int2 NOT NULL DEFAULT 1,
	started timetz NOT NULL DEFAULT now(),
	banned_by uuid NOT NULL REFERENCES public.members(uuid) ON DELETE SET NULL,
	CONSTRAINT bans_pk PRIMARY KEY (member_uuid)
);


-- public.bans foreign keys

ALTER TABLE public.bans ADD CONSTRAINT bans_member_uuid_fkey FOREIGN KEY (member_uuid) REFERENCES public.members("uuid") ON DELETE CASCADE;`

	// create sample members: banner and bannee
	creationQuery := `INSERT INTO public.members (uuid, roles) VALUES ($1, 'admin'), ($2, 'user');`

	// create a sample ban

	banQuery := `INSERT INTO public.bans (member_uuid, reason, ends, can_appeal, mask) VALUES ($1, $2, $3, $4, $5);`

	// execute the queries sequentially

	defer CleanupTestDB(t, pool)

	_, err = pool.Exec(context.Background(), membersQuery)
	require.NoErrorf(t, err, "failed to create members table: %v", err)

	_, err = pool.Exec(context.Background(), creationQuery, uuid, uuid)
	require.NoErrorf(t, err, "failed to create members: %v", err)

	_, err = pool.Exec(context.Background(), query)
	require.NoErrorf(t, err, "failed to create bans table: %v", err)

	_, err = pool.Exec(context.Background(), banQuery, uuid, bi.Reason, bi.Ends, bi.CanAppeal, bi.Mask)

	require.NoErrorf(t, err, "failed to create ban: %v", err)

	user := &Member{
		UUID: uuid,
	}

	err = storage.Ban(context.Background(), user, bi)
	assert.NoErrorf(t, err, "failed to ban user: %v", err)
}

func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "DROP TABLE public.bans;")
	require.NoErrorf(t, err, "failed to drop bans table: %v", err)
	_, err = pool.Exec(context.Background(), "DROP TABLE public.members;")
	require.NoErrorf(t, err, "failed to drop members table: %v", err)
}

func TestRead(t *testing.T) {
	logger := zerolog.Nop()
	// create connection to the test database

	conf := cfg.TestConfig
	dsn := db.CreateDsn(&conf.DBConfig)
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoErrorf(t, err, "failed to create connection pool: %v", err)

	defer pool.Close()

	membersTableQuery := `CREATE TABLE public.members (
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
	following_uri text NOT NULL DEFAULT ''::text,
	visibility text NOT NULL DEFAULT 'public'::text,
	session_timeout int8 NULL,
	public_key_pem text NOT NULL DEFAULT ''::text,
	webfinger varchar NOT NULL,
	custom_fields jsonb NULL,
	CONSTRAINT members_pkey PRIMARY KEY (id),
	CONSTRAINT members_un UNIQUE (nick, email) ,
	CONSTRAINT members_unique UNIQUE (uuid),
	CONSTRAINT members_unique_1 UNIQUE (webfinger)
);`

	memberInsert := `INSERT INTO public.members
	(uuid, nick, email, passhash, reg_timestamp, roles, webfinger)
	VALUES (
		'f47ac10b-58cc-4372-a567-0e02b2c3d479', 
		'test',
		'test@test.com',
		'password',
		'2024-01-01 00:00:00',
		'user',
		'test@localhost'
	);`

	defer CleanupTestDB(t, pool)

	_, err = pool.Exec(context.Background(), membersTableQuery)
	require.NoErrorf(t, err, "failed to create members table: %v", err)

	_, err = pool.Exec(context.Background(), memberInsert)

	require.NoErrorf(t, err, "failed to create member: %v", err)

	storage := NewSQLStorage(nil, pool, &logger, &cfg.TestConfig)

	var member *Member

	member, error := storage.Read(context.Background(), "test", "nick")

	require.NoErrorf(t, error, "failed to read member: %v", error)

	require.NotEmpty(t, member, "member is empty")
}
