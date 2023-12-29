package bootstrap

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
)

func Schemas(ctx context.Context, db *sqlx.DB) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Println("Creating schemas...")
		var wg sync.WaitGroup
		schemaNames := []string{"media", "people", "places", "reviews", "cdn", "members"}
		for i := range schemaNames {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				err := createSchema(ctx, schemaNames[i], db)
				if err != nil {
					fmt.Printf("Error creating schema %s: %v\n", schemaNames[i], err)
				} else {
					fmt.Printf("Created schema %s\n", schemaNames[i])
				}
			}(i)
		}
		wg.Wait()

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
