package search

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/rs/zerolog"

	searchdb "codeberg.org/mjh/LibRate/models/search"
)

func (s *Service) CreateIndex(
	ctx context.Context,
	runtimeStats bool,
	path string,
) error {
	idx, err := buildIndex(path)
	if err != nil {
		return fmt.Errorf("error creating index '%q': %v", path, err)
	}

	if runtimeStats {
		f, err := os.Create("cpu-search.out")
		if err != nil {
			return fmt.Errorf("error creating CPU profile file: %w", err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		ff, err := os.Create("mem-search.out")
		if err != nil {
			return fmt.Errorf("error creating memory profile file: %w", err)
		}
		defer func() {
			pprof.WriteHeapProfile(ff)
			ff.Close()
		}()
	}

	errorCh := make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := indexSite(ctx, idx, s.storage, s.log)
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

func indexSite(
	ctx context.Context,
	idx bleve.Index,
	storage *searchdb.Storage,
	log *zerolog.Logger) error {
	batch := idx.NewBatch()

	var docCount, batchCount int
	start := time.Now()

	data, err := storage.ReadAll(ctx)
	if err != nil {
		return err
	}
	docs, err := searchdb.ToBleveDocument(data, log)
	if err != nil {
		return fmt.Errorf("error converting data to bleve documents: %v", err)
	}
	for i := range docs {
		err = batch.Index(fmt.Sprintf("%s-%s", docs[i].Type, docs[i].ID),
			searchdb.AnonymousDocument{
				Type:   docs[i].Type,
				Fields: docs[i].Fields,
				Data:   docs[i].Data,
			})
		if err != nil {
			return err
		}
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

// TODO: add building partial indices

func buildIndex(path string) (bleve.Index, error) {
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultType = "media"

	genresMapping := buildGenresMapping(textFieldMapping, keywordMapping)
	indexMapping.AddDocumentMapping("genres", genresMapping)
	artistsMapping := buildArtistsMapping(textFieldMapping, keywordMapping)
	mediaMapping := buildMediaMapping(textFieldMapping, keywordMapping)
	indexMapping.AddDocumentMapping("media", mediaMapping)
	indexMapping.AddDocumentMapping("artists", artistsMapping)
	usersMapping := buildUsersMapping(textFieldMapping)
	reviewsMapping := buildReviewsMapping(textFieldMapping, mediaMapping, usersMapping)
	indexMapping.AddDocumentMapping("reviews", reviewsMapping)
	indexMapping.AddDocumentMapping("users", usersMapping)

	return bleve.New(path, indexMapping)
}

func buildGenresMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()
	mapping.StructTagKey = "genres"
	name := bleve.NewTextFieldMapping()
	name.Analyzer = en.AnalyzerName
	name.IncludeInAll = true

	mapping.AddFieldMappingsAt("name", name)
	kinds := bleve.NewKeywordFieldMapping()
	mapping.AddFieldMappingsAt("kinds", kinds)

	descriptions := bleve.NewDocumentMapping()
	descriptions.StructTagKey = "descriptions"
	descriptions.AddFieldMappingsAt("description", textFieldMapping)
	descriptions.AddFieldMappingsAt("language", keywordMapping)
	mapping.AddSubDocumentMapping("descriptions", descriptions)

	return mapping
}

func buildReviewsMapping(textFieldMapping *mapping.FieldMapping, mediaMapping, userMapping *mapping.DocumentMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()
	mapping.StructTagKey = "reviews"

	mapping.AddSubDocumentMapping("media", mediaMapping)
	mapping.AddSubDocumentMapping("user", userMapping)
	mapping.AddFieldMappingsAt("topic", textFieldMapping)
	mapping.AddFieldMappingsAt("body", textFieldMapping)
	added := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("added", added)
	mapping.AddFieldMappingsAt("modified", bleve.NewDateTimeFieldMapping())
	// TODO: uncomment when post/review interactions are added
	/*
		favoriteCount := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("favoriteCount", favoriteCount)
		reblogCount := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("reblogCount", reblogCount)
	*/

	return mapping
}

func buildMediaMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()
	mapping.StructTagKey = "media"

	mapping.AddFieldMappingsAt("kind", keywordMapping)
	mapping.AddFieldMappingsAt("title", textFieldMapping)
	//mapping.AddSubDocumentMapping("artists", artists)
	//mapping.AddSubDocumentMapping("genres", genres)
	//mapping.AddFieldMappingsAt("language", keywordMapping)
	created := bleve.NewDateTimeFieldMapping()
	added := bleve.NewDateTimeFieldMapping()
	modified := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("created", created)
	mapping.AddFieldMappingsAt("added", added)
	mapping.AddFieldMappingsAt("modified", modified)

	return mapping
}

func buildUsersMapping(textFieldMapping *mapping.FieldMapping) (res *mapping.DocumentMapping) {
	mapping := bleve.NewDocumentMapping()
	mapping.StructTagKey = "members"

	mapping.AddFieldMappingsAt("webfinger", textFieldMapping)
	//mapping.AddFieldMappingsAt("instance", textFieldMapping)
	//localAccounts := bleve.NewBooleanFieldMapping()
	//mapping.AddFieldMappingsAt("local", localAccounts)

	mapping.AddFieldMappingsAt("display_name", textFieldMapping)
	mapping.AddFieldMappingsAt("bio", textFieldMapping)

	return mapping
}

func buildArtistsMapping(textFieldMapping, keywordMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	mapping := bleve.NewDocumentMapping()

	mapping.AddFieldMappingsAt("name", textFieldMapping)
	mapping.AddFieldMappingsAt("nicknames", keywordMapping)
	//mapping.AddFieldMappingsAt("roles", keywordMapping)
	//mapping.AddFieldMappingsAt("country", keywordMapping)
	mapping.AddFieldMappingsAt("bio", textFieldMapping)
	added := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("added", added)
	modified := bleve.NewDateTimeFieldMapping()
	mapping.AddFieldMappingsAt("modified", modified)
	//active := bleve.NewBooleanFieldMapping()
	//mapping.AddFieldMappingsAt("active", active)

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
