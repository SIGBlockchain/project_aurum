package contract

import (
	"crypto/ecdsa"
	"errors"
)

/* Yield ... Contains a 32 size byte slice recipient, and a uint64 value */
type Yield struct{}

/* Make Yield ... Recipient will be SHA-256 hashed */
func MakeYield(recipient ecdsa.PublicKey, value uint64) Yield {
	return Yield{}
}

/* Inserts Yield into database */
func InsertYield(y Yield, database string, blockHeight uint32, contractHash []byte, yieldIndex uint8) error {
	/*
		Open database connection,
		 Insert into table the following:
		 height of the block the yield is located in
		 hash of the contract the yield is located in (HEX STRING FORM)
		 index that the yield is in
		 the yield's public key hash as a string
		 the yield's value
		 Close the database connection
	*/
	return errors.New("Incomplete function")
}

/* Serialize ... serialies the yield */
func (y *Yield) Serialize() []byte {
	return []byte{}
}

/* DeserializeYield ... deserializes the yield */
func DeserializeYield(b []byte) Yield {
	return Yield{}
}

type Claim struct{}

func MakeClaim() Claim {
	return Claim{}
}
