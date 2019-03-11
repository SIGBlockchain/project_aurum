package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"

	block "../block"
	_ "github.com/mattn/go-sqlite3"
)

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

func generatePubKey() ecdsa.PublicKey {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return privateKey.PublicKey
}

func TestMakeYield(t *testing.T) {
	testPubKey := generatePubKey()
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

func TestInsertYield(t *testing.T) {
	setUpDB("testDB.dat")
	testPubKey := generatePubKey()
	testYield := MakeYield(testPubKey, 10000000)
	contractHash := block.HashSHA256([]byte{'b', 'l', 'k', 'c', 'h', 'a', 'i', 'n'})
	err := InsertYield(testYield, "testDB.dat", 35, contractHash, 1)
	if err != nil {
		t.Errorf("Failed to insert yield")
	}
}

func TestSerialization(t *testing.T) {
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
