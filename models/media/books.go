package media

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"

	"codeberg.org/mjh/LibRate/models/language"
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

func (ms *Storage) getBook(ctx context.Context, id uuid.UUID) (*Book, error) {
	rows, err := ms.db.Query(ctx, "SELECT * FROM books WHERE media_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("error getting book with ID %s: %v", id.String(), err)
	}
	defer rows.Close()
	var book Book
	for rows.Next() {
		if err := pgxscan.ScanRow(&book, rows); err != nil {
			return nil, fmt.Errorf("error scanning book with ID %s: %v", id.String(), err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows for book with ID %s: %v", id.String(), err)
	}

	return &book, nil
}

func (ms *Storage) AddBook(
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

		tx, err := ms.dbOld.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		batch := &pgx.Batch{}

		batch.Queue(`INSERT INTO media.books (
		title, publisher, publication_date,
		keywords, pages, isbn, asin, cover, summary
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
			book.Title, publisher.ID, book.PublicationDate,
			book.Keywords, book.Pages, book.ISBN, book.ASIN, book.Cover, book.Summary)

		for _, author := range book.Authors {
			batch.Queue(`INSERT INTO media.book_authors (book, person) VALUES ($1, $2)`, *mediaID, author.ID)
		}

		for _, genre := range book.Genres {
			batch.Queue(`INSERT INTO media.book_genres (book, genre) VALUES ($1, $2)`, *mediaID, genre.ID)
		}

		for i := range book.Languages {
			langID, err := language.ReverseLookupLangID(book.Languages[i])
			if err != nil {
				ms.Log.Error().Err(err).Msgf("error adding language %s to book with ID %s", book.Languages[i], *mediaID)
			}
			batch.Queue(`INSERT INTO media.book_languages (book, lang) VALUES ($1, $2)`, *mediaID, langID)
		}

		br := tx.SendBatch(ctx, batch)
		err = br.Close()
		if err != nil {
			return fmt.Errorf("error executing batch: %w", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}

		return nil
	}
}

func (ms *Storage) addBookAsMedia(ctx context.Context, book *Book) (mediaID *uuid.UUID, err error) {
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
