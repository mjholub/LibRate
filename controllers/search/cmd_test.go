// also covers filter.go
package search

import (
	"testing"

	"github.com/blevesearch/bleve/v2"
	"github.com/stretchr/testify/assert"

	"codeberg.org/mjh/LibRate/controllers/search/target"
)

func TestBuildSearchRequest(t *testing.T) {
	opts := Options{
		Query:          "test",
		Sort:           "score",
		SortDescending: true,
		Fuzzy:          false,
		Page:           uint(0),
		PageSize:       uint(10),
		Categories:     []target.Category{target.Genres},
	}

	q := bleve.NewMatchQuery(opts.Query)
	req := buildSearchRequest(&opts, q)
	assert.NotNil(t, req)
	assert.Equalf(t, 10, req.Size, "expected hits per page to equal %d, got %d", 10, req.Size)
	assert.Equalf(t, 0, req.From, "expected page to equal %d, got %d", 0, req.From)
	genreFields := []string{
		"name",
		"kinds",
		"description",
		"language",
		"characteristics",
	}
	assert.Equal(t, req.Fields, genreFields)
}
