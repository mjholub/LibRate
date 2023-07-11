package controllers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

type SearchController struct {
	dbConn *sqlx.DB
}

type SearchResult struct {
	Type string `json:"type" db:"type"`
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func NewSearchController(dbConn *sqlx.DB) *SearchController {
	return &SearchController{
		dbConn: dbConn,
	}
}

func (sc *SearchController) Search(c *fiber.Ctx) error {
	// Parse the search term from the request body
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid request body")
	}
	searchTerm := body["search"]
	// Perform the search (this is just a placeholder - you'll need to implement the actual search logic)
	results, err := performSearch(c.Context(), sc.dbConn, searchTerm)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to perform search"+err.Error())
	}

	// Return the search results
	return c.JSON(results)
}

func performSearch(ctx context.Context, db *sqlx.DB, searchTerm string) (res []SearchResult, err error) {
	// Define the SQL query
	stmt, err := db.PreparexContext(ctx, `
		SELECT 'person' AS type, id::text, first_name AS name
		FROM people.person
		WHERE to_tsvector('english', first_name || ' ' || last_name) @@ plainto_tsquery('english', $1)

		UNION ALL

		SELECT 'group' AS type, id::text, name
		FROM people."group"
		WHERE to_tsvector('english', name) @@ plainto_tsquery('english', $1)

		UNION ALL

		SELECT 'genre' AS type, id::text, name
		FROM media.genres
		WHERE to_tsvector('english', name) @@ plainto_tsquery('english', $1)

		UNION ALL

		SELECT 'studio' AS type, id::text, name
		FROM people.studio
		WHERE to_tsvector('english', name) @@ plainto_tsquery('english', $1)

		UNION ALL

		SELECT 'media' AS type, id::text, title AS name
		FROM media.media
		WHERE to_tsvector('english', title) @@ plainto_tsquery('english', $1)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare search query: %w", err)
	}

	// Perform the database query
	rows, err := stmt.QueryxContext(ctx, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	defer rows.Close()

	// Parse the results
	for rows.Next() {
		var r SearchResult
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over search results: %w", err)
	}

	return res, nil
}
