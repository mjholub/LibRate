package crypt

import (
	"filippo.io/age"
	"github.com/gofrs/uuid/v5"
)

// Keys holds the information used for volatile identification of a particular user
type Keys struct {
	Identifier uuid.UUID
	Private    *age.X25519Identity
	Public     *age.X25519Recipient
}

// @function GenerateKeys
// @description Generates a (volatile) public and private key pair to be used for intermediary encryption in the frontend
// @returns Keys, error
// NOTE: should probably also expect some persistent key based identification
// to be passed as a parameter to make impersonation harder
func GenerateKeys() (k Keys, e error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return Keys{}, err
	}

	recipient := identity.Recipient()

	// generate a highly randomized identifier, so that we can later look up, if a given private key
	// matches it's public key. Without a named identifier, such check would be overly complex/impossible
	id, err := uuid.NewV7()
	if err != nil {
		return Keys{}, err
	}

	return Keys{
		Identifier: id,
		Private:    identity,
		Public:     recipient,
	}, nil
}
