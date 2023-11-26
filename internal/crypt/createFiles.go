package crypt

import (
	"fmt"
	"os"
)

// CreateFiles creates the temporary directory and database file
func CreateFiles() (dbFile *os.File, dbDir string, err error) {
	// TODO: add a goroutine to automatically rotate the generated key
	// also, this can probably be simplified to use sqlx.DB
	dbDir, err = os.MkdirTemp("", "librate-secrets")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temporary directory for secrets: %v", err)
	}

	dbFile, err = os.CreateTemp(dbDir, "secrets.db")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temporary file for secrets: %v", err)
	}

	return dbFile, dbDir, nil
}
