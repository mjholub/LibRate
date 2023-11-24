package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
)

const KeySize = 4096

// chunkBy splits the given slice into chunks of the given size
func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// DEPRECATED: use age instead
// Decrypt decrypts the given data using the given private key
func Decrypt(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	decrypted := make([]byte, 0)

	for range chunkBy[byte](data, priv.N.BitLen()/8) {
		plain, err := rsa.DecryptOAEP(hash, rand.Reader, priv, data, nil)
		if err != nil {
			return []byte{}, err
		}
		decrypted = append(decrypted, plain...)
	}

	return decrypted, nil
}
