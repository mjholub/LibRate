// This package handles the retrieval of data from the CouchDB search database
// Synchronization is handled by triggers in the database
// using the http postgresql extension, without the
// need for postgres->application->couchdb->app->postgres-
// round trip (we need to store the returned document id to ensure deduplication)
package searchdb

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
)

type Storage struct {
	log    *zerolog.Logger
	config *cfg.Search
	client *kivik.Client
}

func Connect(config *cfg.Search, log *zerolog.Logger) (*Storage, error) {
	dsn := fmt.Sprintf("http://%s:%s@%s:%d",
		config.User,
		config.Password,
		config.Host,
		config.Port)

	client, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to couchdb: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if ok, err := client.Ping(ctx); !ok {
		return nil, fmt.Errorf("error establishing connection to the search database: %v", err)
	}

	return &Storage{
		config: config,
		client: client,
		log:    log,
	}, nil
}
