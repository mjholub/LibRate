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
		id BIGSERIAL PRIMARY KEY,
    stars SMALLINT NOT NULL CHECK (stars >= 1 AND stars <= 10),
    comment TEXT,
    topic TEXT,
    attribution TEXT,
    user_id SERIAL REFERENCES public.members(id),
    media_id UUID REFERENCES media.media(id)
		);`,
		)
		if err != nil {
			return fmt.Errorf("failed to create ratings table: %w", err)
		}

		_, err = connection.ExecContext(ctx, `
CREATE TABLE reviews.track_ratings (
	id bigserial NOT NULL,
	track uuid NOT null references media.tracks(media_id),
	stars smallint NOT null check (stars >=1 and stars <=10),
	user_id int references public.members(id)
);`)
		if err != nil {
			return fmt.Errorf("failed to create track_ratings table: %w", err)
		}
		_, err = connection.ExecContext(ctx, `
CREATE TABLE reviews.cast_ratings (
	id bigserial NOT NULL,
	cast_id bigint not null references people.cast(id),
	stars int2 not null check (stars >=1 and stars <=10),
	user_id int references public.members(id)
);
`)
		return nil
	}
}
