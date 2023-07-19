package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"

	"github.com/jmoiron/sqlx"
)

type (
	Keyword struct {
		ID         int32           `json:"_key" db:"id,pk"`
		Keyword    string          `json:"keyword" db:"keyword"`
		TotalStars int32           `json:"stars" db:"total_stars"`
		VoteCount  int32           `json:"vote_count" db:"vote_count"`
		AvgScore   sql.NullFloat64 `json:"avg_score" db:"avg_score"`
	}

	KeywordStorer interface {
		CastVote(ctx context.Context, k Keyword) error
		RemoveVote(ctx context.Context, k Keyword) error
		AddKeyword(ctx context.Context, k Keyword) error
		GetKeyword(ctx context.Context, mediaID uuid.UUID) (Keyword, error)
		GetKeywords(ctx context.Context, mediaID uuid.UUID) ([]Keyword, error)
	}

	KeywordStorage struct {
		db  *sqlx.DB
		log *zerolog.Logger
	}
)

func NewKeywordStorage(db *sqlx.DB, log *zerolog.Logger) *KeywordStorage {
	return &KeywordStorage{db, log}
}

func (ks *KeywordStorage) CastVote(ctx context.Context, k Keyword) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := ks.db.ExecContext(ctx, `
		UPDATE media.keywords SET total_stars = total_stars + $1, vote_count = vote_count + 1 WHERE id = $2`,
			k.TotalStars, k.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func (ks *KeywordStorage) RemoveVote(ctx context.Context, k Keyword) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := ks.db.ExecContext(ctx, `
		UPDATE media.keywords SET total_stars = total_stars - $1, vote_count = vote_count - 1 WHERE id = $2`,
			k.TotalStars, k.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func (ks *KeywordStorage) GetKeywords(ctx context.Context, mediaID uuid.UUID) (keywords []Keyword, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := ks.db.SelectContext(ctx, &keywords, `
SELECT keyword, total_stars::float / vote_count AS avg_score
FROM media.keywords WHERE media_id = $1
ORDER BY avg_score DESC`, mediaID)
		if err != nil {
			return nil, fmt.Errorf("error getting keywords: %w", err)
		}
		return keywords, nil
	}
}

func (ks *KeywordStorage) GetAll(ctx context.Context) (keywords []Keyword, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		err := ks.db.SelectContext(ctx, &keywords, `
			SELECT id, keyword FROM media.keywords`)
		if err != nil {
			return nil, fmt.Errorf("error getting keywords: %w", err)
		}
		return keywords, nil
	}
}

func (ks *KeywordStorage) GetKeyword(ctx context.Context, keyword string, mediaID uuid.UUID) (k Keyword, err error) {
	select {
	case <-ctx.Done():
		return k, ctx.Err()
	default:
		err := ks.db.GetContext(ctx, &k, `
SELECT id, keyword, total_stars::float / vote_count AS avg_score
FROM media.keywords WHERE keyword = $1 AND media_id = $2`, keyword, mediaID)
		if err != nil {
			return k, fmt.Errorf("error getting keyword: %w", err)
		}
		return k, nil
	}
}

func (ks *KeywordStorage) GetKeywordByID(ctx context.Context, id int32) (k Keyword, err error) {
	select {
	case <-ctx.Done():
		return k, ctx.Err()
	default:
		err := ks.db.GetContext(ctx, &k, `
SELECT id, keyword,
CASE WHEN vote_count = 0 THEN NULL
	ELSE	total_stars::float / NULLIF(vote_count, 0) 
	END AS avg_score
FROM media.keywords WHERE id = $1`, id)
		if err != nil {
			return k, fmt.Errorf("error getting keyword: %w", err)
		}
		return k, nil
	}
}

func (ks *KeywordStorage) AddKeyword(ctx context.Context, keyword string, mediaID uuid.UUID) (err error) {
	var k int32
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := ks.db.GetContext(ctx, &k, `
		INSERT INTO media.keywords (keyword, media_id) VALUES ($1, $2) RETURNING id`, keyword, mediaID)
		if err != nil {
			return err
		}
		ks.log.Debug().Msgf("Added keyword: %s for media: %s. ID is: %d", keyword, mediaID.String(), k)
		return nil
	}
}
