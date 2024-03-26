package search

import (
	"codeberg.org/mjh/LibRate/controllers/search/target"
	"github.com/blevesearch/bleve/v2"
	"github.com/samber/lo"
)

func filterByCategories(categories []target.Category) []bleve.FacetRequest {
	categoriesNames := lo.Map(categories, func(c target.Category, _ int) string {
		return c.String()
	})
	if lo.Contains(categoriesNames, target.Union.String()) {
		return nil
	}
	facets := make([]bleve.FacetRequest, len(categories))
	for i := range categories {
		fr := bleve.NewFacetRequest(categories[i].String(), 1)
		facets = append(facets, *fr)
	}
	return facets
}
