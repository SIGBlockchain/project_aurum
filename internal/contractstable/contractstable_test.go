package contractstable

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"strconv"
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
	encodedSenderPubkey, _ := publickey.Encode(&senderPrivateKey.PublicKey)

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

	dbConn.Prepare(sqlstatements.GET_CONTRACT_BY_SENDER_PUBLIC_KEY_AND_NONCE)
	row, err := dbConn.Query(hex.EncodeToString(hashing.New(encodedSenderPubkey)), strconv.FormatUint(contract.StateNonce, 10))
	if err != nil {
		t.Errorf("Failed to acquire rows from contracts table")
	}
	row.Next()
	var result []byte
	err = row.Scan(&result)
	if err != nil {
		t.Errorf("Failed to scan rows: %s", err)
	}
	var cResult *contracts.Contract
	err = cResult.Deserialize(result)
	if err != nil {
		t.Errorf("Failed to deserialize contract")
	}
	if !cResult.Equals(*contract) {
		t.Error("Result contract does not equal to contract.\nWant: " + cResult.ToString() + "\nGot: " + contract.ToString())
	}
}

func TestInsertContractsIntoContractsTable(t *testing.T) {
	// Arrange
	dbConn := setUp("testContracts.db")
	defer tearDown(dbConn, "testContracts.db")

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPubkey, _ := publickey.Encode(&senderPrivateKey.PublicKey)

	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPubKeyHash := hashing.New(encodedRecipientKey)

	var testContracts [](*contracts.Contract)
	contract1, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 1, 1)
	contract2, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 10, 2)
	testContracts = append(testContracts, contract1, contract2)

	// Act
	err := InsertContractsIntoContractsTable(testContracts)

	// Assert
	if err != nil {
		t.Errorf("Failed to insert multiple contracts into contracts table: %s", err)
	}

	for _, contract := range testContracts {
		dbConn.Prepare(sqlstatements.GET_CONTRACT_BY_SENDER_PUBLIC_KEY_AND_NONCE)
		row, err := dbConn.Query(hex.EncodeToString(hashing.New(encodedSenderPubkey)), strconv.FormatUint(contract.StateNonce, 10))
		if err != nil {
			t.Errorf("Failed to acquire rows from contracts table")
		}
		row.Next()
		var result []byte
		err = row.Scan(&result)
		if err != nil {
			t.Errorf("Failed to scan rows: %s", err)
		}
		var cResult *contracts.Contract
		err = cResult.Deserialize(result)
		if err != nil {
			t.Errorf("Failed to deserialize contract")
		}
		if !cResult.Equals(*contract) {
			t.Error("Result contract does not equal to contract.\nWant: " + cResult.ToString() + "\nGot: " + contract.ToString())
		}
	}
}

func TestGetContractsFromContractsTable(t *testing.T) {
	// Arrange
	dbConn := setUp("testContracts.db")
	defer tearDown(dbConn, "testContracts.db")

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	senderPubKeyHash := hashing.New(encodedSenderKey)

	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPubKeyHash := hashing.New(encodedRecipientKey)

	var testContracts [](*contracts.Contract)
	contract1, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 1, 1)
	contract2, _ := contracts.New(1, senderPrivateKey, recipientPubKeyHash, 10, 2)
	contract3, _ := contracts.New(1, recipientPrivateKey, senderPubKeyHash, 100, 1)
	testContracts = append(testContracts, contract1, contract2, contract3)

	err := InsertContractsIntoContractsTable(testContracts)
	if err != nil {
		t.Errorf("Failed to insert multiple contracts into tableContracts table: %s", err)
	}

	// Act
	tableContracts, err := GetContractsFromContractsTable(dbConn, &senderPrivateKey.PublicKey, 1)
	if err != nil {
		t.Errorf("Failed to get contracts from ContractsTable")
	}

	// Assert
	if len(tableContracts) != len(testContracts)-1 {
		t.Errorf("Failed to get the correct number of contracts from ContractsTable")
	}

	for n := 0; n < len(tableContracts)-1; n++ {
		if !((*tableContracts[n]).Equals(*testContracts[n])) {
			t.Errorf("Failed to get the expected contracts from ContractsTable")
		}
	}
}
