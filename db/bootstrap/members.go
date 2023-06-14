package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Members(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS public.members (
			id SERIAL PRIMARY KEY,
			uuid UUID NOT NULL,
			nick VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			passhash VARCHAR(255) NOT NULL,
			reg_timestamp TIMESTAMP DEFAULT NOW() NOT NULL 
		);
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
	`)
		if err != nil {
			return fmt.Errorf("failed to create members table: %w", err)
		}
		return nil
	}
}
