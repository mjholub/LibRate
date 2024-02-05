package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

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
// Likely not due to serialization
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

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()
	var wg sync.WaitGroup
	wg.Add(4)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		conn, err := db.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error acquiring database connection")
			return
		}
		defer conn.Release()
		if err := cacheUsers(ctx, conn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching users")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		conn, err := db.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error acquiring database connection")
			return
		}
		defer conn.Release()
		if err := cacheMedia(ctx, conn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching media")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		conn, err := db.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error acquiring database connection")
			return
		}
		defer conn.Release()
		if err := cachePeople(ctx, conn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching artists")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)
	go func(context.Context, string, *redis.Client, int64) {
		defer wg.Done()
		conn, err := db.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error acquiring database connection")
			return
		}
		defer conn.Release()
		if err := cacheGroups(ctx, conn, cacheServer, updateDelta); err != nil {
			log.Error().Err(err).Msg("error caching groups")
			return
		}
	}(ctx, dsn, cacheServer, updateDelta)

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

		currentVal, err := cacheServer.Get(ctx, "studios").Bytes()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("error getting studios from cache: %v", err)
		}

		var studioList []studio
		if currentVal != nil {
			if e := json.Unmarshal(currentVal, &studioList); e != nil {
				return fmt.Errorf("error unmarshalling studios: %v", err)
			}
		}
		for studios.Next() {
			var s studio

			if err = studios.Scan(&s.Name, &s.Kind); err != nil {
				return fmt.Errorf("error scanning studio row: %v", err)
			}

			studioList = append(studioList, s)
		}
		studios.Close()

		studioData, err := json.MarshalNoEscape(lo.Uniq(studioList))
		if err != nil {
			return fmt.Errorf("error marshalling studios: %v", err)
		}

		if err = cacheServer.Set(ctx, "studios", studioData, 0).Err(); err != nil {
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
	dbConn *pgxpool.Conn,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		webfingers, err := dbConn.Query(ctx,
			`SELECT webfinger FROM public.members WHERE modified > $1`,
			updateDelta,
		)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error querying for updated data: %v", err)
		}

		if err != pgx.ErrNoRows {
			var users []string
			currentVal, err := cacheServer.Get(ctx, "users").Bytes()
			if err != nil && err != redis.Nil {
				return fmt.Errorf("error getting users from cache: %v", err)
			}

			if currentVal != nil {
				if e := json.Unmarshal(currentVal, &users); e != nil {
					return fmt.Errorf("error unmarshalling users: %v", e)
				}
			}
			for webfingers.Next() {
				var wf string

				if err = webfingers.Scan(&wf); err != nil {
					return fmt.Errorf("error scanning webfinger row: %v", err)
				}
				users = append(users, wf)
			}
			webfingers.Close()

			userData, err := json.MarshalNoEscape(lo.Uniq(users))

			if err := cacheServer.Set(ctx, "users", lo.Uniq(userData), 0).Err(); err != nil {
				return fmt.Errorf("error setting users in cache: %v", err)
			}
		}
		return nil
	}
}

func cacheMedia(ctx context.Context,
	conn *pgxpool.Conn,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		mediaRows, err := conn.Query(ctx, `SELECT m.title, m.kind, c.source 
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
			currentVal, err := cacheServer.Get(ctx, "media").Bytes()
			if err != nil && err != redis.Nil {
				return fmt.Errorf("error getting media from cache: %v", err)
			}

			if currentVal != nil {
				if e := json.Unmarshal(currentVal, &mediaList); e != nil {
					return fmt.Errorf("error unmarshalling media: %v", e)
				}
			}
			for mediaRows.Next() {
				var sm media.SimplifiedMedia

				if err = mediaRows.Scan(&sm.Title, &sm.Kind, &sm.ImageSource); err != nil {
					return fmt.Errorf("error scanning media row: %v", err)
				}
				mediaList = append(mediaList, sm)
			}
			mediaRows.Close()
			var mediaData []byte

			mediaData, err = json.MarshalNoEscape(lo.Uniq(mediaList))
			if err != nil {
				return fmt.Errorf("error marshalling media: %v", err)
			}

			if err = cacheServer.Set(ctx, "media", mediaData, 0).Err(); err != nil {
				return fmt.Errorf("error setting media in cache: %v", err)
			}
		}
		return nil
	}
}

func cachePeople(ctx context.Context,
	conn *pgxpool.Conn,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		people, err := conn.Query(ctx, `SELECT 
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
			currentVal, err := cacheServer.Get(ctx, "people").Bytes()
			if err != nil && err != redis.Nil {
				return fmt.Errorf("error getting people from cache: %v", err)
			}

			if currentVal != nil {
				if err := json.Unmarshal(currentVal, &peopleList); err != nil {
					return fmt.Errorf("error unmarshalling people: %v", err)
				}
			}
			for people.Next() {
				var p person

				if err = people.Scan(&p.FullName, &p.NickNames, &p.Roles); err != nil {
					return fmt.Errorf("error scanning people row: %v", err)
				}

				peopleList = append(peopleList, p)
			}
			people.Close()

			// FIXME: nested slices prevent applying lo.Uniq
			peopleData, err := json.MarshalNoEscape(peopleList)
			if err != nil {
				return fmt.Errorf("error marshalling people: %v", err)
			}

			if err = cacheServer.Set(ctx, "people", peopleData, 0).Err(); err != nil {
				return fmt.Errorf("error setting people in cache: %v", err)
			}
		}

		return nil
	}
}

func cacheGroups(
	ctx context.Context,
	conn *pgxpool.Conn,
	cacheServer *redis.Client,
	updateDelta int64,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		groups, err := conn.Query(ctx,
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
			currentVal, err := cacheServer.Get(ctx, "groups").Bytes()
			if err != nil && err != redis.Nil {
				return fmt.Errorf("error getting groups from cache: %v", err)
			}

			if currentVal != nil {
				e := json.Unmarshal(currentVal, &groupList)
				if e != nil {
					return fmt.Errorf("error unmarshalling groups: %v", e)
				}
			}
			for groups.Next() {
				var g group

				if err = groups.Scan(&g.Name, &g.Kind); err != nil {
					return fmt.Errorf("error scanning group row: %v", err)
				}

				groupList = append(groupList, g)
			}
			groups.Close()

			groupData, err := json.MarshalNoEscape(lo.Uniq(groupList))
			if err != nil {
				return fmt.Errorf("error marshalling groups: %v", err)
			}

			if err = cacheServer.Set(ctx, "groups", groupData, 0).Err(); err != nil {
				return fmt.Errorf("error setting groups in cache: %v", err)
			}
		}
		return nil
	}
}
