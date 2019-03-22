package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

/* <Testing adjunct> function for setting up database */
func setUpDB(database string) error {
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return err
	}

	statement, err2 := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS uy ( 
		block_height INTEGER,
		contract_hash TEXT, 
		yield_index INTEGER, 
		holder TEXT, 
		value INTEGER);`)

	if err2 != nil {
		return err2
	}
	statement.Exec()
	conn.Close()
	return nil
}

func tearDown(database string) {
	err := os.Remove(database)
	if err != nil {
		log.Fatal(err)
	}
}

/* <Testing adjunct> generates a random public key */
func generatePubKey() ecdsa.PublicKey {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return privateKey.PublicKey
}

/* Makes a yield and tests if it matches the values */
func TestMakeYield(t *testing.T) {
	testPubKey := generatePubKey()
	encodedPubKey := keys.EncodePublicKey(&testPubKey)
	hashedKey := block.HashSHA256(encodedPubKey)
	expectedRecipient := hashedKey[:]
	expectedValue := uint64(1000000)
	actual := MakeYield(&testPubKey, 1000000)

	if !bytes.Equal(expectedRecipient, actual.Recipient) {
		t.Errorf("Recipients do not match")
	}
	if expectedValue != actual.Value {
		t.Errorf("Values do not match")
	}
}

/* Simple test to make sure insertion is working */
func TestInsertYield(t *testing.T) {
	err := setUpDB("testDB.db")
	if err != nil {
		t.Errorf("Failed to set up database")
		log.Fatal(err)
	}
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(&testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err2 := InsertYield(testYield, "testDB.db", 35, contractHash, 1)
	if err2 != nil {
		t.Errorf("Failed to insert yield")
		log.Fatal(err2)
	}

	dbConn, _ := sql.Open("sqlite3", "testDB.db")
	//check if a row is actually inserted
	sqlQuery := `SELECT * FROM uy;`
	rows, err2 := dbConn.Query(sqlQuery)
	if err2 != nil {
		t.Errorf("failed to query db")
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
		//check if each value is correct
		assert.Equal(t, uint64(35), dbHeight, "The height value in the database is wrong")
		assert.Equal(t, hex.EncodeToString(contractHash), dbContract, "The contract hash in database is wrong")
		assert.Equal(t, uint8(1), dbIndex, "The yield index in database is wrong")
		assert.Equal(t, hex.EncodeToString(testYield.Recipient), dbHolder, "The holder in database is wrong")
		assert.Equal(t, testYield.Value, dbValue, "The value in the database is wrong")
	} else {
		t.Errorf("InsertYield failed to insert row into database")
	}

	rows.Close()
	dbConn.Close()
}

/* Tests both serialization and deserialization */
func TestYieldSerialization(t *testing.T) {
	testPubKey := generatePubKey()
	expected := MakeYield(&testPubKey, 200000)
	serialized := expected.Serialize()
	deserialized := DeserializeYield(serialized)
	if !cmp.Equal(deserialized, expected) {
		t.Errorf("Yield structs do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(reserialized, serialized) {
		t.Errorf("Byte strings do not match")
	}
}

func TestMakeClaimCase1A(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")

	amt := uint64(500)
	pubKey := generatePubKey()
	yield := MakeYield(&pubKey, amt)
	contractHash := block.HashSHA256([]byte("blokchain"))
	blockHeight := uint64(35)
	yieldIndex := uint16(1)

	err := InsertYield(yield, "testDB.db", blockHeight, contractHash, yieldIndex)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}

	testClaim, err := MakeClaim("testDB.db", pubKey, amt)
	if err != nil {
		t.Errorf("Failed to claim yield")
	}
	if !bytes.Equal(testClaim.PreviousContractHash, contractHash) {
		t.Errorf("Contract hashes do not match")
	}
	if testClaim.BlockIndex != 35 {
		t.Errorf("Block indeces do not match")
	}
	if testClaim.YieldIndex != 1 {
		t.Errorf("Yield indeces do not match")
	}
	if !cmp.Equal(testClaim.PublicKey, pubKey) {
		t.Errorf("Public Keys do not match")
	}
	// Case 3
	_, shouldNotBeNil := MakeClaim("testDB.db", pubKey, 1)
	if shouldNotBeNil == nil {
		t.Errorf("Made claim on empty yield pool")
	}
}

func TestClaimSerialization(t *testing.T) {
	testPubKey := generatePubKey()
	prevHash := block.HashSHA256(([]byte("Something")))
	expected := Claim{PreviousContractHash: prevHash, BlockIndex: uint64(2), YieldIndex: uint16(3), PublicKey: testPubKey}
	serialized := expected.Serialize()
	deserialized := DeserializeClaim(serialized)

	//using reflect because cmp does not suppor unexported fields
	if !reflect.DeepEqual(deserialized, expected) {
		t.Errorf("Claim structs do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(reserialized, serialized) {
		t.Errorf("Byte strings do not match")
	}
}
