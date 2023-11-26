package crypt

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	sqlite3 "github.com/dm764/go-sqlcipher/v4"
	"github.com/gofrs/uuid/v5"
)

// cryptoStorage is a helper struct to create an instance of SQLite-cypher
// that implements fiber.Storage interface.
type CryptoStorage struct {
	db         *sql.DB
	gcInterval time.Duration
	done       chan struct{}
	sqlSelect  string
	sqlInsert  string
	sqlDelete  string
	sqlReset   string
	sqlGC      string
}

// Get returns the value of a key in the storage.
func (s *CryptoStorage) Get(key string) ([]byte, error) {
	if len(key) <= 0 {
		return nil, nil
	}
	row := s.db.QueryRow(s.sqlSelect, key)
	// Add db response to data
	var (
		data       = []byte{}
		exp  int64 = 0
	)
	if err := row.Scan(&data, &exp); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// If the expiration time has already passed, then return nil
	if exp != 0 && exp <= time.Now().Unix() {
		return nil, nil
	}

	return data, nil
}

// Set sets the value of a key in the storage.
// If the key does not exist, it will be created.
func (s *CryptoStorage) Set(key string, val []byte, exp time.Duration) error {
	if len(key) <= 0 || len(val) <= 0 {
		return nil
	}
	var expSeconds int64
	if exp != 0 {
		expSeconds = time.Now().Add(exp).Unix()
	}
	_, err := s.db.Exec(s.sqlInsert, key, val, expSeconds)
	return err
}

// Delete deletes a key in the storage.
func (s *CryptoStorage) Delete(key string) error {
	if len(key) <= 0 {
		return nil
	}
	_, err := s.db.Exec(s.sqlDelete, key)
	return err
}

// Reset resets the storage, removing all keys.
func (s *CryptoStorage) Reset() error {
	_, err := s.db.Exec(s.sqlReset)
	return err
}

// Close closes the storage.
func (s *CryptoStorage) Close() error {
	close(s.done)
	return s.db.Close()
}

func (s *CryptoStorage) gcTicker() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case t := <-ticker.C:
			s.gc(t)
		}
	}
}

// gc deletes all expired entries
func (s *CryptoStorage) gc(t time.Time) {
	_, _ = s.db.Exec(s.sqlGC, t.Unix())
}

func (s *CryptoStorage) Conn() *sql.DB {
	return s.db
}

// CreateCryptoStorage creates a SQLite-cypher encrypted storage for X25519 keys
// It needs to be called inside main function so that the temporary directory it uses
// is not discarded upon return.
func CreateCryptoStorage(dbFile string) (conn *CryptoStorage, err error) {
	key, err := generateStorageKey()
	if err != nil {
		return nil, err
	}

	dsn := dbFile + fmt.Sprintf(
		"?_pragma_key=%s&_pragma_cipher_page_size=4096", key)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to sqlite3 secrets storage: %v", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(1 * time.Second)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging secrets database: %v", err)
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	Table := "secrets"
	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		k TEXT PRIMARY KEY,
		v BLOB,
		e INTEGER DEFAULT 0
	);`, Table))
	if err != nil {
		return nil, fmt.Errorf("error creating secrets table: %v", err)
	}

	encrypted, err := sqlite3.IsEncrypted(dbFile)
	if err != nil {
		return nil, fmt.Errorf("error checking encryption status: %v", err)
	}
	if !encrypted {
		return nil, fmt.Errorf("go-sqlcipher: secrets database not encrypted")
	}

	store := &CryptoStorage{
		db:         db,
		gcInterval: 10 * time.Minute,
		done:       make(chan struct{}),
		sqlSelect:  fmt.Sprintf(`SELECT v, e FROM %s WHERE k=?;`, Table),
		sqlInsert:  fmt.Sprintf("INSERT OR REPLACE INTO %s (k, v, e) VALUES (?,?,?)", Table),
		sqlDelete:  fmt.Sprintf("DELETE FROM %s WHERE k=?", Table),
		sqlReset:   fmt.Sprintf("DELETE FROM %s;", Table),
		sqlGC:      fmt.Sprintf("DELETE FROM %s WHERE e <= ? AND e != 0", Table),
	}

	go store.gcTicker()

	return store, nil
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
