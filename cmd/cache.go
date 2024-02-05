package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/media"
)

type person struct {
	FullName  string         `json:"full_name"`
	NickNames sql.NullString `json:"nick_names"`
	Roles     []string       `json:"roles"`
}

// Populate cache performs a delta update of the redis
// cache with searchable basic data, like artist names,
// for quick search and retrieval.
func PopulateCache(
	cacheServer *redis.Client,
	db *pgxpool.Pool,
	log *zerolog.Logger,
	config *cfg.Config,
	testMode bool,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Redis.UpdateFrequency)*time.Second)

	defer cancel()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("error starting transaction to update cache: %v", err)
	}

	// nolint: errcheck // in case of error, the transaction will be rolled back
	defer tx.Rollback(ctx)

	lastUpdateVal, err := cacheServer.Get(ctx, "last_update").Result()
	// first run scenario
	if err == redis.Nil {
		// time.Unix is slightly faster than time.Now
		// especially since it's not returned anywhere to the presentation layer
		// we can use raw integer timestamps
		// Also, add the number of seconds specified by the update frequency,
		// since we'll later count backwards from the last update time in
		// the SQL query
		lastUpdateVal = strconv.FormatInt(time.Unix(0, 0).
			Add(
				time.Duration(
					config.Redis.UpdateFrequency)*time.Second).Unix(), 10)
	}
	if err != nil && err != redis.Nil {
		log.Error().Err(err).Msg("error getting last update time from cache")
	}

	// convert the string value returned by redis to int64
	lastUpdate, err := strconv.ParseInt(lastUpdateVal, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("error parsing last update time from cache")
	}

	webfingers, err := tx.Query(ctx, `SELECT webfinger FROM public.members WHERE modified > $1`, lastUpdate-config.Redis.UpdateFrequency)
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

		usernames, err := json.Marshal(users)
		if err != nil {
			return fmt.Errorf("error marshalling usernames: %v", err)
		}

		if err = cacheServer.Set(ctx, "users", usernames, 0).Err(); err != nil {
			return fmt.Errorf("error setting users in cache: %v", err)
		}
	}

	mediaRows, err := tx.Query(ctx, `SELECT m.title, m.kind, c.source 
		FROM media.media AS m 
		JOIN media.media_images AS mi
		ON m.id = mi.media_id 
		JOIN cdn.images AS c ON mi.image_id = c.id
		WHERE m.modified > $1 AND mi.is_main = true
`, lastUpdate-config.Redis.UpdateFrequency)

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

		mediaData, err := json.Marshal(mediaList)
		if err != nil {
			return fmt.Errorf("error marshalling media: %v", err)
		}

		if err = cacheServer.Set(ctx, "media", mediaData, 0).Err(); err != nil {
			return fmt.Errorf("error setting media in cache: %v", err)
		}
	}

	people, err := tx.Query(ctx, `SELECT 
	CONCAT(first_name, ' ', last_name), 
	nick_names, 
	roles 
	FROM people.person WHERE modified > $1`,
		lastUpdate-config.Redis.UpdateFrequency)
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

		peopleData, err := json.Marshal(peopleList)
		if err != nil {
			return fmt.Errorf("error marshalling people data: %v", err)
		}

		if err = cacheServer.Set(ctx, "people", peopleData, 0).Err(); err != nil {
			return fmt.Errorf("error setting people in cache: %v", err)
		}
	}

	groups, err := tx.Query(ctx,
		`SELECT name, kind FROM people.group WHERE modified > $1`,
		lastUpdate-config.Redis.UpdateFrequency)
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

		groupData, err := json.Marshal(groupList)
		if err != nil {
			return fmt.Errorf("error marshalling groups: %v", err)
		}

		if err = cacheServer.Set(ctx, "groups", groupData, 0).Err(); err != nil {
			return fmt.Errorf("error setting groups in cache: %v", err)
		}
	}

	studios, err := tx.Query(ctx, `
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

		studioData, err := json.Marshal(studioList)
		if err != nil {
			return fmt.Errorf("error marshalling studios: %v", err)
		}

		if err = cacheServer.Set(ctx, "studios", studioData, 0).Err(); err != nil {
			return fmt.Errorf("error setting studios in cache: %v", err)
		}
	}

	// we can roll back the transaction here, as we're only reading
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// set the update time to now
	err = cacheServer.Set(ctx, "last_update", time.Now().Unix(), 0).Err()
	if err != nil {
		return fmt.Errorf("error setting last update time in cache: %v", err)
	}
	if testMode {
		return nil
	}

	time.Sleep(time.Duration(config.Redis.UpdateFrequency) * time.Second)
	// nolint: errcheck // if errors occur in the next iteration and exceed the limit, it'll be handled appropriately
	PopulateCache(cacheServer, db, log, config, false)

	return nil
}
