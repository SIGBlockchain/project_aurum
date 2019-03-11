package contract

import (
	"crypto/ecdsa"
)

/*Contains a 32 size byte slice recipient,
and a uint64 value */
type Yield struct{}

// Recipient will be SHA-256 hashed
func MakeYield(recipient ecdsa.PublicKey, value uint64) Yield {
	return Yield{}
}

// func InsertYield(y Yield, database string, blockHeight uint32, string contractHash, uint64 ) {
// 	// Open database connection,
// 	// Insert into table:
// 	// height of the block the yield is located in
// 	// hash of the contract the yield is located in
// 	//
// }

func (y *Yield) Serialize() []byte {
	return []byte{}
}

func DeserializeYield(b []byte) Yield {
	return Yield{}
}
