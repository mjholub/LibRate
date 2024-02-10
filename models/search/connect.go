// This package handles the retrieval of data from the CouchDB search database
// Synchronization is handled by triggers in the database
// using the http postgresql extension, without the
// need for postgres->application->couchdb->app->postgres-
// round trip (we need to store the returned document id to ensure deduplication)
package searchdb

import (
	"fmt"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"

	"codeberg.org/mjh/LibRate/cfg"
)

type Storage struct {
	config *cfg.Search
	client *kivik.Client
}

func Connect(config *cfg.Search) (*Storage, error) {
	dsn := fmt.Sprintf("http://%s:%s@%s:%d",
		config.User,
		config.Password,
		config.Host,
		config.Port)

	client, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to couchdb: %w", err)
	}

	return &Storage{
		config: config,
		client: client,
	}, nil
}
