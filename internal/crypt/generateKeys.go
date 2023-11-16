package crypt

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type Keys struct {
	PubKeyBlock  *pem.Block
	PrivKeyBlock *pem.Block
	// Identifier is used to facilitate key retrieval from redis
	Identifier string
}

func GenerateKeys() (k Keys, e error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return k, err
	}

	privKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}

	pubKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privKey.PublicKey),
	}
	// nolint: gosec
	identifier := md5.Sum(pubKeyBlock.Bytes)

	return Keys{
		PubKeyBlock:  pubKeyBlock,
		PrivKeyBlock: privKeyBlock,
		Identifier:   string(identifier[:]),
	}, nil
}
