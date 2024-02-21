package meili

import (
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
)

type SearchResponse struct {
	Categories     []string                 `json:"categories"`
	TotalHits      int64                    `json:"totalHits"`
	ProcessingTime int64                    `json:"processingTime"`
	Page           int64                    `json:"page"`
	TotalPages     int64                    `json:"totalPages"`
	HitsPerPage    int64                    `json:"hitsPerPage"`
	Data           []map[string]interface{} `json:"data"`
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

	rawResults := make([]map[string]interface{}, 0)

	for i, category := range processed.Categories {
		for _, hit := range results[category][i].Hits {
			for k, v := range hit.(map[string]interface{}) {
				results := make(map[string]interface{})
				results[k] = v

				rawResults = append(rawResults, results)
			}
		}
	}
	processed.Data = cleanHitsData(rawResults)

	return processed
}

// FIXME: the first key gets processed properly, but for remaining names (keys)
// the auxiliary data is incorrectly combined.
// Example:
// genreA: (it's proper description)
// genreB, genreC, genreD: description of genreB
func cleanHitsData(data []map[string]interface{}) []map[string]interface{} {
	cleanedData := make([]map[string]interface{}, 0)
	seen := make(map[string]bool)

	for _, item := range data {
		if formatted, ok := item["_formatted"].(map[string]interface{}); ok {
			// Unnest the value of the _formatted key
			for k, v := range formatted {
				item[k] = v
			}
			delete(item, "_formatted")
		}

		// Remove _id and _rev keys
		delete(item, "_id")
		delete(item, "_rev")

		// Deduplicate
		id, ok := item["name"].(string)
		if !ok {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		cleanedData = append(cleanedData, item)
	}

	return cleanedData
}
