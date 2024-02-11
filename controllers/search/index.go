package search

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/rs/zerolog"

	searchdb "codeberg.org/mjh/LibRate/models/search"
)

func CreateIndex(ctx context.Context, path string, storage *searchdb.Storage, log *zerolog.Logger) error {
	idx, _ := buildIndexMapping()

	fullIndex, err := bleve.New(path, idx)
	if err != nil {
		return fmt.Errorf("error creating index '%q': %v", path, err)
	}

	errorCh := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := indexSite(ctx, fullIndex, storage, log)
		if err != nil {
			errorCh <- fmt.Errorf("error indexing site: %v", err)
			return
		}
	}()
	wg.Wait()
	close(errorCh)

	if err, ok := <-errorCh; ok {
		return err
	}

	return nil
}

// TODO: implement
func (s *Service) PartialIndex(ctx context.Context, path string) error {
	/*
			indexMapping := bleve.NewIndexMapping()
		textFieldMapping := bleve.NewTextFieldMapping()
			textFieldMapping.Analyzer = en.AnalyzerName

			keywordMapping := bleve.NewTextFieldMapping()
			keywordMapping.Analyzer = keyword.Name
	*/
	return nil
}

func indexSite(ctx context.Context, idx bleve.Index, storage *searchdb.Storage, log *zerolog.Logger) error {
	batch := idx.NewBatch()

	var docCount, batchCount int
	start := time.Now()

	for i := range searchdb.AllTargets {
		storage.ReadAll(ctx, searchdb.AllTargets[i])
		batch.Index(searchdb.AllTargets[i].String(), searchdb.AllTargets[i])
		batchCount++

		if batchCount >= 100 {
			if err := idx.Batch(batch); err != nil {
				return fmt.Errorf("error indexing batch: %v", err)
			}

			batch = idx.NewBatch()
			batchCount = 0
		}
	}
	// flush the last batch
	if batchCount > 0 {
		if err := idx.Batch(batch); err != nil {
			return fmt.Errorf("error indexing last batch: %v", err)
		}
	}

	docCount++
	indexTime := time.Since(start)
	indexDuration := float64(indexTime) / float64(time.Second)
	perDoc := float64(indexTime) / float64(docCount)
	log.Info().Msgf("Indexed %d documents in %v (%.2f docs/sec, %.2f ms/doc)",
		docCount, indexTime, float64(docCount)/indexDuration, float64(perDoc)/float64(time.Millisecond))

	return nil
}

func buildIndexMapping() (mapping.IndexMapping, error) {
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	indexMapping := bleve.NewIndexMapping()

	genresMapping := buildGenresMapping(textFieldMapping, keywordMapping)
	indexMapping.AddDocumentMapping("genres", genresMapping)
	artistsMapping := buildArtistsMapping(textFieldMapping, keywordMapping)
	mediaMapping := buildMediaMapping(textFieldMapping, keywordMapping, artistsMapping, genresMapping)
	indexMapping.AddDocumentMapping("media", mediaMapping)
	indexMapping.AddDocumentMapping("artists", artistsMapping)
	usersMapping := buildUsersMapping(textFieldMapping)
	reviewsMapping := buildReviewsMapping(textFieldMapping, mediaMapping, usersMapping)
	indexMapping.AddDocumentMapping("reviews", reviewsMapping)
	indexMapping.AddDocumentMapping("users", usersMapping)

	return indexMapping, nil
}

func buildGenresMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()
	mapping.AddFieldMappingsAt("name", textFieldMapping)
	mapping.AddFieldMappingsAt("kinds", keywordMapping)
	childGenresMapping := bleve.NewDocumentMapping()

	mapping.AddSubDocumentMapping("children", childGenresMapping)

	description := bleve.NewTextFieldMapping()
	mapping.AddFieldMappingsAt("description", description)
	mapping.AddFieldMappingsAt("language", description)

	characteristics := bleve.NewTextFieldMapping()

	mapping.AddFieldMappingsAt("characteristics", characteristics)

	return mapping
}

func buildReviewsMapping(textFieldMapping *mapping.FieldMapping, mediaMapping, userMapping *mapping.DocumentMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()

	mapping.AddSubDocumentMapping("media", mediaMapping)
	mapping.AddSubDocumentMapping("user", userMapping)
	mapping.AddFieldMappingsAt("topic", textFieldMapping)
	mapping.AddFieldMappingsAt("comment", textFieldMapping)
	date := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("date", date)
	favoriteCount := bleve.NewNumericFieldMapping()
	mapping.AddFieldMappingsAt("favoriteCount", favoriteCount)
	reblogCount := bleve.NewNumericFieldMapping()
	mapping.AddFieldMappingsAt("reblogCount", reblogCount)

	return mapping
}

func buildMediaMapping(textFieldMapping, keywordMapping *mapping.FieldMapping, artists, genres *mapping.DocumentMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()

	mapping.AddFieldMappingsAt("kind", keywordMapping)
	mapping.AddFieldMappingsAt("title", textFieldMapping)
	mapping.AddSubDocumentMapping("artists", artists)
	mapping.AddSubDocumentMapping("genres", genres)
	mapping.AddFieldMappingsAt("language", keywordMapping)
	released := bleve.NewDateTimeFieldMapping()
	added := bleve.NewDateTimeFieldMapping()
	modified := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("released", released)
	mapping.AddFieldMappingsAt("added", added)
	mapping.AddFieldMappingsAt("modified", modified)

	return mapping
}

func buildUsersMapping(textFieldMapping *mapping.FieldMapping) (res *mapping.DocumentMapping) {
	mapping := bleve.NewDocumentMapping()

	mapping.AddFieldMappingsAt("webfinger", textFieldMapping)
	mapping.AddFieldMappingsAt("instance", textFieldMapping)
	localAccounts := bleve.NewBooleanFieldMapping()
	mapping.AddFieldMappingsAt("local", localAccounts)

	mapping.AddFieldMappingsAt("displayName", textFieldMapping)
	mapping.AddFieldMappingsAt("bio", textFieldMapping)

	return mapping
}

func buildArtistsMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()

	mapping.AddFieldMappingsAt("artist_name", textFieldMapping)
	mapping.AddFieldMappingsAt("roles", keywordMapping)
	mapping.AddFieldMappingsAt("country", keywordMapping)
	mapping.AddFieldMappingsAt("bio", textFieldMapping)
	added := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("added", added)
	modified := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("modified", modified)
	active := bleve.NewBooleanFieldMapping()
	mapping.AddFieldMappingsAt("active", active)
	mapping.AddSubDocumentMapping("associatedArtists", mapping)

	return mapping
}

// TODO: implement when posts are added
func buildPostsMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	return nil
}

// TODO: implement when tags are added
func buildTagsMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	return nil
}

// TODO: implement when groups are added
func buildGroupsMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	return nil
}
