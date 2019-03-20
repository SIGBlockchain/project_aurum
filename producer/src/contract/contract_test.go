package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"

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
	os.Remove(database)
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
	//defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(&testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err2 := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
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

	var height uint32
	var contract_hash string
	var index uint8
	var holder string
	var value uint64

	if rows.Next() {
		if err3 := rows.Scan(&height, &contract_hash, &index, &holder, &value); err3 != nil {
			log.Fatal(err3)
		}
	} else {
		t.Errorf("InsertYield failed to insert row into database")
	}
}

/* Tests both serialization and deserialization */
func TestYieldSerialization(t *testing.T) {
	testPubKey := generatePubKey()
	expected := MakeYield(&testPubKey, 200000)
	serialized := expected.Serialize()
	deserialized := DeserializeYield(serialized)
	if !reflect.DeepEqual(deserialized, expected) {
		t.Errorf("Yield structs do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(reserialized, serialized) {
		t.Errorf("Byte strings do not match")
	}
}

/*
func TestMakeClaimCase1A(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(&testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	testClaim, err := MakeClaim("testDB.dat", testPubKey, 10000000)
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
	if !cmp.Equal(testClaim.PublicKey, testPubkey) {
		t.Errorf("Public Keys do not match")
	}
	// Case 3
	_, shouldNotBeNil := MakeClaim("testDB.dat", testPubKey, 1)
	if shouldNotBeNil == nil {
		t.Errorf("Made claim on empty yield pool")
	}
}

func TestMakeClaimCase1B(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 50000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	testClaim, err := MakeClaim("testDB.dat", testPubKey, 10000000)
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
	if !cmp.Equal(testClaim.PublicKey, testPubkey) {
		t.Errorf("Public Keys do not match")
	}
	// Case 3
	_, shouldNotBeNil := MakeClaim("testDB.dat", testPubKey, 1)
	if shouldNotBeNil == nil {
		t.Errorf("Made claim on empty yield pool")
	}
}

func TestMakeClaimCase2(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	testClaim, err := MakeClaim("testDB.dat", testPubKey, 50000)
	if err == nil {
		t.Errorf("Should have an error that resembles change")
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
	if !cmp.Equal(testClaim.PublicKey, testPubkey) {
		t.Errorf("Public Keys do not match")
	}
	// Case 3
	_, shouldNotBeNil := MakeClaim("testDB.dat", testPubKey, 1)
	if shouldNotBeNil == nil {
		t.Errorf("Made claim on empty yield pool")
	}
}

func TestMakeClaimPriorityCase1(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYieldFloor := MakeYield(testPubKey, 10000000)
	testYieldCeiling := MakeYield(testPubKey, 10000010)
	contractHashFloor := block.HashSHA256([]byte("This should not be claimed"))
	contractHashCeiling := block.HashSHA256([]byte("This should be claimed"))
	err := InsertYield(testYieldFloor, "testDB.dat", 35, contractHashFloor, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	err = InsertYield(testYieldFloor, "testDB.dat", 35, contractHashCeiling, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	testClaim, err := MakeClaim("testDB.dat", testPubKey, 10000008)
	if err == nil {
		t.Errorf("Should have an error that resembles change")
	}
	if !bytes.Equal(testClaim.PreviousContractHash, contractHashCeiling) {
		t.Errorf("Contract hashes do not match")
	}
	if testClaim.BlockIndex != 35 {
		t.Errorf("Block indeces do not match")
	}
	if testClaim.YieldIndex != 1 {
		t.Errorf("Yield indeces do not match")
	}
	if !cmp.Equal(testClaim.PublicKey, testPubkey) {
		t.Errorf("Public Keys do not match")
	}
}

func TestMakeClaimPriorityCase2(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYieldFloor := MakeYield(testPubKey, 10000000)
	testYieldCeiling := MakeYield(testPubKey, 10000010)
	contractHashFloor := block.HashSHA256([]byte("This should be claimed"))
	contractHashCeiling := block.HashSHA256([]byte("This should not be claimed"))
	err := InsertYield(testYieldFloor, "testDB.dat", 35, contractHashFloor, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	err = InsertYield(testYieldFloor, "testDB.dat", 35, contractHashCeiling, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
	testClaim, err := MakeClaim("testDB.dat", testPubKey, 10000002)
	if err != nil {
		t.Errorf("Failed to make claim on lesser yield")
	}
	if !bytes.Equal(testClaim.PreviousContractHash, contractHashCeiling) {
		t.Errorf("Contract hashes do not match")
	}
	if testClaim.BlockIndex != 35 {
		t.Errorf("Block indeces do not match")
	}
	if testClaim.YieldIndex != 1 {
		t.Errorf("Yield indeces do not match")
	}
	if !cmp.Equal(testClaim.PublicKey, testPubkey) {
		t.Errorf("Public Keys do not match")
	}
}

func TestClaimSerialization(t *testing.T) {
	setUpDB("testDB.db")
	defer tearDown("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 50000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 2)
	expected, err := MakeClaim(50000)
	serialized := expected.Serialize()
	deserialized := DeserializeClaim(serialized)
	if !cmp.Equal(deserialized, expected) {
		t.Errorf("Claim structs do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(reserialized, serialized) {
		t.Errorf("Byte strings do not match")
	}
}
*/
