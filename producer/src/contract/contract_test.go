package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	block "../block"
	keys "../keys"
	_ "github.com/mattn/go-sqlite3"
)

/* <Testing adjunct> function for setting up database */
func setUpDB(database string) {
	conn, _ := sql.Open("sqlite3", database)
	statement, _ := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS unclaimed_yields( 
		block_height INTEGER,
		contract_hash TEXT, 
		index INTEGER, 
		holder TEXT, 
		value INTEGER)`)
	statement.Exec()
	conn.Close()
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
	encodedPubKey := keys.EncodePublicKey(testPubKey)
	hashedKey := block.HashSHA256(encodedPubKey)
	var expectedRecipient []byte = hashedKey[:]
	expectedValue := 1000000
	actual := MakeYield(testPubKey, 1000000)
	if bytes.Equal(expectedRecipient, actual.Recipient) {
		t.Errorf("Recipients do not match")
	}
	if expectedValue != actual.Value {
		t.Errorf("Values do not match")
	}
}

/* Simple test to make sure insertion is working */
func TestInsertYield(t *testing.T) {
	defer tearDown("testDB.db")
	setUpDB("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
}

/* Tests both serialization and deserialization */
func TestYieldSerialization(t *testing.T) {
	testPubKey := generatePubKey()
	expected := MakeYield(testPubKey, 200000)
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

func TestMakeClaim(t *testing.T) {
	defer tearDown("testDB.db")
	setUpDB("testDB.db")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 10000000)
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
	_, shouldReturnError := MakeClaim("testDB.dat", testPubKey, 1)
	if shouldReturnError == nil {
		t.Errorf("Made claim on empty yield pool")
	}
}
