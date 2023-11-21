package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStorageKey(t *testing.T) {
	key, err := generateStorageKey()
	assert.NoError(t, err)
	assert.Equal(t, 64, len(key), "generated key was meant to equal 64, was %d", len(key))
}

func TestCreateCryptoStorage(t *testing.T) {
	conn, err := CreateCryptoStorage()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}
