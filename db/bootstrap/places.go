package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Places(ctx context.Context, db *sqlx.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.Exec(`
		CREATE SCHEMA IF NOT EXISTS places;`)
		if err != nil {
			return fmt.Errorf("failed to create places schema: %w", err)
		}
		_, err = db.Exec(`
		CREATE ENUM IF NOT EXISTS places.kind AS ('country', 'city', 'venue', 'other');`)
		if err != nil {
			return fmt.Errorf("failed to create places kind enum: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS places.place (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			kind places.kind NOT NULL,
			lat FLOAT NOT NULL,
			lng FLOAT NOT NULL,
			country SMALLINT REFERENCES places.country(id)	
		);`)
		if err != nil {
			return fmt.Errorf("failed to create place table: %w", err)
		}
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS places.country (
			id SMALLINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			code VARCHAR(2) NOT NULL
		);`)
		if err != nil {
			return fmt.Errorf("failed to create country table: %w", err)
		}
		_, err = db.Exec(`
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
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS places.venue (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			active BOOLEAN NOT NULL,
			street VARCHAR(255) NOT NULL,
			zip VARCHAR(255) NOT NULL,
			unit VARCHAR(255) NOT NULL,
			city UUID REFERENCES places.city(uuid),
			country SMALLINT REFERENCES places.country(id)
		);`)
		if err != nil {
			return fmt.Errorf("failed to create venue table: %w", err)
		}
		return nil
	}
}
