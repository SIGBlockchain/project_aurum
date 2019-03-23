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
func InsertYield(y Yield, database string, blockHeight uint64, contractHash []byte, yieldIndex uint8) error {
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

/*Serialize serializes the yield */
func (y *Yield) Serialize() []byte {
	s := make([]byte, 40) //32 bytes for hash and 8 bytes for value
	copy(s[0:32], y.Recipient)
	binary.LittleEndian.PutUint64(s[32:40], y.Value)
	return s
}

/*DeserializeYield deserializes the yield */
func DeserializeYield(b []byte) Yield {
	recipient := make([]byte, 32)
	copy(recipient, b[0:32])
	value := binary.LittleEndian.Uint64(b[32:40])
	return Yield{Recipient: recipient, Value: value}
}

/*Claim contains information about which yield it is claiming*/
type Claim struct {
	PreviousContractHash []byte
	BlockIndex           uint64
	YieldIndex           uint8
	PublicKey            ecdsa.PublicKey
}

/*MakeClaim will query the database to find the best yield to claim */
func MakeClaim(database string, claimant ecdsa.PublicKey, value uint64) (Claim, error) {
	claimantHexHash := hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&claimant)))
	//open datase first
	dbConn, err := sql.Open("sqlite3", database)
	if err != nil {
		log.Printf("Error! unable to open databse: %v\n", err)
		return Claim{}, err
	}
	defer dbConn.Close()

	var dbHeight uint64
	var dbContract string
	var dbIndex uint8
	var dbHolder string
	var dbValue uint64

	//first query the db for any yeilds equal or above
	sqlQuery := `SELECT * FROM uy WHERE value >= $1 AND holder = $2 ORDER BY value ASC LIMIT 1;`
	rows, err := dbConn.Query(sqlQuery, value, claimantHexHash)
	if err != nil {
		log.Printf("Error! Unable to execte query: %v\n", err)
		return Claim{}, err
	}

	if !rows.Next() {
		//do another query if the first one wasn't successful
		sqlQuery = `SELECT * FROM uy WHERE holder = $1 ORDER BY value DESC LIMIT 1;`
		rows, err = dbConn.Query(sqlQuery, claimantHexHash)
		if err != nil {
			log.Printf("Error! Unable to execte query: %v\n", err)
			return Claim{}, err
		}
		if !rows.Next() {
			//no results found in either query, no yields to claim
			return Claim{}, errors.New("No yields to claim")
		}
	}

	if err = rows.Scan(&dbHeight, &dbContract, &dbIndex, &dbHolder, &dbValue); err != nil {
		return Claim{}, err
	}
	c := Claim{BlockIndex: dbHeight, YieldIndex: dbIndex, PublicKey: claimant}
	var rtnErr error //nil by default
	//determine if there is change
	if dbValue > value {
		rtnErr = ChangeError{Change: dbValue - value}
	} else if dbValue < value {
		rtnErr = DeficitError{Deficit: value - dbValue}
	}
	c.PreviousContractHash, _ = hex.DecodeString(dbContract)

	rows.Close()
	return c, rtnErr
}

/* Serialize ... serialies the claim */
func (y *Claim) Serialize() []byte {
	encodedPubKey := keys.EncodePublicKey(&y.PublicKey)
	keyLen := uint16(len(encodedPubKey))
	len := 32 + 8 + 1 + 2 + len(encodedPubKey)
	s := make([]byte, len)
	copy(s[0:32], y.PreviousContractHash)
	binary.LittleEndian.PutUint64(s[32:40], y.BlockIndex)
	s[40] = y.YieldIndex
	binary.LittleEndian.PutUint16(s[41:43], keyLen)
	copy(s[43:43+keyLen], encodedPubKey)
	return s
}

/* DeserializeYield ... deserializes the claim */
func DeserializeClaim(b []byte) Claim {
	c := Claim{}
	c.PreviousContractHash = make([]byte, 32)
	copy(c.PreviousContractHash, b[0:32])
	c.BlockIndex = binary.LittleEndian.Uint64(b[32:40])
	c.YieldIndex = b[40]

	keylen := binary.LittleEndian.Uint16(b[41:43])
	decodedPubKey := make([]byte, keylen)
	copy(decodedPubKey, b[43:43+keylen])
	c.PublicKey = *keys.DecodePublicKey(decodedPubKey)
	return c
}
