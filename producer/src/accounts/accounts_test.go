package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"

	"testing"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"

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
		public_key_hash TEXT,
		balance INTEGER,
		nonce INTEGER);`)

	if err != nil {
		t.Error(err)
	}
	statement.Exec()
	//conn.Close()

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPublicKey := sender.PublicKey
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipient := recipientPrivateKey.PublicKey
	c := MakeContract(1, *sender, senderPublicKey, 1000, 0)
	// INSERT ABOVE VALUES INTO TABLE
	statement, err = conn.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		// return errors.New("Failed to prepare a statement for further queries")
	}
	_, err = statement.Exec(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey)), 1000, 0)
	if err != nil {
		fmt.Println(err)
		// return errors.New("Failed to execute query")
	}

	validContract := MakeContract(1, *sender, recipient, 1000, 1)
	falseContractInsufficientFunds := MakeContract(1, *recipientPrivateKey, senderPublicKey, 2000, 0)
	if !ValidateContract(validContract, table) {
		t.Errorf("Valid contract regarded as invalid")
	}
	if ValidateContract(falseContractInsufficientFunds, table) {
		t.Errorf("Invalid contract regarded as valid")
	}
}

func TestContractSerialization(t *testing.T) {
	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPublicKey := sender.PublicKey
	contract := MakeContract(1, *sender, senderPublicKey, 1000, 20)
	serialized := contract.Serialize()
	deserialized := Contract{}.Deserialize(serialized)
	if !cmp.Equal(contract, deserialized, cmp.AllowUnexported(Contract{})) {
		t.Errorf("Contracts (struct) do not match")
	}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(serialized, reserialized) {
		t.Errorf("Contracts (byte slice) do not match")
	}
}
