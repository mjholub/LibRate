package search

import (
	"context"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/storage/redis/v3"
	"github.com/rs/zerolog"
	"github.com/samber/mo"

	"codeberg.org/mjh/LibRate/controllers/search/target"
	searchdb "codeberg.org/mjh/LibRate/models/search"
)

type (
	Service struct {
		validation *validator.Validate
		storage    *searchdb.Storage
		i          bleve.Index
		cache      *redis.Storage
		log        *zerolog.Logger
	}

	Indexer interface {
		CreateIndex(path string) error
	}

	// Options defines a set of optional search parameters.
	Options struct {
		// Query is the search query used for a text search.
		Query string `json:"query" query:"q" default:""`

		// Sort is the field, that should be sorted by.
		// When left empty, the default sorting is used.
		Sort string `json:"sort,omitempty" query:"sort,omitempty" validate:"oneof=score added modified name"`

		// LocalFirst determines whether the results from the current instance should be
		// preferred over remote results.
		LocalFirst bool `json:"localFirst,omitempty" query:"local_first" default:"true"`

		// SortDescending defines the sort order.
		SortDescending bool `json:"SortDescending" query:"desc" default:"true"`

		// Fuzzy defines whether to use fuzzy or wildcard search.
		Fuzzy bool `json:"fuzzy,omitempty" query:"fuzzy" default:"false"`

		// Page is current page.
		Page uint `json:"page" query:"page" default:"0"`

		// PageSize defines the number of hits returned per page.
		//
		// PageSize is infinite when set to 0 (i.e. infinite scroll).
		PageSize uint `json:"pageSize" query:"pageSize" default:"10" validate:"gte=0,lte=180"`

		// Categories are the categories to search in. By default,
		// a Union category is performed to search in all categories.
		Categories []target.Category `json:"categories" query:"category" validate:"unique,dive" default:"union"`
	}
)

func NewService(
	ctx context.Context,
	validation *validator.Validate,
	storage *searchdb.Storage,
	indexPath string,
	cache *redis.Storage,
	log *zerolog.Logger,
) mo.Result[*Service] {
	return mo.Try(func() (*Service, error) {
		idx, err := bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.
				Errorf(
					`Missing search index.
				 Create one with lrctl search build.
				(available at https://codeberg.org/mjh/lrctl).
				Any request to /api/search/ will return 501 Not Implemented!
				`)
		}

		return &Service{validation, storage, idx, cache, log}, nil
	})
}

func ServiceNoIndex(
	validation *validator.Validate,
	storage *searchdb.Storage,
	cache *redis.Storage,
	log *zerolog.Logger,
) *Service {
	return &Service{validation, storage, nil, cache, log}
}
