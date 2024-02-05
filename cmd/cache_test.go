package cmd

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	webfinger       string
	mediaID         uuid.UUID
	mediaTitle      string
	mediaKind       string
	personFirstName string
	personLastName  string
	groupName       string
	studioName      string
}

func TestPopulateCache(t *testing.T) {
	dsn := db.CreateDsn(&cfg.TestConfig.DBConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbConn, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer dbConn.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.TestConfig.Redis.Host + ":" + strconv.Itoa(cfg.TestConfig.Redis.Port),
		Password: cfg.TestConfig.Redis.Password,
		DB:       cfg.TestConfig.Redis.CacheDB,
	})

	defer redisClient.Close()

	err = createTestTables(ctx, dbConn)
	assert.NoErrorf(t, err, "error creating test tables: %v", err)
	defer func() {
		err = cleanupTestDB(ctx, dbConn)
		assert.NoErrorf(t, err, "error cleaning up test DB: %v", err)
	}()
	err = createTestTrigger(ctx, dbConn)
	require.NoErrorf(t, err, "error creating test trigger: %v", err)

	data := testData{
		webfinger:  "test1@test.com",
		mediaID:    uuid.FromStringOrNil("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
		mediaTitle: "Greatest hits",
		mediaKind:  "album",
		groupName:  "The Trashmen",
		studioName: "Giebli Studio",
	}

	err = insertTestData(ctx, dbConn, &data)
	require.NoErrorf(t, err, "error adding data to test tables: %v", err)

	_, err = dbConn.Exec(context.Background(), `
	INSERT INTO cdn.images (source, id) VALUES ('http://example.com/image.jpg', 1);
	`)

	require.NoErrorf(t, err, "error adding data to test tables: %v", err)

	// if logger writes anything to stdout, the test will fail
	buf := new(bytes.Buffer)

	log := zerolog.New(os.Stdout).With().Timestamp().Logger().Output(buf)

	err = PopulateCache(ctx, redisClient, dsn, &log, &cfg.TestConfig, true)

	assert.NoErrorf(t, err, "error populating cache: %v", err)
	assert.Emptyf(t, buf.String(), "unexpected log output: %s", buf.String())
}

func insertTestData(ctx context.Context, dbConn *pgxpool.Pool, data *testData) error {
	_, err := dbConn.Exec(ctx, `
	INSERT INTO public.members (webfinger) VALUES ($1)`,
		data.webfinger)
	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}

	_, err = dbConn.Exec(ctx, `INSERT INTO media.media (id, title, kind) VALUES ($1, $2, $3)`,
		data.mediaID, data.mediaTitle, data.mediaKind)
	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}

	_, err = dbConn.Exec(ctx, `INSERT INTO media.media_images (media_id, is_main) VALUES ($1, true)`,
		data.mediaID)

	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}
	_, err = dbConn.Exec(ctx, `INSERT INTO people.person (first_name, last_name) VALUES ($1, $2)`,
		data.personFirstName, data.personLastName)

	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}
	_, err = dbConn.Exec(ctx, `INSERT INTO people.group (name) VALUES ($1)`,
		data.groupName)

	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}
	_, err = dbConn.Exec(ctx, `INSERT INTO people.studio (name) VALUES ($1)`,
		data.studioName)

	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}

	// get the latest image id and write that to the cdn.images table
	var imageID int
	err = dbConn.QueryRow(ctx, `SELECT image_id FROM media.media_images WHERE media_id = $1`, data.mediaID).Scan(&imageID)
	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}

	_, err = dbConn.Exec(ctx, `INSERT INTO cdn.images (source, id) VALUES ($1, $2)`,
		"http://example.com/image.jpg", imageID)

	if err != nil {
		return fmt.Errorf("error adding data to test tables: %w", err)
	}

	return nil
}

func cleanupTestDB(ctx context.Context, dbConn *pgxpool.Pool) error {
	var err error
	_, err = dbConn.Exec(ctx, `DROP TABLE public.members`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE cdn.images`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE media.media_images`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE media.media`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE people.person`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE people.group`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}
	_, err = dbConn.Exec(ctx, `DROP TABLE people.studio`)
	if err != nil {
		return fmt.Errorf("error cleaning up test DB: %w", err)
	}

	return nil
}

func createTestTables(ctx context.Context, dbConn *pgxpool.Pool) error {
	var err error
	_, err = dbConn.Exec(ctx, `CREATE TABLE IF NOT EXISTS
	public.members (webfinger text PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS 
		media`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS
	cdn`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS
	people`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE media.media
	(id uuid NOT NULL PRIMARY KEY, title text NOT NULL, kind text NOT NULL)`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE media.media_images
	(media_id uuid NOT NULL REFERENCES media.media(id) ON DELETE CASCADE, 
	image_id bigserial NOT NULL PRIMARY KEY, is_main boolean NOT NULL)`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE INDEX media_images_media_id_idx ON media.media_images (media_id)`)

	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE cdn.images
	(source text NOT NULL, id bigint NOT NULL
	REFERENCES media.media_images(image_id) ON DELETE CASCADE)`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE people.person
	(first_name text NOT NULL PRIMARY KEY, last_name text NOT NULL DEFAULT '',
	roles text[] NOT NULL DEFAULT ARRAY['musician'], nick_names text[] NULL)`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE people.group
	(name text NOT NULL PRIMARY KEY, kind text NOT NULL DEFAULT 'band')`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	_, err = dbConn.Exec(ctx, `CREATE TABLE people.studio
	(name text NOT NULL PRIMARY KEY, kind text NOT NULL DEFAULT 'music')`)
	if err != nil {
		return fmt.Errorf("error creating table or schema: %w", err)
	}

	return nil
}

func createTestTrigger(ctx context.Context, dbConn *pgxpool.Pool) error {
	_, err := dbConn.Exec(ctx, `
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
	if err != nil {
		return err
	}
	return nil
}

func BenchmarkPopulateCache(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := db.CreateDsn(&cfg.TestConfig.DBConfig)
	dbConn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		b.Fatalf("error establishing DB connection: %v", err)
	}

	log := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	log.Debug().Msg("DB connection established")

	defer dbConn.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.TestConfig.Redis.Host + ":" + strconv.Itoa(cfg.TestConfig.Redis.Port),
		Password: cfg.TestConfig.Redis.Password,
		DB:       cfg.TestConfig.Redis.CacheDB,
	})

	log.Debug().Msg("Redis connection established")

	defer redisClient.Close()

	err = createTestTables(ctx, dbConn)
	if err != nil {
		b.Fatalf("error creating test tables: %v", err)
	}
	defer func() {
		err = cleanupTestDB(ctx, dbConn)
		if err != nil {
			b.Fatalf("error cleaning up test DB: %v", err)
		}
	}()

	err = createTestTrigger(ctx, dbConn)
	require.NoError(b, err)

	log.Debug().Msg("Test tables created")

	var uuids []uuid.UUID

	for i := 0; i < 100; i++ {
		uuid, err := uuid.NewV4()
		if err != nil {
			b.Fatalf("error creating UUID: %v", err)
		}

		uuids = append(uuids, uuid)
	}
	words := []string{
		"a",
		"action",
		"again",
		"all",
		"among",
		"appear",
		"artist",
		"authority",
		"bar",
		"behavior",
		"billion",
		"box",
		"buy",
		"card",
		"century",
		"child",
		"clearly",
		"community",
		"contain",
		"create",
		"daughter",
		"defense",
		"development",
		"discussion",
		"drop",
		"edge",
		"energy",
		"evening",
		"executive",
		"fail",
		"few",
		"fine",
		"focus",
		"forward",
		"garden",
		"good",
		"guy",
		"head",
		"high",
		"hot",
		"I",
		"include",
		"institution",
		"item",
		"kind",
		"laugh",
		"leg",
		"likely",
		"loss",
		"make",
		"may",
		"memory",
		"miss",
		"mother",
		"my",
		"need",
		"none",
		"occur",
		"ok",
		"option",
		"own",
		"particularly",
		"per",
		"picture",
		"policy",
		"practice",
		"problem",
		"protect",
		"quickly",
		"ready",
		"record",
		"remove",
		"rest",
		"role",
		"science",
		"seek",
		"set",
		"shot",
		"since",
		"skin",
		"something",
		"space",
		"stand",
		"stock",
		"stuff",
		"support",
		"teacher",
		"thank",
		"they",
		"three",
		"too",
		"treat",
		"TV",
		"use",
		"vote",
		"weapon",
		"where",
		"wide",
		"woman",
		"wrong",
	}

	var webfingers, groupNames, studioNames []string

	for i := range words {
		webfingers = append(webfingers, words[i]+"@test.com")
	}
	log.Debug().Msg("webfingers created")
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	for i := range words {
		groupNames = append(groupNames, strings.Join([]string{words[i], words[r.Intn(98)]}, " "))
		studioNames = append(studioNames, strings.Join([]string{words[i], words[r.Intn(98)]}, " "))
	}

	log.Debug().Msg("groupNames and studioNames created")

	inputData := make([]testData, 100)

	for j := range inputData {
		inputData[j] = testData{
			webfinger:       webfingers[j],
			mediaID:         uuids[j],
			mediaTitle:      strings.ToUpper(words[j]),
			mediaKind:       "album",
			personFirstName: words[j],
			personLastName:  words[r.Intn(98)],
			groupName:       groupNames[j],
			studioName:      studioNames[j],
		}

		err = insertTestData(ctx, dbConn, &inputData[j])
		log.Debug().Msgf("inserted data: %+v\n", inputData[j])
		if err != nil {
			b.Fatalf("error adding data to test tables: %v", err)
		}
	}

	log.Debug().Msg("Test data inserted")

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = PopulateCache(ctx, redisClient, dsn, &log, &cfg.TestConfig, true)
		if err != nil {
			b.Fatalf("error populating cache: %v", err)
		}
	}
}
