package meili

import (
	"context"
	"fmt"
	"sync"

	"codeberg.org/mjh/LibRate/controllers/search/target"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

func (s *Service) CreateAllIndexes(ctx context.Context) error {
	docs, err := s.searchdb.ReadAll(ctx)
	if err != nil {
		return err
	}
	docData := map[string][]any{
		"genres":  lo.ToAnySlice(docs.Genres),
		"members": lo.ToAnySlice(docs.Members),
		"studios": lo.ToAnySlice(docs.Studios),
		"ratings": lo.ToAnySlice(docs.Ratings),
		"artists": lo.ToAnySlice(docs.Artists),
		"media":   lo.ToAnySlice(docs.Media),
	}

	errorCh := make(chan error, len(docData))
	var wg sync.WaitGroup

	s.log.Info().Msgf("creating %d indexes", len(docData))
	wg.Add(len(docData))
	for name, data := range docData {
		go func(name string, data []any) {
			defer wg.Done()
			if _, err := s.client.Index(name).AddDocuments(data); err != nil {
				errorCh <- fmt.Errorf("error building index %s: %w", name, err)
			}
		}(name, data)
	}
	wg.Wait()

	close(errorCh)
	errorSlice := make([]error, 0, len(docData))

	s.log.Debug().Msg("performing final error check")
	for e := range errorCh {
		errorSlice = append(errorSlice, e)
		// combine all errors into one
		switch len(errorSlice) {
		case 0:
			return nil
		case 1:
			return errorSlice[0]
		default:
			return fmt.Errorf("errors building indexes: %v", errorSlice)
		}
	}

	return nil
}

func (s *Service) RunQuery(opts *Options) (
	categorisedResult map[string][]meilisearch.SearchResponse, err error) {
	attributesToCrop := []string{"_rev", "_id"}
	categoriesList := lo.Map(opts.Categories, func(c target.Category, _ int) string {
		return c.String()
	})

	var indexes []string
	// run a query through all indexes
	if lo.Contains(categoriesList, "union") {
		indexes = []string{"genres", "members", "studios", "ratings", "artists", "media"}
	} else {
		indexes = categoriesList
	}

	requests := make([]meilisearch.SearchRequest, len(indexes))
	for i, index := range indexes {
		request := meilisearch.SearchRequest{
			HitsPerPage:      opts.PageSize,
			Query:            opts.Query,
			AttributesToCrop: attributesToCrop,
			IndexUID:         index,
		}
		requests[i] = request
	}

	multiReq := &meilisearch.MultiSearchRequest{Queries: requests}

	multiRes, err := s.client.MultiSearch(multiReq)
	if err != nil {
		return nil, fmt.Errorf("error running multi search: %w", err)
	}

	categorisedResult = lop.GroupBy(multiRes.Results, func(result meilisearch.SearchResponse) string {
		return result.IndexUID
	})

	cleanedResult := lo.OmitBy(categorisedResult, func(k string, v []meilisearch.SearchResponse) bool {
		nonEmpty := lo.Filter(v, func(r meilisearch.SearchResponse, _ int) bool {
			return r.Hits != nil && len(r.Hits) > 0
		})
		return len(nonEmpty) == 0
	})

	return cleanedResult, nil
}
