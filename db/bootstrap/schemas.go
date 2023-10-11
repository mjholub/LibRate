package bootstrap

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Schemas(ctx context.Context, db *sqlx.DB) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Println("Creating schemes...")
		errChan := make(chan error)
		schemaNames := []string{"media", "people", "places", "reviews", "cdn", "members"}
		for i := range schemaNames {
			go func(i int) {
				errChan <- createSchema(ctx, schemaNames[i], db)
			}(i)
		}
		for i := 0; i < len(schemaNames); i++ {
			err = <-errChan
			if err != nil {
				return err
			}
		}
		close(errChan)

		return nil
	}
}

func createSchema(ctx context.Context, name string, db *sqlx.DB) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", name))
		if err != nil {
			return fmt.Errorf("failed to create schema %s: %v", name, err)
		}
		return nil
	}
}
