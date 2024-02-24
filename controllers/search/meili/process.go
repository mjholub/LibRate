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

func (s *Service) processResults(
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
			s.log.Debug().Msgf("hit: %v", hit)
			for k, v := range hit.(map[string]interface{}) {
				results := make(map[string]interface{})
				results[k] = v

				rawResults = append(rawResults, results)
			}
		}
	}
	s.log.Debug().Msgf("rawResults: %v", rawResults)

	processed.Data = s.cleanHitsData(rawResults)

	return processed
}

// FIXME: the first key gets processed properly, but for remaining names (keys)
// the auxiliary data is incorrectly combined.
// Example:
// genreA: (it's proper description)
// genreB, genreC, genreD: description of genreB
func (s *Service) cleanHitsData(data []map[string]interface{}) []map[string]interface{} {
	cleanedData := make([]map[string]interface{}, 0)
	for _, item := range data {
		if formatted, ok := item["_formatted"].(map[string]interface{}); ok {
			// Unnest the value of the _formatted key
			for k, v := range formatted {
				item[k] = v
			}
			delete(item, "_formatted")
		}

		// Remove _rev keys
		delete(item, "_rev")

		s.log.Debug().Msgf("item (after deleting _rev): %v", item)

		cleanedData = append(cleanedData, item)
	}

	s.log.Debug().Msgf("cleanedData: %v", cleanedData)

	/* the following code is a bit off. We need to do
		* something similar to this Clojure code:
		(defn clean-data [data]
	  (->> data
	       (filter #(and (not (empty? %))
	                     (every? #(contains? % %2) [:_id :name :kinds :descriptions])))
	       (group-by :_id)
	       (map (fn [[_id maps]]
	              (apply merge maps)))
	       (map #(dissoc % :_id))))
	*/
	cleanedData = lo.Map(cleanedData, func(item map[string]interface{}, _ int) map[string]interface{} {
		return lo.OmitByValues(item, []any{nil, ""})
	})
	s.log.Debug().Msgf("cleanedData (after omitting nil and empty strings): %v", cleanedData)

	// Group by _id
	groupedData := lo.GroupBy(cleanedData, func(item map[string]interface{}) interface{} {
		return item["_id"]
	})

	s.log.Debug().Msgf("groupedData: %v", groupedData)
	mergedData := lo.Map(lo.Values(groupedData), func(items []map[string]interface{}, _ int) map[string]interface{} {
		return lo.Assign(items...)
	})
	s.log.Debug().Msgf("mergedData: %v", mergedData)

	// Remove _id keys
	final := lo.Map(mergedData, func(item map[string]interface{}, _ int) map[string]interface{} {
		return lo.OmitByKeys(item, []string{"_id"})
	})

	return final
}

func cleanGenresData(data []map[string]interface{}) []map[string]interface{} {
	filtered := lo.Filter(data, func(item map[string]interface{}, _ int) bool {
		return lo.Every(lo.Keys(item), []string{"_id", "name", "kinds", "descriptions"})
	})
	return filtered
}

func cleanMembersData(data []map[string]interface{}) []map[string]interface{} {
	return lo.Filter(data, func(item map[string]interface{}, _ int) bool {
		return lo.Every(lo.Keys(item), []string{"_id", "webfinger"})
	})
}

func cleanStudiosArtistsData(data []map[string]interface{}) []map[string]interface{} {
	return lo.Filter(data, func(item map[string]interface{}, _ int) bool {
		return lo.Every(lo.Keys(item), []string{"_id", "name"})
	})
}

func cleanMediaData(data []map[string]interface{}) []map[string]interface{} {
	return lo.Filter(data, func(item map[string]interface{}, _ int) bool {
		return lo.Every(lo.Keys(item), []string{"_id", "title", "kind"})
	})
}

func cleanRatingData(data []map[string]interface{}) []map[string]interface{} {
	return lo.Filter(data, func(item map[string]interface{}, _ int) bool {
		return lo.Every(lo.Keys(item), []string{"_id", "user", "media_title"})
	})
}
