package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Review(ctx context.Context, connection *sqlx.DB) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err = connection.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS reviews;`,
		)
		if err != nil {
			return fmt.Errorf("failed to create reviews schema: %w", err)
		}
		_, err = connection.ExecContext(ctx, ` 
		CREATE TABLE IF NOT EXISTS reviews.ratings (
		uuid UUID PRIMARY KEY,
    stars SMALLINT NOT NULL CHECK (stars >= 1 AND stars <= 10),
    comment TEXT,
    topic TEXT,
    attribution TEXT,
    user_id UNSIGNED INTEGER REFERENCES public.members(id),
    media_id UUID REFERENCES media.media(id),
		);`,
		)
		if err != nil {
			return fmt.Errorf("failed to create ratings table: %w", err)
		}
		return nil
	}
}
