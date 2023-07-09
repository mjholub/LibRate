package models

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
)

type Keyword struct {
	ID      int32     `json:"_key" db:"id,pk"`
	Keyword string    `json:"keyword" db:"keyword"`
	MediaID uuid.UUID `json:"mediaid" db:"media_id"`
	Stars   uint8     `json:"stars"`
}

func CastVote(ctx context.Context, k Keyword) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		cfg := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
		db, err := db.Connect(&cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.ExecContext(ctx, `
		UPDATE media.keywords SET total_stars = total_stars + $1, vote_count = vote_count + 1 WHERE id = $2`,
			k.Stars, k.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func RemoveVote(ctx context.Context, k Keyword) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		cfg := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
		db, err := db.Connect(&cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.ExecContext(ctx, `
		UPDATE media.keywords SET total_stars = total_stars - $1, vote_count = vote_count - 1 WHERE id = $2`,
			k.Stars, k.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func GetKeywords(ctx context.Context, mediaID uuid.UUID) (keywords []Keyword, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		cfg := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
		db, err := db.Connect(&cfg)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		err = db.SelectContext(ctx, &keywords, `
SELECT keyword, total_stars::fload / vote_count AS avg_score
FROM media.keywords WHERE media_id = $1
ORDER BY avg_score DESC`, mediaID)
		if err != nil {
			return nil, fmt.Errorf("error getting keywords: %w", err)
		}
		return keywords, nil
	}
}

func GetKeyword(ctx context.Context, keyword string, mediaID uuid.UUID) (k Keyword, err error) {
	select {
	case <-ctx.Done():
		return k, ctx.Err()
	default:
		cfg := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
		db, err := db.Connect(&cfg)
		if err != nil {
			return k, err
		}
		defer db.Close()

		err = db.GetContext(ctx, &k, `
SELECT id, keyword, total_stars::float / vote_count AS avg_score
FROM media.keywords WHERE keyword = $1 AND media_id = $2`, keyword, mediaID)
		if err != nil {
			return k, fmt.Errorf("error getting keyword: %w", err)
		}
		return k, nil
	}
}

func AddKeyword(ctx context.Context, keyword string, mediaID uuid.UUID) (err error) {
	var k int32
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		cfg := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
		db, err := db.Connect(&cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		err = db.GetContext(ctx, &k, `
		INSERT INTO media.keywords (keyword, media_id) VALUES ($1, $2) RETURNING id`, keyword, mediaID)
		if err != nil {
			return err
		}
		log.Debug().Msgf("Added keyword: %s for media: %s. ID is: %d", keyword, mediaID.String(), k)
		return nil
	}
}
