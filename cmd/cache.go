package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/media"
)

type person struct {
	FullName  string           `json:"full_name"`
	NickNames []sql.NullString `json:"nick_names"`
	Roles     []string         `json:"roles"`
}

// Populate cache performs a delta update of the redis
// cache with searchable basic data, like artist names,
// for quick search and retrieval.
// PERF: this is quite slow. For 100 entries in each table, it takes
// over 7 seconds on a high-end desktop machine
// Likely culprit is the JSON marshalling, which is quite slow
// Try using e.g. cap'n'proto or another serialization format
func PopulateCache(
	ctx context.Context,
	cacheServer *redis.Client,
	dsn string,
	log *zerolog.Logger,
	config *cfg.Config,
	testMode bool,
) error {
	var lastUpdate int64
	err := cacheServer.Get(ctx, "last_update").Scan(&lastUpdate)
	// first run scenario
	if err == redis.Nil {
		// time.Unix is slightly faster than time.Now
		// especially since it's not returned anywhere to the presentation layer
		// we can use raw integer timestamps
		// Also, add the number of seconds specified by the update frequency,
		// since we'll later count backwards from the last update time in
		// the SQL query
		lastUpdate = time.Unix(0, 0).
			Add(
				time.Duration(
					config.Redis.UpdateFrequency) * time.Second).Unix()
	}
	if err != nil && err != redis.Nil {
		log.Error().Err(err).Msg("error getting last update time from cache")
	}

	updateDelta := lastUpdate - config.Redis.UpdateFrequency

	var wg sync.WaitGroup
	wg.Add(4)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		if err := cacheUsers(ctx, dsn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching users")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		if err := cacheMedia(ctx, dsn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching media")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		if err := cachePeople(ctx, dsn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching artists")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		if err := cacheGroups(ctx, dsn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching groups")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	studios, err := db.Query(ctx, `
	SELECT name, kind 
	FROM people.studio 
	WHERE modified > $1`,
		lastUpdate-config.Redis.UpdateFrequency)

	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("error querying for updated data: %v", err)
	}

	if err != pgx.ErrNoRows {
		type studio struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		}

		var studioList []studio
		for studios.Next() {
			var s studio

			if err = studios.Scan(&s.Name, &s.Kind); err != nil {
				return fmt.Errorf("error scanning studio row: %v", err)
			}

			studioList = append(studioList, s)
		}
		studios.Close()
		var studioData bytes.Buffer

		enc := bin.NewBorshEncoder(&studioData)

		err = enc.Encode(studioList)
		if err != nil {
			return fmt.Errorf("error marshalling studios: %v", err)
		}

		if err = cacheServer.Set(ctx, "studios", studioData.Bytes(), 0).Err(); err != nil {
			return fmt.Errorf("error setting studios in cache: %v", err)
		}
	}

	wg.Wait()

	// set the update time to now
	err = cacheServer.Set(ctx, "last_update", time.Now().Unix(), 0).Err()
	if err != nil {
		return fmt.Errorf("error setting last update time in cache: %v", err)
	}
	if testMode {
		return nil
	}

	time.Sleep(time.Duration(config.Redis.UpdateFrequency) * time.Second)
	db.Close()
	// nolint: errcheck // if errors occur in the next iteration and exceed the limit, it'll be handled appropriately
	PopulateCache(ctx, cacheServer, dsn, log, config, false)

	return nil
}

func cacheUsers(ctx context.Context,
	dsn string,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		dbConn, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("error connecting to database: %v", err)
		}
		defer dbConn.Close()

		webfingers, err := dbConn.Query(ctx,
			`SELECT webfinger FROM public.members WHERE modified > $1`,
			updateDelta,
		)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error querying for updated data: %v", err)
		}

		if err != pgx.ErrNoRows {
			var users []string
			for webfingers.Next() {
				var wf string

				if err = webfingers.Scan(&wf); err != nil {
					return fmt.Errorf("error scanning webfinger row: %v", err)
				}

				users = append(users, wf)
			}
			webfingers.Close()
			var usernames bytes.Buffer

			enc := bin.NewBorshEncoder(&usernames)

			err := enc.Encode(users)
			if err != nil {
				return fmt.Errorf("error marshalling usernames: %v", err)
			}

			if err = cacheServer.Set(ctx, "users", usernames.Bytes(), 0).Err(); err != nil {
				return fmt.Errorf("error setting users in cache: %v", err)
			}
		}
		return nil
	}
}

func cacheMedia(ctx context.Context,
	dsn string,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		db, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("error connecting to database: %v", err)
		}
		defer db.Close()

		mediaRows, err := db.Query(ctx, `SELECT m.title, m.kind, c.source 
		FROM media.media AS m 
		JOIN media.media_images AS mi
		ON m.id = mi.media_id 
		JOIN cdn.images AS c ON mi.image_id = c.id
		WHERE m.modified > $1 AND mi.is_main = true
`, updateDelta)

		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error querying for updated data in media tables: %v", err)
		}

		if err != pgx.ErrNoRows {
			var mediaList []media.SimplifiedMedia
			for mediaRows.Next() {
				var sm media.SimplifiedMedia

				if err = mediaRows.Scan(&sm.Title, &sm.Kind, &sm.ImageSource); err != nil {
					return fmt.Errorf("error scanning media row: %v", err)
				}
				mediaList = append(mediaList, sm)
			}
			mediaRows.Close()
			var mediaData bytes.Buffer

			enc := bin.NewBorshEncoder(&mediaData)

			err := enc.Encode(mediaList)
			if err != nil {
				return fmt.Errorf("error marshalling media: %v", err)
			}

			if err = cacheServer.Set(ctx, "media", mediaData.Bytes(), 0).Err(); err != nil {
				return fmt.Errorf("error setting media in cache: %v", err)
			}
		}
		return nil
	}
}

func cachePeople(ctx context.Context,
	dsn string,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		db, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("error connecting to database: %v", err)
		}
		defer db.Close()

		people, err := db.Query(ctx, `SELECT 
	CONCAT(first_name, ' ', last_name), 
	nick_names, 
	roles 
	FROM people.person WHERE modified > $1`,
			updateDelta)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error querying for updated data in people table: %v", err)
		}

		if err != pgx.ErrNoRows {
			var peopleList []person
			for people.Next() {
				var p person

				if err = people.Scan(&p.FullName, &p.NickNames, &p.Roles); err != nil {
					return fmt.Errorf("error scanning people row: %v", err)
				}

				peopleList = append(peopleList, p)
			}
			people.Close()
			var peopleData bytes.Buffer

			enc := bin.NewBorshEncoder(&peopleData)

			err := enc.Encode(peopleList)
			if err != nil {
				return fmt.Errorf("error marshalling people: %v", err)
			}

			if err = cacheServer.Set(ctx, "people", peopleData.Bytes(), 0).Err(); err != nil {
				return fmt.Errorf("error setting people in cache: %v", err)
			}
		}

		return nil
	}
}

func cacheGroups(
	ctx context.Context,
	dsn string,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		db, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("error connecting to database: %v", err)
		}
		defer db.Close()

		groups, err := db.Query(ctx,
			`SELECT name, kind FROM people.group WHERE modified > $1`,
			updateDelta)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error querying for updated data in group tables: %v", err)
		}

		if err != pgx.ErrNoRows {
			type group struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			}

			var groupList []group
			for groups.Next() {
				var g group

				if err = groups.Scan(&g.Name, &g.Kind); err != nil {
					return fmt.Errorf("error scanning group row: %v", err)
				}

				groupList = append(groupList, g)
			}
			groups.Close()
			var groupData bytes.Buffer

			enc := bin.NewBorshEncoder(&groupData)

			err := enc.Encode(groupList)
			if err != nil {
				return fmt.Errorf("error marshalling groups: %v", err)
			}

			if err = cacheServer.Set(ctx, "groups", groupData.Bytes(), 0).Err(); err != nil {
				return fmt.Errorf("error setting groups in cache: %v", err)
			}
		}
		return nil
	}
}
