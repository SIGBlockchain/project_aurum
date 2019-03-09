package contract

import (
	"crypto/ecdsa"
	"errors"
)

// Contains a 32 size byte slice recipient and a uint64 value
type Yield struct{}

type Claim struct{}

type Contract struct{}

// Recipient will be SHA-256 hashed
func MakeYield(recipient ecdsa.PublicKey, value uint64) Yield {
	return Yield{}
}

// Must remove UY's from database
// Checks database for a yield
func MakeClaim(database string, value uint64) (Claim, error) {
	return Claim{}, errors.New("Incomplete function")
}

func MakeContract(version uint32, database string, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64) Contract {
	// Assign version
	// Get list of claims
	// Get list of yields
	return Contract{}
}

func (c *Contract) Serialize() []byte {
	return []byte{}
}

func (y *Yield) Serialize() []byte {
	return []byte{}
}

func (k *Claim) Serialize() []byte {
	return []byte{}
}

func DeserializeContract(b []byte) Contract {
	return Contract{}
}
func DeserializeClaim(b []byte) Claim {
	return Claim{}
}
func DeserializeYield(b []byte) Yield {
	return Yield{}
}
