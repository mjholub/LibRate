package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"
)

// TODO: parameterize the search
/*
func Search() fiber.Handler {
	return func(c *fiber.Ctx) error {
return nil
	}
}
*/

func SearchMedia(c *fiber.Ctx) error {
	// Parse the search term from the request body
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	config := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	searchTerm := body["search"]
	dConn, err := db.Connect(&config)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to database",
		})
	}

	// Perform the search (this is just a placeholder - you'll need to implement the actual search logic)
	media, err := performSearch(dConn, searchTerm)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to perform search",
		})
	}

	// Return the search results
	return c.JSON(media)
}

func performSearch(db *sqlx.DB, searchTerm string) ([]models.Media, error) {
	// Create a context for the database query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define the SQL query
	query := `
		SELECT *
		FROM media
		WHERE title LIKE $1 OR description LIKE $1
	`

	// Perform the database query
	rows, err := db.QueryContext(ctx, query, fmt.Sprintf("%%%s%%", searchTerm))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the results
	var media []models.Media
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(m.UUID.String(), &m.Name, &m.Kind); err != nil {
			return nil, err
		}
		media = append(media, m)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return media, nil
}
