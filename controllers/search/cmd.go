package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

// RunQuery performs a search on the bleve index
func (s *Service) RunQuery(ctx context.Context, opts *Options) (res *bleve.SearchResult, err error) {
	queryVal := buildUniversalQuery(opts.Query, opts.Fuzzy)
	req := buildSearchRequest(opts, queryVal)
	s.log.Debug().Msgf("request: %+v", &req)
	return s.i.Search(req)
}

func buildUniversalQuery(queryVal string, fuzzy bool) query.Query {
	if queryVal == "" {
		wildcardQ := bleve.NewWildcardQuery("*")
		wildcardQ.SetField("query")
		return wildcardQ
	}
	if fuzzy {
		fuzzyQ := bleve.NewFuzzyQuery(queryVal)
		fuzzyQ.Term = queryVal
		fuzzyQ.SetField("query")
		return fuzzyQ
	} else {
		termQ := bleve.NewTermQuery(queryVal)
		termQ.Term = queryVal
		termQ.SetField("query")
		return termQ
	}
}

func buildSearchRequest(opts *Options, queryVal query.Query) *bleve.SearchRequest {
	req := bleve.NewSearchRequest(queryVal)
	if opts.Sort != "" && strings.Contains(opts.Sort, ",") {
		req.SortBy(strings.Split(opts.Sort, ","))
	} else if opts.Sort != "" {
		req.SortBy([]string{opts.Sort})
	}
	req.Size = int(opts.PageSize)
	req.From = int(opts.Page * opts.PageSize)

	if facets := filterByCategories(opts.Categories); facets != nil {
		fmt.Printf("facets: %+v", facets)
		for _, facet := range facets {
			req.AddFacet(facet.Field, &facet)
		}
	}

	req.Fields = []string{"Fields", "Data", "Type"}

	return req
}
