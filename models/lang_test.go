package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseLookupLangID(t *testing.T) {
	engID, err := ReverseLookupLangID("English")
	assert.NoError(t, err)
	assert.Equal(t, 0, engID)
}
