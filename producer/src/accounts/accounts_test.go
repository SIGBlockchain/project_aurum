package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"testing"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"

	_ "github.com/mattn/go-sqlite3"
)

func TestValidateContract(t *testing.T) {
	table := "table.db"
	conn, err := sql.Open("sqlite3", table)
	defer conn.Close()

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

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPublicKey := sender.PublicKey
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipient := recipientPrivateKey.PublicKey
	c := MakeContract(1, *sender, senderPublicKey, 1000, 0)
	// INSERT ABOVE VALUES INTO TABLE
	sqlQuery := fmt.Sprintf("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (\"%s\", %d, %d)", hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey))), 1000, 0)
	log.Printf("Executing query %s", sqlQuery)
	_, err = conn.Exec(sqlQuery)
	if err != nil {
		t.Errorf("Unable to insert values into table: %v", err)
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
	//if !cmp.Equal(contract, deserialized, cmp.AllowUnexported(Contract{})) {			// could not figure this out
	//	t.Errorf("Contracts (struct) do not match")										// reflect.DeepEquals...?
	//}
	reserialized := deserialized.Serialize()
	if !bytes.Equal(serialized, reserialized) {
		t.Errorf("Contracts (byte slice) do not match")
	}
}
