package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func CDN(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cdn.images (
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
