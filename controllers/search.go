package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"github.com/microcosm-cc/bluemonday"

	h "codeberg.org/mjh/LibRate/internal/handlers"
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
		wsAddr string
		log    *zerolog.Logger
		dbConn *pgxpool.Pool
	}
	// Search result holds the fields into which the results
	// of a full text search are marshalled
	SearchResult struct {
		Type string `json:"type" db:"type"`
		ID   string `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}

	wsClient struct {
		isClosing bool
		mu        sync.Mutex
	}
)

func NewSearchController(dbConn *pgxpool.Pool, log *zerolog.Logger, wsAddr string) *SearchController {
	return &SearchController{
		log:    log,
		wsAddr: wsAddr,
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

// GetWSAddress returns the address of the websocket server, minus the protocol,
// since that would initialize a new connection
func (sc *SearchController) GetWSAddress(c *fiber.Ctx) error {
	return c.JSON(sc.wsAddr)
}

func (sc *SearchController) WSHandler(c *websocket.Conn) {
	clients := make(map[*websocket.Conn]*wsClient)
	register := make(chan *websocket.Conn)
	broadcast := make(chan string) // or []byte?
	unregister := make(chan *websocket.Conn)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case conn := <-register:
				clients[conn] = &wsClient{}
				sc.log.Debug().Msg("websocket connection registered")
			case message := <-broadcast:
				sc.log.Debug().Msgf("broadcasting message: %s", message)
				for conn, c := range clients {
					go func(conn *websocket.Conn, c *wsClient) {
						c.mu.Lock()
						defer c.mu.Unlock()
						if c.isClosing {
							return
						}

						sc.log.Debug().Msgf("searching for: %s", message)
						res, err := performSearch(context.Background(), sc.dbConn, message)
						if err != nil {
							errChan <- err
							c.isClosing = true
							return
						}
						sc.log.Debug().Msgf("search results: %v", res)

						if err := conn.WriteJSON(res); err != nil {
							c.isClosing = true
							sc.log.Error().Err(err).Msg("failed to write to websocket")
						}
						if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
							errChan <- fmt.Errorf("failed to write close message to websocket: %w", err)
							c.isClosing = true
						}
						conn.Close()
						unregister <- conn
					}(conn, c)
				}
			case conn := <-unregister:
				delete(clients, conn)

				sc.log.Debug().Msg("websocket connection unregistered")

			}
		}
	}()

	err := <-errChan
	if err != nil {
		sc.log.Error().Err(err).Msg("websocket error")
	}
}

// performSearch performs a full text search on the database
func performSearch(ctx context.Context, db *pgxpool.Pool, searchTerm string) (res []SearchResult, err error) {
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
