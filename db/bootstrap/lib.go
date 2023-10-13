package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lib/pq"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

// CreateEnumType reads the sql file containing the function
// responsible for creating the enum types only if they do not exist already,
// to work around the lack of idempotency with type creation in postgres
func createEnumType(ctx context.Context, db *sqlx.DB, typeName, schema string, values ...string) error {
	if len(values) == 0 {
		return errors.New("no values for the enum type were provided, but are required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory %v", err)
	}

	librateDir := filepath.Join(home, ".local", "share", "LibRate", "lib")

	sqlFile, err := os.ReadFile(filepath.Join(librateDir, "create_type_if_ne.sql"))
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	hash, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	tempFile, err := os.CreateTemp("", fmt.Sprintf("temp_create_enum_%s.sql", hash.String()))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	replacements := map[string]string{
		"{{schema}}":      schema,
		"{{typeName}}":    typeName,
		"{{enum_values}}": formatValues(values),
	}

	re := regexp.MustCompile(`{{\w+}}`)
	substitutedSQL := re.ReplaceAllStringFunc(string(sqlFile), func(match string) string {
		return replacements[match]
	})

	_, err = tempFile.WriteString(substitutedSQL)
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	_, err = sqlx.LoadFileContext(ctx, db, tempFile.Name())
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		// skip if the type already exists (we're using a custom exception)
		// TODO: check if we can get rid of an external PL/pgSQL and just check here
		// if the pgErr.Code == "42P07" || pgErr.Code == "42710"
		if ok && pgErr.Error() == "pq: type_exists" {
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
