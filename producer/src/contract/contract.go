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
func InsertYield(y Yield, database string, blockHeight uint64, contractHash []byte, yieldIndex uint16) error {
	dbConn, err := sql.Open("sqlite3", database)
	defer dbConn.Close()
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
	YieldIndex           uint16
	PublicKey            ecdsa.PublicKey
}

func MakeClaim(database string, claimant ecdsa.PublicKey, value uint64) (Claim, error) {
	claimantHexHash := hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&claimant)))
	//open datase first
	dbConn, err := sql.Open("sqlite3", database)
	if err != nil {
		log.Printf("Error! unable to open databse: %v\n", err)
		return Claim{}, err
	}
	defer dbConn.Close()

	//first query the db for any yeilds equal or above
	sqlQuery := `SELECT * FROM uy WHERE value >= $1 AND holder = $2 ORDER BY value ASC LIMIT 1;`
	rows, err2 := dbConn.Query(sqlQuery, value, claimantHexHash)
	if err2 != nil {
		log.Printf("Error! Unable to execte query: %v\n", err2)
	}

	var dbHeight uint64
	var dbContract string
	var dbIndex uint16
	var dbHolder string
	var dbValue uint64

	if rows.Next() {
		if err3 := rows.Scan(&dbHeight, &dbContract, &dbIndex, &dbHolder, &dbValue); err3 != nil {
			log.Fatal(err3)
		}
		c := Claim{}
		var rtnErr error //nil by default
		//determine if there is change
		if dbValue > value {
			fmt.Printf("make a change error")
			rtnErr = ChangeError{Change: dbValue - value}
		}
		c.PreviousContractHash, _ = hex.DecodeString(dbContract)
		c.BlockIndex = dbHeight
		c.YieldIndex = dbIndex
		c.PublicKey = claimant

		rows.Close()
		return c, rtnErr
	}

	//else we need to select a yield that has less value
	return Claim{}, errors.New("Incomplete function")
}

/* Serialize ... serialies the claim */
func (y *Claim) Serialize() []byte {
	encodedPubKey := keys.EncodePublicKey(&y.PublicKey)
	keyLen := uint16(len(encodedPubKey))
	len := 32 + 8 + 2 + 2 + len(encodedPubKey)
	s := make([]byte, len)
	copy(s[0:32], y.PreviousContractHash)
	binary.LittleEndian.PutUint64(s[32:40], y.BlockIndex)
	binary.LittleEndian.PutUint16(s[40:42], y.YieldIndex)
	binary.LittleEndian.PutUint16(s[42:44], keyLen)
	copy(s[44:44+keyLen], encodedPubKey)
	return s
}

/* DeserializeYield ... deserializes the claim */
func DeserializeClaim(b []byte) Claim {
	c := Claim{}
	c.PreviousContractHash = make([]byte, 32)
	copy(c.PreviousContractHash, b[0:32])
	c.BlockIndex = binary.LittleEndian.Uint64(b[32:40])
	c.YieldIndex = binary.LittleEndian.Uint16(b[40:42])

	keylen := binary.LittleEndian.Uint16(b[42:44])
	decodedPubKey := make([]byte, keylen)
	copy(decodedPubKey, b[44:44+keylen])
	c.PublicKey = *keys.DecodePublicKey(decodedPubKey)
	return c
}
