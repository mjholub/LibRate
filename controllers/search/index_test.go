package search

import (
	"testing"

	// "github.com/goccy/go-json"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex(t *testing.T) {
	mapping, err := buildIndex("test.bleve")
	require.NotNil(t, mapping)
	assert.Nil(t, err)
	// TODO: uncomment when the PR to fix the stackoverflow in this function gets merged
	//
	//	err = mapping.Validate()
	//	assert.Nilf(t, err, "invalid mapping")
}

func TestBuildGenresMapping(t *testing.T) {
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = en.AnalyzerName
	keywordMapping := bleve.NewKeywordFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	m := buildGenresMapping(textMapping, keywordMapping)
	require.NotNil(t, m, "expected non-nil document mapping for genres")
	cache := registry.NewCache()
	err := m.Validate(cache)
	require.Nilf(t, err, "error validating document mapping for genres: %v", err)
}

func TestBuildReviewsMapping(t *testing.T) {
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = en.AnalyzerName
	keywordMapping := bleve.NewKeywordFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	cache := registry.NewCache()

	artistsMapping := buildArtistsMapping(textMapping, keywordMapping)
	assert.NotNil(t, artistsMapping)
	err := artistsMapping.Validate(cache)
	require.Nilf(t, err, "error validating document mapping for artists: %v", err)
	genresMapping := buildGenresMapping(textMapping, keywordMapping)
	assert.NotNil(t, genresMapping)
	usersMapping := buildUsersMapping(textMapping)
	assert.NotNil(t, usersMapping)
	err = usersMapping.Validate(cache)
	require.Nilf(t, err, "error validating document mapping for users: %v", err)
	mediaMapping := buildMediaMapping(textMapping, keywordMapping)
	assert.NotNil(t, mediaMapping)
	err = mediaMapping.Validate(cache)
	require.Nilf(t, err, "error validating media cache: %v", err)

	reviewsMapping := buildReviewsMapping(textMapping, mediaMapping, usersMapping)
	require.NotNil(t, reviewsMapping)
	err = reviewsMapping.Validate(cache)
	require.Nilf(t, err, "error validating document mapping for reviews: %v", err)
}

func TestBuildUsersMapping(t *testing.T) {
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = en.AnalyzerName

	cache := registry.NewCache()

	usersMapping := buildUsersMapping(textMapping)
	assert.NotNil(t, usersMapping)
	err := usersMapping.Validate(cache)
	require.Nilf(t, err, "error validating document mapping for users: %v", err)
}

func TestBuildMediaMapping(t *testing.T) {
	cache := registry.NewCache()

	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = en.AnalyzerName
	keywordMapping := bleve.NewKeywordFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	mediaMapping := buildMediaMapping(textMapping, keywordMapping)
	assert.NotNil(t, mediaMapping)
	err := mediaMapping.Validate(cache)
	require.Nilf(t, err, "error validating media cache: %v", err)
}

func TestBuildArtistsMapping(t *testing.T) {

	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = en.AnalyzerName
	keywordMapping := bleve.NewKeywordFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	cache := registry.NewCache()
	m := buildArtistsMapping(textMapping, keywordMapping)
	require.NotNil(t, m)
	err := m.Validate(cache)
	require.NoErrorf(t, err, "error validating document mapping for artists (atomic test): %v", err)
}
