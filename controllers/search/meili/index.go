package meili

import (
	"context"

	"github.com/meilisearch/meilisearch-go"

	"codeberg.org/mjh/LibRate/cfg"
	searchdb "codeberg.org/mjh/LibRate/models/search"
)

type Meili struct {
	searchdb *searchdb.Storage
	client   *meilisearch.Client
}

func Connect(conf *cfg.Search) (*Meili, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.MeiliHost,
		APIKey: conf.MeiliKey,
	})

	return &Meili{client: client}, nil
}

func (m *Meili) CreateAllIndexes(ctx context.Context) error {
	genres, err := m.searchdb.ReadGenres(ctx)
	if err != nil {
		return err
	}
	_, err = m.client.Index("genres").AddDocuments(genres)
	if err != nil {
		return err
	}
	return nil
}

func (m *Meili) HandleSearch()
