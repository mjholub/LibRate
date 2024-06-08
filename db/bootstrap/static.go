package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CDN(ctx context.Context, db *pgxpool.Pool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS cdn.images (
			id BIGSERIAL PRIMARY KEY,
			source VARCHAR(255) NOT NULL,
			thumbnail VARCHAR(255) NOT NULL,
			alt VARCHAR(255)
		);`)
		if err != nil {
			return fmt.Errorf("failed to create cdn table: %w", err)
		}
		return nil
	}
}
