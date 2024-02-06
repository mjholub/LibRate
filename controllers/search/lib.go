package search

import (
	"codeberg.org/mjh/LibRate/controllers/search/target"
	"github.com/go-playground/validator/v10"
)

type (
	Service struct {
		validation *validator.Validate
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

		// Page is current page.
		Page uint `json:"page" default:"0"`

		// PageSize defines the number of hits returned per page.
		//
		// PageSize is infinite when set to 0 (i.e. infinite scroll).
		PageSize uint `json:"pageSize" default:"10" validate:"gte=0,lte=180"`

		// Filters is a list of filters, that reduce the search result. All filters
		// are combined with AND logic in addition with the search query.
		Filters []interface{} `json:"filter" validate:"unique,dive"`

		// Aggregations is a map of aggregations, to perform aggregations on fields.
		// The provided map key can be used to identify the corresponding bucket in
		// the result.
		Aggregations map[string]Aggregation `json:"aggregations"`
	}

	TermsFilter struct {
		Category target.Category `json:"category"`
		Terms    []string        `json:"terms"`
	}
)

func NewService(validation *validator.Validate) *Service {
	return &Service{validation}
}
