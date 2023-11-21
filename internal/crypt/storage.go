package crypt

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	sqlite3 "github.com/dm764/go-sqlcipher/v4"
	"github.com/gofrs/uuid/v5"
)

// CreateCryptoStorage creates a SQLite-cypher encrypted storage for X25519 keys
func CreateCryptoStorage() (conn *sql.DB, err error) {
	key, err := generateStorageKey()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		":memory:?_pragma_key=%s&_pragma_cipher_page_size=4096", key)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to sqlite3 secrets storage: %v", err)
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	encrypted, err := sqlite3.IsEncrypted(":memory:")
	if err != nil {
		return nil, fmt.Errorf("error checking encryption status")
	}
	if !encrypted {
		return nil, fmt.Errorf("go-sqlcipher: error checking encryption status")
	}

	_, err = db.Exec(`CREATE TABLE keys(
		id CHARACTER(36) PRIMARY KEY,
		private TEXT NOT NULL,
		public TEXT NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("error creating keys table: %v", err)
	}

	return db, nil
}

func generateStorageKey() (string, error) {
	part1, _ := uuid.NewV7()
	part2, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("error generating UUID: %v", err)
	}
	return url.QueryEscape(
		strings.ReplaceAll(part1.String()+part2.String(), "-", "")), nil
}
