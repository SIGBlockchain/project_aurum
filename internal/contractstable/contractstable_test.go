package contractstable

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

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
	contract, _ := contracts.New(1, senderPrivateKey, encodedRecipientKey, 1, 1)

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
	var testContracts [](*contracts.Contract)

	contract1, _ := contracts.New(1, senderPrivateKey, encodedRecipientKey, 1, 1)
	testContracts = append(testContracts, contract1)

	contract2, _ := contracts.New(1, senderPrivateKey, encodedRecipientKey, 10, 2)
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

	var testContracts [](*contracts.Contract)
	contract1, _ := contracts.New(1, senderPrivateKey, encodedRecipientKey, 1, 1)
	contract2, _ := contracts.New(1, senderPrivateKey, encodedRecipientKey, 10, 2)
	testContracts = append(testContracts, contract1, contract2)

	err := InsertContractsIntoContractsTable(testContracts)

	if err != nil {
		t.Errorf("Failed to insert contracts into contracts table: %s", err)
	}

	// Act
	contracts, err := GetContractsFromContractsTable(dbConn, &senderPrivateKey.PublicKey, 1)
	if err != nil {
		t.Errorf("Failed to get contracts from contracts table")
	}

	// Assert
	if len(contracts) != len(testContracts) {
		t.Errorf("Failed to get the correct number of contracts from contracts table")
	}

	for n := 0; n < len(contracts); n++ {
		result := contracts[n].Equals(*testContracts[n])
		if !result {
			t.Errorf("Failed to get the expected contracts from contracts table")
		}
	}
}
