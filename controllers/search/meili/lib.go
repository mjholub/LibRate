package meili

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers/search/target"
	searchdb "codeberg.org/mjh/LibRate/models/search"
)

type (
	Service struct {
		searchdb   *searchdb.Storage
		client     *meilisearch.Client
		log        *zerolog.Logger
		validation *validator.Validate
	}

	// Options defines a set of optional search parameters.
	Options struct {
		// Query is the search query used for a text search.
		Query string `json:"query" query:"q" default:""`

		// Sort is the field, that should be sorted by.
		// When left empty, the default sorting is used.
		Sort string `json:"sort,omitempty" query:"sort,omitempty" validate:"oneof=score added modified name"`

		// SortDescending defines the sort order.
		SortDescending bool `json:"sortDescending" query:"desc" default:"true"`

		// Fuzzy defines whether to use fuzzy or wildcard search.
		Fuzzy bool `json:"fuzzy,omitempty" query:"fuzzy" default:"false"`

		// Page is current page.
		Page uint `json:"page" query:"page" default:"0"`

		// PageSize defines the number of hits returned per page.
		//
		// PageSize is infinite when set to 0 (i.e. infinite scroll).
		PageSize int64 `json:"pageSize" query:"pageSize" default:"10" validate:"gte=0,lte=180"`

		// Categories are the categories to search in. By default,
		// a Union category is performed to search in all categories.
		Categories []target.Category `json:"categories" query:"category" validate:"unique,dive" default:"union"`
	}
)

func Connect(
	ctx context.Context,
	conf *cfg.SearchConfig,
	log *zerolog.Logger,
	v *validator.Validate,
	searchStorage *searchdb.Storage,
) (*Service, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   fmt.Sprintf("http://%s:%d/", conf.Meili.Host, conf.Meili.Port),
		APIKey: conf.Meili.MasterKey,
	})

	err := retry.Do(
		func() error {
			if client.IsHealthy() {
				return nil
			}
			return fmt.Errorf("meilisearch client healthcheck failed")
		},
		retry.Attempts(5),
		retry.Delay(500*time.Millisecond),
		retry.OnRetry(func(n uint, _ error) {
			log.Warn().Msgf("retrying meilisearch client healthcheck, attempt %d", n)
		},
		))
	if err != nil {
		return nil, fmt.Errorf("meilisearch client healthcheck failed (too many attempts)")
	}

	// Create the indexes
	svc := &Service{
		searchdb:   searchStorage,
		client:     client,
		log:        log,
		validation: v,
	}
	start := time.Now()
	if err := svc.CreateAllIndexes(ctx); err != nil {
		return nil, fmt.Errorf("error creating indexes: %w", err)
	}
	svc.log.Info().Msgf("indexes created in %s", time.Since(start))

	return svc, nil
}
