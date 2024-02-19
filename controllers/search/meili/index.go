package meili

import (
	"context"

	"github.com/meilisearch/meilisearch-go"
)

func (s *Service) CreateAllIndexes(ctx context.Context) error {
	docs, err := s.searchdb.ReadAll(ctx)
	if err != nil {
		return err
	}
	docData := map[string]any{
		"genres":  docs.Genres,
		"members": docs.Members,
		"studios": docs.Studios,
		"ratings": docs.Ratings,
		"artists": docs.Artists,
		"media":   docs.Media,
	}

	for name, data := range docData {
		if _, err := s.client.Index(name).AddDocuments(data); err != nil {
			return err
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
