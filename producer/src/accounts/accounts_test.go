package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"

	// "database/sql"
	// "os"
	"testing"
	// _ "github.com/mattn/go-sqlite3"
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
	c := MakeContract(1, *sender, senderPublicKey, 1000, 0)
	// insert contract information into table*****************************
	conn, err = sql.Open("sqlite3", table)
	if err != nil {
		t.Errorf("Unable to open sqlite3 database")
	}
	defer conn.Close()

	statement, err = conn.Prepare("INSERT INTO account_balances (public_key, balance, nonce) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		t.Errorf("Failed to prepare a statement for further queries")
	}
	_, err = statement.Exec(c.SenderPubKey, c.Value, c.Nonce)
	if err != nil {
		t.Errorf("Failed to execute query")
	}
	fmt.Println("DONE ADDING TO TABLE")
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
