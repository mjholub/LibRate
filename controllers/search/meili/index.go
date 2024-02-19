package meili

import (
	"context"
	"fmt"
	"sync"

	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
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

	s.log.Info().Msg("creating union index")
	// build the summary (aka "union") index
	if _, err = s.client.Index("union").AddDocuments(docs); err != nil {
		return fmt.Errorf("error building index union: %w", err)
	}

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

func (s *Service) RunQuery(opts *Options) (res *meilisearch.SearchResponse, err error) {
	req := meilisearch.SearchRequest{
		HitsPerPage: opts.PageSize,
	}

	return s.client.Index(opts.Categories[0].String()).Search(opts.Query, &req)
}
