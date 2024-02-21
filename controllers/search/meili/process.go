package meili

import (
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
)

type SearchResponse struct {
	Categories     []string      `json:"categories"`
	TotalHits      int64         `json:"totalHits"`
	ProcessingTime int64         `json:"processingTime"`
	Page           int64         `json:"page"`
	TotalPages     int64         `json:"totalPages"`
	HitsPerPage    int64         `json:"hitsPerPage"`
	Data           []interface{} `json:"data"`
}

func processResults(
	hitsPerPage int64,
	results map[string][]meilisearch.SearchResponse,
) (processed SearchResponse) {
	processed.Categories = lo.Keys(results)
	// sum the values of TotalHits from each response
	processed.TotalHits = lo.Reduce(lo.Flatten(lo.Values(results)),
		func(acc int64, r meilisearch.SearchResponse, i int) int64 {
			return acc + r.TotalHits
		}, int64(0))
	processed.ProcessingTime = lo.Reduce(lo.Flatten(lo.Values(results)),
		func(acc int64, r meilisearch.SearchResponse, i int) int64 {
			return acc + r.ProcessingTimeMs
		}, int64(0))
	// do a floor division of hitsPerPage / TotalHits to get the total pages
	processed.TotalPages = (processed.TotalHits + hitsPerPage - 1) / hitsPerPage
	processed.Page = 1

	for i, category := range processed.Categories {
		// TODO: remove _rev and _id from the response
		processed.Data = append(processed.Data, results[category][i].Hits)
	}

	return processed
}
