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
	"github.com/google/go-cmp/cmp"
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

func TestMakeClaimSingle(t *testing.T) {
	// Create table
	setUpDB("testDB")
	// Single element in table
	contractHash := block.HashSHA256([]byte("previous contract"))
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testPubKey := testPrivKey.PublicKey
	pubKeyBytes := testPubKey.X.Bytes()
	pubKeyBytes = append(pubKeyBytes, testPubKey.Y.Bytes()...)
	pubKeyStr := string(pubKeyBytes)
	expected := Claim{
		PreviousContractHash: contractHash,
		YieldIndex:           0,
		Holder:               testPubKey,
	}
	statement, _ = conn.Prepare(
		`INSERT INTO unclaimed_yields(
			block_height,
			contract_hash, 
			index, 
			holder, 
			value) 
			values(?,?,?)`)
	statement.Exec(35, contractHash, 0, pubKeyStr, 250)
	conn.Close()
	actual, err := MakeClaim(testDB, 250)
	if err != nil {
		t.Errorf("Failed to make claim on valid available yield")
	}
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
	conn.Close()
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

func TestMakeContract(t *testing.T) {
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
	// Insert Yields
	// Exact amount
	// Broke
	// Change back
}

func TestSerializeContract(t *testing.T) {
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testPubKey := testPrivKey.PublicKey
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
	conn.Close()
	c, err := MakeContract(1, testDB, testPrivKey, testPubKey, 50000)
	serialized := c.Serialize()
	deserialized := DeserializeContract(serialized)
	if !cmp.Equal(c, deserialized) {
		t.Errorf("Contracts don't match")
	}
}
