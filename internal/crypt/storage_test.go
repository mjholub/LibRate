package crypt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateStorageKey(t *testing.T) {
	key, err := generateStorageKey()
	assert.NoError(t, err)
	assert.Equal(t, 64, len(key), "generated key was meant to equal 64, was %d", len(key))
}

func TestCreateCryptoStorage(t *testing.T) {
	testDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	testFile, err := os.CreateTemp(testDir, "test.db")
	require.NoError(t, err)
	defer func() {
		testFile.Close()
		os.RemoveAll(testDir)
	}()

	conn, err := CreateStorage(testFile.Name(), "test")
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}
