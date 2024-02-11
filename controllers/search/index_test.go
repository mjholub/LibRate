package search

import (
	"testing"

	// "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndexMapping(t *testing.T) {
	mapping, err := buildIndexMapping()
	require.NotNil(t, mapping)
	assert.Nil(t, err)
	// TODO: uncomment when the PR to fix the stackoverflow in this function gets merged
	//
	//	err = mapping.Validate()
	//	assert.Nilf(t, err, "invalid mapping")
}
