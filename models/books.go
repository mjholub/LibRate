package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lib/pq"
)

type (
	Book struct {
		MediaID         *uuid.UUID     `json:"media_id" db:"media_id,pk,unique"`
		Title           string         `json:"title" db:"title"`
		Authors         []Person       `json:"author" db:"author"`
		Publisher       Studio         `json:"publisher" db:"publisher"`
		PublicationDate sql.NullTime   `json:"publication_date" db:"publication_date"`
		Genres          []Genre        `json:"genres" db:"genres"`
		Keywords        pq.StringArray `json:"keywords,omitempty" db:"keywords,omitempty"`
		Languages       []string       `json:"languages" db:"languages"`
		Pages           int16          `json:"pages" db:"pages"`
		ISBN            sql.NullString `json:"isbn,omitempty" db:"isbn,unique,omitempty"`
		ASIN            sql.NullString `json:"asin,omitempty" db:"asin,unique,omitempty"`
		Cover           sql.NullString `json:"cover,omitempty" db:"cover,omitempty"`
		Summary         string         `json:"summary" db:"summary"`
	}

	BookValues interface {
		[]string | string | int16 | time.Time | []Person
	}
)

//nolint:gochecknoglobals //needed for iterative check during addition
var BookKeys = []string{
	"media_id", "title", "authors",
	"genres", "edition", "languages",
}

func (ms *MediaStorage) getBook(ctx context.Context, id uuid.UUID) (Book, error) {
	stmt, err := ms.db.PrepareContext(ctx, "SELECT * FROM books WHERE media_id = $1")
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

func (ms *MediaStorage) AddBook(
	ctx context.Context,
	book *Book,
	publisher *Studio,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if publisher == nil {
			return fmt.Errorf("publisher cannot be nil")
		}
		var err error
		publisher.ID, err = ms.Ps.GetID(ctx, publisher.Name, "studio")
		if err != nil {
			return fmt.Errorf("error getting publisher ID: %w", err)
		}

		mediaID, err := ms.addBookAsMedia(ctx, book)
		if err != nil {
			return err
		}

		_, err = ms.db.NamedExecContext(ctx, `
		INSERT INTO media.books (
		title, publisher, publication_date,
		keywords, pages, isbn, asin, cover, summary
		) VALUES (
		:title, :publisher, :publication_date,
		:keywords, :pages, :isbn, :asin, :cover, :summary
		`, book)
		if err != nil {
			return fmt.Errorf("error adding book: %w", err)
		}

		if err := ms.addBookAuthors(ctx, *mediaID, book.Authors); err != nil {
			return err
		}

		if err := ms.addBookGenres(ctx, *mediaID, book.Genres); err != nil {
			return err
		}

		if err := ms.addBookLanguages(ctx, *mediaID, book.Languages); err != nil {
			return err
		}

		return nil
	}
}

func (ms *MediaStorage) addBookAsMedia(ctx context.Context, book *Book) (mediaID *uuid.UUID, err error) {
	var created time.Time
	if book.PublicationDate.Valid {
		created = book.PublicationDate.Time
	} else {
		created = time.Now()
	}

	media := Media{
		Title:    book.Title,
		Kind:     "book",
		Created:  created,
		Creators: book.Authors,
	}

	mediaID, err = ms.Add(ctx, &media)
	if err != nil {
		return nil, fmt.Errorf("error adding media: %w", err)
	}
	ms.Log.Info().Msgf("added media with ID %s", mediaID)

	return mediaID, nil
}

func (ms *MediaStorage) addBookAuthors(ctx context.Context, bookID uuid.UUID, authors []Person) error {
	for i := range authors {
		authorID := authors[i].ID
		_, err := ms.db.ExecContext(ctx, `
		INSERT INTO media.book_authors (
		book, person
		) VALUES (
		$1, $2
		)`, bookID, authorID)
		if err != nil {
			return fmt.
				Errorf("error adding author %s %s to book with ID %s: %v",
					authors[i].FirstName, authors[i].LastName, bookID.String(), err)
		}
	}
	return nil
}

func (ms *MediaStorage) addBookGenres(ctx context.Context, bookID uuid.UUID, genres []Genre) error {
	for i := range genres {
		genreID := genres[i].ID
		_, err := ms.db.ExecContext(ctx, `
		INSERT INTO media.book_genres (
		book, genre
		) VALUES (
		$1, $2
		)`, bookID, genreID)
		if err != nil {
			return fmt.Errorf("error adding genre %s to book with ID %s: %v", genres[i].Name, bookID.String(), err)
		}
	}
	return nil
}

func (ms *MediaStorage) addBookLanguages(ctx context.Context, bookID uuid.UUID, languages []string) error {
	for i := range languages {
		langID, err := ReverseLookupLangID(languages[i])
		if err != nil {
			ms.Log.Error().Err(err).Msgf("error adding language %s to book with ID %s", languages[i], bookID)
		}
		_, err = ms.db.ExecContext(ctx, `
		INSERT INTO media.book_languages (
		book, lang
		) VALUES (
		$1, $2
		)`, bookID, langID)
		if err != nil {
			ms.Log.Error().Err(err).Msgf("error adding language %s to book with ID %s", languages[i], bookID)
		}
	}
	return nil
}
