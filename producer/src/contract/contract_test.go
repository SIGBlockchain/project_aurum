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

func TestMakeClaim(t *testing.T) {
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	statement, _ := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS unclaimed_yields( 
		contract_hash TEXT, 
		index INTEGER, 
		holder TEXT, 
		value INTEGER)`)
	statement.Exec()
	// First case is a with an equal sized claim
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
			contract_hash, 
			index, 
			holder, 
			value) 
			values(?,?,?)`)
	statement.Exec(contractHash, 0, pubKeyStr, 250)
	actual, err := MakeClaim(testDB, 250)
	if err != nil {
		t.Errorf("Failed to make claim on valid available yield")
	}
	// If cmp.Equal is still referred to as undefined,
	// inform test-maker of possible need for change
	// only after you have fixed above errors
	if cmp.Equal(expected, actual) {
		t.Errorf("Claims do not match")
		conn.Close()
		os.Remove(testDB)
	}
	// At this point, that particular yield should be removed
	actual, err = MakeClaim(testDB, 200)
	if err == nil {
		t.Errorf("Made claim to yield that shouldn't exist")
		conn.Close()
		os.Remove(testDB)
	}
	// Now to test a large Claim with smaller unclaimed-yield
	contractHash = block.HashSHA256([]byte("this is a new yield"))
	expected.PreviousContractHash = contractHash
	expected.YieldIndex = 3
	statement, _ = conn.Prepare(
		`INSERT INTO unclaimed_yields(
			contract_hash, 
			index, 
			holder, 
			value) 
			values(?,?,?)`)
	statement.Exec(contractHash, 3, pubKeyStr, 250)
	actual, err = MakeClaim(testDB, 500)
	if err != nil {
		t.Errorf("Failed to make claim on valid available yield")
	}
	if cmp.Equal(expected, actual) {
		t.Errorf("Claims do not match")
		conn.Close()
		os.Remove(testDB)
	}
	conn.Close()
	os.Remove(testDB)
}
