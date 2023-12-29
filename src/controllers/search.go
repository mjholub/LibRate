package controllers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"github.com/microcosm-cc/bluemonday"
)

type (
	// ISearchController is the interface for the search controller.
	// It is mostly useful for automatically generating mocks for unit tests.
	ISearchController interface {
		Search(c *fiber.Ctx) error
	}

	// SearchController is the controller for search endpoints
	// It provides a bridge between the HTTP layer and the database layer
	SearchController struct {
		dbConn *sqlx.DB
	}
	// Search result holds the fields into which the results
	// of a full text search are marshalled
	SearchResult struct {
		Type string `json:"type" db:"type"`
		ID   string `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}
)

func NewSearchController(dbConn *sqlx.DB) *SearchController {
	return &SearchController{
		dbConn: dbConn,
	}
}

// Search calls the private function performSearch to perform a full text search
func (sc *SearchController) Search(c *fiber.Ctx) error {
	policy := bluemonday.StrictPolicy()
	// Parse the search term from the request body
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid request body")
	}
	searchTerm := policy.Sanitize(body["search"])
	// Perform the search and handle any errors
	// The context is acquired from the request
	results, err := performSearch(c.Context(), sc.dbConn, searchTerm)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to perform search"+err.Error())
	}

	// Return the search results
	return c.JSON(results)
}

// performSearch performs a full text search on the database
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
