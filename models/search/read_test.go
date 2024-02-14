// this file also covers connect.go and enum.go
package searchdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
)

func TestReadAll(t *testing.T) {
	conf := cfg.TestConfig

	log := zerolog.Nop()
	storage, err := Connect(&conf.CouchDB, &log)
	require.NoErrorf(t, err, "failed to connect to search database: %s", err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := storage.ReadAll(ctx)
	fmt.Printf("data: %v\n", data)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data)
}

func TestReadGenres(t *testing.T) {
	conf := cfg.TestConfig

	log := zerolog.Nop()
	storage, err := Connect(&conf.CouchDB, &log)
	require.NoErrorf(t, err, "failed to connect to search database: %s", err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := storage.ReadGenres(ctx)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 100)
	sampleGenre := data[1]
	fmt.Printf("%+v", sampleGenre)
	assert.NotEmpty(t, sampleGenre.ID)
	assert.NotEmpty(t, sampleGenre.Rev)
	// must call require to avoid index out of range error if no data
	require.NotEmpty(t, sampleGenre.Name)
	assert.Contains(t, []string{"music", "film", "book", "game"}, sampleGenre.Kinds[0])
	assert.Greater(t, len(sampleGenre.Descriptions), 0)
}

func TestToBleveDocument(t *testing.T) {
	testData := CombinedData{
		Genres: []Genre{
			{
				ID:    "genre_1",
				Rev:   "1-1234567890abcdef",
				Name:  "Vaporwave",
				Kinds: []string{"music", "film"},
				Descriptions: [][]GenreDescription{
					{
						{
							Language:    "en",
							Description: "A genre of music and film that is characterized by a nostalgic or surrealist fascination with retrofuturism, the 1980s, and 1990s, and postmodernism.",
						},
						{
							Language:    "de",
							Description: "Ein Genre von Musik und Film, das durch eine nostalgische oder surrealistische Faszination für Retrofuturismus, die 1980er und 1990er Jahre und Postmodernismus gekennzeichnet ist.",
						},
					},
				},
			},
		},
		Artists: []Artist{
			{
				ID:        "artist_1",
				Rev:       "1-1234567890abcdef",
				Name:      "Macintosh Plus",
				Nicknames: []string{"Vektroid", "New Dreams Ltd."},
				Bio:       "Macintosh Plus is the pseudonym of producer Ramona Andra Xavier, also known as Vektroid. She is best known for her 2011 album Floral Shoppe, which is considered a defining work of the vaporwave genre.",
				Added:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				Modified:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		Members: []Member{
			{
				ID:          "member_1",
				Rev:         "1-1234567890abcdef",
				Bio:         "lorem ipsum",
				Webfinger:   "lain@navi.wired",
				DisplayName: "Lain Iwakura",
			},
		},
	}

	docs, err := ToBleveDocument(&testData)
	require.NoError(t, err)
	assert.Equal(t, 3, len(docs))
	assert.Contains(t, docs, "genre_1")
}
