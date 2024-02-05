package cmd

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateCache(t *testing.T) {
	dsn := db.CreateDsn(&cfg.TestConfig.DBConfig)
	dbConn, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	defer dbConn.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.TestConfig.Redis.Host + ":" + strconv.Itoa(cfg.TestConfig.Redis.Port),
		Password: cfg.TestConfig.Redis.Password,
		DB:       cfg.TestConfig.Redis.CacheDB,
	})

	defer func() {
		_, err = dbConn.Exec(context.Background(), `DROP TABLE public.members`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE cdn.images`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE media.media_images`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE media.media`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE people.person`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE people.group`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
		_, err = dbConn.Exec(context.Background(), `DROP TABLE people.studio`)
		require.NoErrorf(t, err, "error dropping table: %v", err)
	}()

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS
	public.members (webfinger text PRIMARY KEY)`)
	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE SCHEMA IF NOT EXISTS 
		media`)
	assert.NoErrorf(t, err, "error creating schema: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE SCHEMA IF NOT EXISTS
	cdn`)

	assert.NoErrorf(t, err, "error creating schema: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE SCHEMA IF NOT EXISTS
	people`)
	assert.NoErrorf(t, err, "error creating schema: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE media.media
	(id uuid NOT NULL PRIMARY KEY, title text NOT NULL, kind text NOT NULL)`)
	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE media.media_images
	(media_id uuid NOT NULL REFERENCES media.media(id) ON DELETE CASCADE, 
	image_id bigserial NOT NULL PRIMARY KEY, is_main boolean NOT NULL)`)
	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE cdn.images
	(source text NOT NULL, id bigint NOT NULL
	REFERENCES media.media_images(image_id) ON DELETE CASCADE)`)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE people.person
	(first_name text NOT NULL PRIMARY KEY, last_name text NOT NULL DEFAULT '',
	roles text[] NOT NULL DEFAULT ARRAY['musician'], nick_names text[] NULL)`)
	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE people.group
	(name text NOT NULL PRIMARY KEY, kind text NOT NULL DEFAULT 'band')`)
	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `CREATE TABLE people.studio
	(name text NOT NULL PRIMARY KEY, kind text NOT NULL DEFAULT 'music')`)

	assert.NoErrorf(t, err, "error creating table: %v", err)

	_, err = dbConn.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION modified() RETURNS TRIGGER AS $$
BEGIN
  NEW.modified = EXTRACT(EPOCH FROM now()) * 1000::bigint;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add new columns and triggers for each table
DO $$
DECLARE
  _tbl text;
  _schema text;
	_trigger_name text;
BEGIN
  FOR _schema, _tbl IN (SELECT table_schema, table_name 
    FROM information_schema.tables WHERE table_schema IN
    ('public', 'media', 'people') AND table_name IN ('members', 'media', 'person', 'group', 'studio'))
  LOOP
		_trigger_name := _tbl || '_modified_trigger';
    EXECUTE format('ALTER TABLE %I.%I ADD COLUMN IF NOT EXISTS modified bigint', _schema, _tbl);
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I.%I', _trigger_name, _schema, _tbl);
    EXECUTE format('CREATE TRIGGER %I BEFORE INSERT OR UPDATE ON %I.%I FOR EACH ROW EXECUTE FUNCTION modified()', _trigger_name, _schema, _tbl);
  END LOOP;
END
$$;
	`)
	require.NoErrorf(t, err, "error adding update data to test tables: %v", err)

	_, err = dbConn.Exec(context.Background(), `
	INSERT INTO public.members (webfinger) VALUES ('test1@test.com');
	INSERT INTO media.media (id, title, kind) VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Greatest hits', 'album');
	INSERT INTO media.media_images (media_id, is_main) VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', true);
	INSERT INTO people.person (first_name, last_name) VALUES ('test', 'person');
	INSERT INTO people.group (name) VALUES ('The Trashmen');
	INSERT INTO people.studio (name) VALUES ('Giebli Studio');
`)
	require.NoErrorf(t, err, "error adding data to test tables: %v", err)

	_, err = dbConn.Exec(context.Background(), `
	INSERT INTO cdn.images (source, id) VALUES ('http://example.com/image.jpg', 1);
	`)

	// if logger writes anything to stdout, the test will fail
	buf := new(bytes.Buffer)

	log := zerolog.New(os.Stdout).With().Timestamp().Logger().Output(buf)

	err = PopulateCache(redisClient, dbConn, &log, &cfg.TestConfig, true)

	assert.NoErrorf(t, err, "error populating cache: %v", err)
	assert.Emptyf(t, buf.String(), "unexpected log output: %s", buf.String())
}
