package contract

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"
	_ "github.com/mattn/go-sqlite3"
)

/* Yield ... Contains a 32 size byte slice recipient, and a uint64 value */
type Yield struct {
	Recipient []byte //32 byte hash value
	Value     uint64
}

/* Make Yield ... Recipient will be SHA-256 hashed */
func MakeYield(recipient *ecdsa.PublicKey, value uint64) Yield {
	return Yield{Recipient: block.HashSHA256(keys.EncodePublicKey(recipient)), Value: value}
}

/* Inserts Yield into database */
func InsertYield(y Yield, database string, blockHeight uint32, contractHash []byte, yieldIndex uint8) error {
	dbConn, err := sql.Open("sqlite3", database)
	if err != nil {
		log.Fatal(err)
		return err
	}
	sqlStatement := `INSERT INTO uy VALUES ($1, $2, $3, $4, $5);`
	_, err2 := dbConn.Exec(sqlStatement, blockHeight, hex.EncodeToString(contractHash), yieldIndex, hex.EncodeToString(y.Recipient), y.Value)
	if err2 != nil {
		fmt.Println("Failed to insert")
		return err2
	}

	dbConn.Close()
	return nil
}

/* Serialize ... serialies the yield */
func (y *Yield) Serialize() []byte {
	s := make([]byte, 40) //32 bytes for hash and 8 bytes for value
	copy(s[0:32], y.Recipient)
	binary.LittleEndian.PutUint64(s[32:40], y.Value)
	return s
}

/* DeserializeYield ... deserializes the yield */
func DeserializeYield(b []byte) Yield {
	recipient := make([]byte, 32)
	copy(recipient, b[0:32])
	value := binary.LittleEndian.Uint64(b[32:40])
	return Yield{Recipient: recipient, Value: value}
}

/*
Contains the contract hash of the claimed yield,
the block index containing the contract of the claimed yield,
the index of the claimed yield in the contract,
and the public key of the claimant
*/
type Claim struct {
	PreviousContractHash []byte
	BlockIndex           uint64
	YieldIndex           uint32
	PublicKey            ecdsa.PublicKey
}

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

/* Serialize ... serialies the claim */
func (y *Claim) Serialize() []byte {
	encodedPubKey := keys.EncodePublicKey(&y.PublicKey)
	keyLen := uint16(len(encodedPubKey))
	len := 32 + 8 + 4 + 2 + len(encodedPubKey)
	s := make([]byte, len)
	copy(s[0:32], y.PreviousContractHash)
	binary.LittleEndian.PutUint64(s[32:40], y.BlockIndex)
	binary.LittleEndian.PutUint32(s[40:44], y.YieldIndex)
	binary.LittleEndian.PutUint16(s[44:46], keyLen)
	copy(s[46:46+keyLen], encodedPubKey)
	return s
}

/* DeserializeYield ... deserializes the claim */
func DeserializeClaim(b []byte) Claim {
	c := Claim{}
	c.PreviousContractHash = make([]byte, 32)
	copy(c.PreviousContractHash, b[0:32])
	c.BlockIndex = binary.LittleEndian.Uint64(b[32:40])
	c.YieldIndex = binary.LittleEndian.Uint32(b[40:44])

	keylen := binary.LittleEndian.Uint16(b[44:46])
	decodedPubKey := make([]byte, keylen)
	copy(decodedPubKey, b[46:46+keylen])
	c.PublicKey = *keys.DecodePublicKey(decodedPubKey)
	return c
}
