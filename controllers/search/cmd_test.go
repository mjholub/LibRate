// also covers filter.go
package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// The goal is to perform a search operation that will return
// a result more or less equal to that of running
// bleve query page-index.bleve "Neofolk"
// as we know that the index should contain the word "Neofolk"
func TestRunQuery(t *testing.T) {
	opts := Options{
		Query:          "Neofolk",
		Sort:           "",
		SortDescending: true,
		Fuzzy:          true,
		Page:           uint(0),
		PageSize:       uint(10),
		Categories:     []target.Category{target.Genres},
	}
	idx, err := bleve.Open("../../page-index.bleve")
	require.NoErrorf(t, err, "error opening index: %v", err)
	f, err := idx.Fields()
	require.NoError(t, err)
	fmt.Printf("%+v", f)
	logger := zerolog.Nop()

	s := &Service{
		i:   idx,
		log: &logger,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := s.RunQuery(ctx, &opts)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.GreaterOrEqual(t, uint64(1), res.Total)
	assert.NotEmpty(t, res.Status.Successful)
}
