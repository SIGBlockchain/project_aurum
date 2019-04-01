package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"
)

func TestValidateContract(t *testing.T) {
	table := "table.dat"
	conn, err := sql.Open("sqlite3", table)
	defer func() {
		os.Remove(table)
	}()
	statement, err := conn.Prepare(
		`CREATE TABLE IF NOT EXISTS account_balances ( 
		public_key TEXT, 
		balance INTEGER, 
		nonce INTEGER);`)

	if err != nil {
		t.Error(err)
	}
	statement.Exec()
	conn.Close()

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPublicKey := sender.PublicKey
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipient := recipientPrivateKey.PublicKey
	MakeContract(1, *sender, senderPublicKey, 1000, 0)
	validContract := MakeContract(1, *sender, recipient, 1000, 1)
	falseContractInsufficientFunds := MakeContract(1, *recipientPrivateKey, senderPublicKey, 2000, 0)
	if !ValidateContract(validContract, table) {
		t.Errorf("Valid contract regarded as invalid")
	}
	if ValidateContract(falseContractInsufficientFunds, table) {
		t.Errorf("Invalid contract regarded as invalid")
	}
}

func TestContractSerialization(t *testing.T) {
	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPublicKey := sender.PublicKey
	contract := MakeContract(1, *sender, senderPublicKey, 1000, 20)
	serialized := contract.Serialize()
	deserialized := Contract{}.Deserialize(serialized)
	if !cmp.Equal(contract, deserialized) {
		t.Errorf("Contracts (struct) do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(serialized, reserialized) {
		t.Errorf("Contracts (byte slice) do not match")
	}
}
