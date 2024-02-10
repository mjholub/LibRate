package searchdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/cfg"
)

func TestReadAll(t *testing.T) {
	conf := cfg.TestConfig

	storage, err := Connect(&conf.CouchDB)
	require.NoErrorf(t, err, "failed to connect to search database: %s", err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := storage.ReadAll(ctx, TargetDB(Genres))
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data)
}
