package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// CreateEnumType reads the sql file containing the function
// responsible for creating the enum types only if they do not exist already,
// to work around the lack of idempotency with type creation in postgres
func createEnumType(ctx context.Context, db *pgxpool.Pool, typeName, schema string, values ...string) error {
	if len(values) == 0 {
		return errors.New("no values for the enum type were provided, but are required")
	}

	_, err := db.Exec(ctx, fmt.Sprintf("CREATE TYPE %s.%s AS ENUM (%s)",
		schema, typeName, formatValues(values)))
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("type assertion to *pq.Error failed. Error: %v", err)
		}
		// skip if the type already exist
		if pgErr.Code == "42P07" || pgErr.Code == "42710" {
			return nil
		}
		return fmt.Errorf(
			"failed to create ENUM type %s on schema %s: %s", typeName, schema, pgErr.Error())
	}

	return nil
}

func formatValues(values []string) (fmted string) {
	array := pq.Array(values) // something like &[foo bar]
	// remove the ampersand and brackets, wrap the values in single quotes and add separating commas
	fmted = strings.ReplaceAll(fmt.Sprintf("%s", array), "&[", "'")
	fmted = strings.ReplaceAll(fmted, "]", "'")
	fmted = strings.ReplaceAll(fmted, " ", "', '")
	return fmted
}
