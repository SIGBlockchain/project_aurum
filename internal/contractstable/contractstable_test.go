package contractstable

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

func setUp(dbName string) *sql.DB {
	conn, err := sql.Open("sqlite3", dbName)
	if err != nil {
		panic("Failed to open database")
	}

	statement, _ := conn.Prepare(sqlstatements.CREATE_CONTRACT_TABLE)
	statement.Exec()

	return conn
}

func tearDown(dbConn *sql.DB, dbName string) {
	dbConn.Close()
	os.Remove(dbName)
}

func TestInsertContractIntoContractsTable(t *testing.T) {
	// Arrange
	dbConn := setUp("testContracts.db")
	defer tearDown(dbConn, "testContracts.db")

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPubKeyHash := hashing.New(encodedRecipientKey)
	contract, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 1, 1)

	// Act
	err := InsertContractIntoContractsTable(contract)

	// Assert
	if err != nil {
		t.Errorf("Failed to insert contract into contracts table: %s", err)
	}

}

func TestInsertContractsIntoContractsTable(t *testing.T) {
	// Arrange
	dbConn := setUp("testContracts.db")
	defer tearDown(dbConn, "testContracts.db")

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPubKeyHash := hashing.New(encodedRecipientKey)
	var testContracts [](*contracts.Contract)

	contract1, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 1, 1)
	testContracts = append(testContracts, contract1)

	contract2, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 10, 2)
	testContracts = append(testContracts, contract2)

	// Act
	err := InsertContractsIntoContractsTable(testContracts)

	// Assert
	if err != nil {
		t.Errorf("Failed to insert contracts into contracts table: %s", err)
	}

}

func TestGetContractsFromContractsTable(t *testing.T) {
	// Arrange
	dbConn := setUp("testContracts.db")
	defer tearDown(dbConn, "testContracts.db")

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPubKeyHash := hashing.New(encodedRecipientKey)
	var testContracts [](*contracts.Contract)

	contract1, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 1, 1)
	err := InsertContractIntoContractsTable(contract1)
	testContracts = append(testContracts, contract1)

	contract2, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 10, 2)
	err2 := InsertContractIntoContractsTable(contract2)
	testContracts = append(testContracts, contract2)

	if err != nil || err2 != nil {
		t.Errorf("Failed to insert contract into tableContracts table: %s", err)
	}

	// Act
	tableContracts, err := GetContractsFromContractsTable(dbConn, &senderPrivateKey.PublicKey, 1)
	if err != nil {
		t.Errorf("Failed to get tableContracts from tableContracts table")
	}

	// Assert
	if len(tableContracts) != len(testContracts) {
		t.Errorf("Failed to get the correct number of tableContracts from tableContracts table")
	}

	for n := 0; n < len(tableContracts); n++ {
		result := (*tableContracts[n]).Equals(*testContracts[n])
		if !result {
			t.Errorf("Failed to get the expected tableContracts from tableContracts table")
		}
	}
}
