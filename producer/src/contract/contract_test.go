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

func TestGetUnclaimedYield(t *testing.T) {
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	statement, _ := conn.Prepare(
		`CREATE TABLE IF NOT EXiSTS unclaimed_yields( 
		contract_hash TEXT, 
		index INTEGER, 
		holder TEXT, 
		value INTEGER)`)
	statement.Exec()
	contractHash := block.HashSHA256([]byte("previous contract"))
	statement, _ = conn.Prepare(
		`INSERT INTO unclaimed_yields(
			contract_hash, 
			index, 
			holder, 
			value) 
			values(?,?,?)`)
	statement.Exec(contractHash, 0, 250)
	expected := Claim{}
	expected.PreviousContractHash = contractHash
	expected.YieldIndex = 0
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	expected.Holder = testPrivKey.PublicKey
	actual, err := MakeClaim(testDB, 250)
	if err != nil {
		t.Errorf("Failed to make claim on valid available yield")
	}

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
	conn.Close()
	os.Remove(testDB)

}
