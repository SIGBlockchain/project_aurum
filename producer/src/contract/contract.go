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

/*
Contains the contract hash of the claimed yield,
the block index containing the contract of the claimed yield,
the index of the claimed yield in the contract,
and the public key of the claimant
*/
type Claim struct{}

/*
Should scan the Unclaimed Yield Pool for a yield

Prioritize yields that are closest to the value parameter,
ie. MIN(abs(value - yieldValue0), abs(value - yieldValue1), ... abs(value - yieldValueN))

Case 1:
If the claimed yield is less than or equal to the value, return
the claim and a nil for the error

Case 2:
If the claimed yield exceeds the value, return
the claim as usual and a custom error struct with
the difference as a uint64 field called "Change"

Case 3:
If there are no yields left in the Pool,
return an empty Claim struct and a custom error struct
that simply states there are insufficient funds
*/
func MakeClaim(database string, claimant ecdsa.PublicKey, value uint64) (Claim, error) {
	return Claim{}, errors.New("Incomplete function")
}
