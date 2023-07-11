package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

type (
	Book struct {
		MediaID         *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		Title           string     `json:"title" db:"title"`
		Authors         []Person   `json:"author" db:"author"`
		Publisher       string     `json:"publisher" db:"publisher"`
		PublicationDate time.Time  `json:"publication_date" db:"publication_date"`
		Genres          []string   `json:"genres" db:"genres"`
		Keywords        []string   `json:"keywords,omitempty" db:"keywords,omitempty"`
		Languages       []string   `json:"languages" db:"languages"`
		Pages           int16      `json:"pages" db:"pages"`
		ISBN            string     `json:"isbn,omitempty" db:"isbn,unique,omitempty"`
		ASIN            string     `json:"asin,omitempty" db:"asin,unique,omitempty"`
		Cover           string     `json:"cover,omitempty" db:"cover,omitempty"`
		Summary         string     `json:"summary" db:"summary"`
	}

	BookValues interface {
		[]string | string | int16 | time.Time | []Person
	}
)

func (ms *MediaStorage) getBook(ctx context.Context, id uuid.UUID) (Book, error) {
	stmt, err := ms.db.PrepareContext(ctx, "SELECT * FROM books WHERE media_id = ?")
	if err != nil {
		return Book{}, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	var book Book
	if err := row.Scan(&book); err != nil {
		return Book{}, fmt.Errorf("error scanning row: %w", err)
	}

	return book, nil
}

func addBook(ctx context.Context, db *sqlx.DB, keys []string, book Book) error {
	if !lo.Every(BookKeys, keys) {
		quoted := make([]string, len(BookKeys))
		for i, key := range BookKeys {
			quoted[i] = fmt.Sprintf("'%s'", key)
		}
		return fmt.Errorf("keys not a subset of book keys (%s)", strings.Join(quoted, ", "))
	}

	kvs := lo.Associate(keys, func(key string) (keys string, values interface{}) {
		switch key {
		case "media_id":
			return uuid.Must(uuid.NewV4()).String(), book.MediaID
		case "authors":
			return "authors", book.Authors
		default:
			return key, values
		}
	})
	_, err := db.NamedExecContext(ctx, "INSERT INTO books (:keys) VALUES (:values)", kvs)
	if err != nil {
		return err
	}
	return nil
}
