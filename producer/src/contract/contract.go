package contract

import "crypto/ecdsa"

// Contains a 32 size byte slice recipient and a uint64 value
type Yield struct{}

type Claim struct{}

type Contract struct{}

// Recipient will be SHA-256 hashed
func MakeYield(recipient ecdsa.PublicKey, value uint64) Yield {
	return Yield{}
}

func MakeClaim(sender ecdsa.PrivateKey) Claim {
	return Claim{}
}

// Must remove UY's from database
func GetUnclaimedYield(database string, value uint64) Claim {
	return Claim{}
}

func MakeContract(version uint32, database string, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64) Contract {
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
