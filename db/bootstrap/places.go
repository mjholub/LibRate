package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Places(ctx context.Context, db *pgxpool.Pool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		placeTypes := []string{"country", "city", "venue", "other"}
		err := createEnumType(ctx, db, "place_kind", "places", placeTypes...)
		if err != nil {
			return err
		}

		_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS places.country (
			id SMALLINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			code VARCHAR(2) NOT NULL
		);`)
		if err != nil {
			return fmt.Errorf("failed to create country table: %w", err)
		}
		_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS places.place (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			kind places.place_kind NOT NULL,
			lat FLOAT NOT NULL,
			lng FLOAT NOT NULL,
			country SMALLINT REFERENCES places.country(id)	
		);`)
		if err != nil {
			return fmt.Errorf("failed to create place table: %w", err)
		}
		_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS places.city (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			lat FLOAT NOT NULL,
			lng FLOAT NOT NULL,
			country SMALLINT REFERENCES places.country(id)
		);`)
		if err != nil {
			return fmt.Errorf("failed to create city table: %w", err)
		}
		_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS places.venue (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			active BOOLEAN NOT NULL,
			street VARCHAR(255) NOT NULL,
			zip VARCHAR(255),
			unit VARCHAR(255),
			city UUID REFERENCES places.city(uuid),
			country SMALLINT REFERENCES places.country(id)
		);`)
		if err != nil {
			return fmt.Errorf("failed to create venue table: %w", err)
		}
		return nil
	}
}
