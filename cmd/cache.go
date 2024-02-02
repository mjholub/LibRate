package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
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

	time.Sleep(time.Duration(config.Redis.UpdateFrequency) * time.Second)
}
