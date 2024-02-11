package search

import (
	"context"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-playground/validator/v10"
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
		log        *zerolog.Logger
	}

	Indexer interface {
		CreateIndex(path string) error
	}

	// Options defines a set of optional search parameters.
	Options struct {
		// Query is the search query used for a text search.
		Query string `json:"query"`

		// Sort is the field, that should be sorted by.
		// When left empty, the default sorting is used.
		Sort string `json:"sort" validate:"oneof=score added modified weighed_score review_count"`

		// LocalFirst determines whether the results from the current instance should be
		// preferred over remote results.
		LocalFirst bool `json:"localFirst" default:"true"`

		// SortDescending defines the sort order.
		SortDescending bool `json:"SortDescending" default:"true"`

		// Fuzzy defines whether to use fuzzy or wildcard search.
		Fuzzy bool `json:"fuzzy" default:"false"`

		// Page is current page.
		Page uint `json:"page" default:"0"`

		// PageSize defines the number of hits returned per page.
		//
		// PageSize is infinite when set to 0 (i.e. infinite scroll).
		PageSize uint `json:"pageSize" default:"10" validate:"gte=0,lte=180"`

		// Categories are the categories to search in. By default,
		// a Union category is performed to search in all categories.
		Categories []target.Category `json:"categories" validate:"unique,dive"`

		// Filters is a list of filters, that reduce the search result. All filters
		// are combined with AND logic in addition with the search query.
		Filters []interface{} `json:"filter" validate:"unique,dive"`

		// Aggregations is a map of aggregations, to perform aggregations on fields.
		// The provided map key can be used to identify the corresponding bucket in
		// the result.
		Aggregations []interface{} `json:"aggregations"`
	}
)

func NewService(
	ctx context.Context,
	validation *validator.Validate,
	storage *searchdb.Storage,
	indexPath string,
	log *zerolog.Logger,
) mo.Result[*Service] {
	return mo.Try(func() (*Service, error) {
		idx, err := bleve.Open(indexPath)
		if err != nil {
			if retry := CreateIndex(ctx, indexPath, storage, log); retry != nil {
				return nil, fmt.Errorf("tried to create the missing index, but failed: %v", retry)
			}
		}

		return &Service{validation, storage, idx, log}, nil
	})
}
