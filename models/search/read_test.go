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

	data, err := storage.ReadAll(ctx, TargetDB(Genres))
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
