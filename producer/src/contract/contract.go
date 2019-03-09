package contract

import (
	"crypto/ecdsa"
	"errors"
)

/*Contains a 32 size byte slice recipient,
and a uint64 value */
type Yield struct{}

/*
Consists of a previous contract hash,
a claimed yield index,
and a public key
*/
type Claim struct{}

/*
Consists of a version number,
an array of Claims,
and an array of Yields
*/
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

func MakeContract(version uint32, database string, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64) (Contract, error) {
	// Assign version
	// Get list of claims
	/*
		Should have a SumSoFar variable
		While the SumSoFar is less than the value, keep
		Making claims
		If you run out of claims to make before
		SumSoFar >= value, return an error
		and restore all the removed yields in the database
		If SumSoFar manages to exceed value,
		Take the difference and make a Yield that goes to yourself.
		That will be the change you get back
	*/
	// Get list of yields
	/*
		At minimum one yield for the recipient passed in as parameter
		Possible change you get back
	*/
	return Contract{}, errors.New("Incomplete function")
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
