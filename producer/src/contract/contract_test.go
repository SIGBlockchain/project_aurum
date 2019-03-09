package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

	block "../block"
	_ "github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"
)

func TestMakeYield(t *testing.T) {
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testPubKey := testPrivKey.PublicKey
	pubKeyBytes := testPubKey.X.Bytes()
	pubKeyBytes = append(pubKeyBytes, testPubKey.Y.Bytes()...)
	hashedKey := block.HashSHA256(pubKeyBytes)
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

// Make Claim Cases:
/*
Makes claim appropriately with one unclaimed-yield
Does not make claim when there are no yields
A priority test - not yet available
Priority test would either be block height,
absolute value difference between desired value and available yield values
smallest values first
largest values first
*/

func TestMakeClaimSingle(t *testing.T) {
	// Create table
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	statement, _ := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS unclaimed_yields( 
		block_height INTEGER,
		contract_hash TEXT, 
		index INTEGER, 
		holder TEXT, 
		value INTEGER)`)
	statement.Exec()
	// Single element in table
	contractHash := block.HashSHA256([]byte("previous contract"))
	expected := Claim{}
	expected.PreviousContractHash = contractHash
	expected.YieldIndex = 0
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubKeyBytes := testPubKey.X.Bytes()
	pubKeyBytes = append(pubKeyBytes, testPubKey.Y.Bytes()...)
	pubKeyStr := string(pubKeyBytes)
	expected.Holder = testPrivKey.PublicKey
	statement, _ = conn.Prepare(
		`INSERT INTO unclaimed_yields(
			block_height,
			contract_hash, 
			index, 
			holder, 
			value) 
			values(?,?,?)`)
	statement.Exec(35, contractHash, 0, pubKeyStr, 250)
	actual, err := MakeClaim(testDB, 250)
	if err != nil {
		t.Errorf("Failed to make claim on valid available yield")
	}
	// If cmp.Equal is still referred to as undefined,
	// inform test-maker of possible need for change
	// only after you have fixed above errors
	if !(cmp.Equal(expected, actual)) {
		t.Errorf("Claims do not match")
		conn.Close()
		os.Remove(testDB)
	}
	conn.Close()
	os.Remove(testDB)
}

func TestMakeClaimEmpty(t *testing.T) {
	// Create table
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	statement, _ := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS unclaimed_yields( 
		block_height INTEGER,
		contract_hash TEXT, 
		index INTEGER, 
		holder TEXT, 
		value INTEGER)`)
	statement.Exec()
	// No elements in table
	// Claim should be empty
	expected := Claim{}
	actual, err := MakeClaim(testDB, 250)
	if err == nil {
		t.Errorf("Made claim on unavailable unclaimed-yield")
	}
	if !(cmp.Equal(expected, actual)) {
		t.Errorf("Claims do not match")
		conn.Close()
		os.Remove(testDB)
	}
	conn.Close()
	os.Remove(testDB)
}
