package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

// Populate cache performs a delta update of the redis
// cache with searchable basic data, like artist names,
// for quick search and retrieval.
func PopulateCache(
	cacheServer *redis.Client,
	db *pgxpool.Pool,
	log *zerolog.Logger,
	config *cfg.Config,
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
	if err != nil {
		log.Error().Err(err).Msg("error getting last update time from cache")
	}

	// convert the string value returned by redis to int64
	lastUpdate, err := strconv.ParseInt(lastUpdateVal, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("error parsing last update time from cache")
	}

	// TODO: refactor the media column so that we can more easily find authors of media
	query := `
		SELECT webfinger FROM public.members WHERE modified > $1
		UNION ALL
		SELECT m.id, m.title, m.kind 
		FROM media.media AS m 
		JOIN media.media_images AS mi 
		ON m.id = mi.media_id 
		WHERE m.modified > $1 AND mi.is_main = true
		UNION ALL
		SELECT CONCAT(first_name, ' ', last_name), nick_names FROM people.person WHERE modified > $1 
		UNION ALL
		SELECT name, kind FROM people.groups WHERE modified > $1 
		UNION ALL
		SELECT name, kind FROM people.studio WHERE modified > $1
	`

	rows, err := tx.Query(ctx, query, lastUpdate-config.Redis.UpdateFrequency)
	if err != nil {
		return fmt.Errorf("error querying for updated data: %v", err)
	}

	var errorCount uint8

	// NOTE: not parallelizing this because to count the rows we'd
	// need to loop over rows.Next() twice, giving this operation a quadratic time complexity
	for rows.Next() {
		var k, v string

		if errorCount > config.Redis.MaxUpdateErrors || errorCount > 254 {
			return fmt.Errorf("too many errors updating cache, exiting")
		}

		err = rows.Scan(&k, &v)
		if err != nil {
			errorCount++
			log.Error().Err(err).Msg("error scanning row")
			continue
		}

		err = cacheServer.Set(ctx, k, v, 0).Err()
		if err != nil {
			errorCount++
			log.Error().Err(err).Msgf("error setting key %s in cache", k)
		}
	}

	// we can roll back the transaction here, as we're only reading
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// set the update time to now
	err = cacheServer.Set(ctx, "last_update", time.Now().Format(time.RFC3339), 0).Err()
	if err != nil {
		return fmt.Errorf("error setting last update time in cache: %v", err)
	}

	time.Sleep(time.Duration(config.Redis.UpdateFrequency) * time.Second)
	errorCount = 0
	// nolint: errcheck // if errors occur in the next iteration and exceed the limit, it'll be handled appropriately
	PopulateCache(cacheServer, db, log, config)

	return nil
}
