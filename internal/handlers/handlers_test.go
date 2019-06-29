package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"

	"github.com/SIGBlockchain/project_aurum/internal/requests"
	_ "github.com/mattn/go-sqlite3"
)

func TestHandleAccountInfoRequest(t *testing.T) {
	req, err := requests.NewAccountInfoRequest("", "xyz")
	if err != nil {
		t.Errorf("failed to create new account info request : %v", err)
	}

	rr := httptest.NewRecorder()
	dbConn, err := sql.Open("sqlite3", constants.AccountsTable)
	if err != nil {
		t.Errorf("failed to open database connection : %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			t.Errorf("failed to close database connection : %v", err)
		}
		if err := os.Remove(constants.AccountsTable); err != nil {
			t.Errorf("failed to remove database : %v", err)
		}
	}()
	statement, _ := dbConn.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()

	handler := http.HandlerFunc(HandleAccountInfoRequest(dbConn))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned with wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var walletAddress = hashing.New(publickey.Encode(&privateKey.PublicKey))
	err = accountstable.InsertAccountIntoAccountBalanceTable(dbConn, walletAddress, 1337)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	req, err = requests.NewAccountInfoRequest("", hex.EncodeToString(walletAddress))
	if err != nil {
		t.Errorf("failed to create new account info request : %v", err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned with wrong status code: got %v want %v", status, http.StatusFound)
	}
	type AccountInfo struct {
		WalletAddress string
		Balance       uint64
		StateNonce    uint64
	}
	var accInfo AccountInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &accInfo); err != nil {
		t.Errorf("failed to unmarshall response body: %v", err)
	}

	if accInfo.WalletAddress != hex.EncodeToString(walletAddress) {
		t.Errorf("failed to get correct wallet address: got %s want %s", accInfo.WalletAddress, walletAddress)
	}
	if accInfo.Balance != 1337 {
		t.Errorf("failed to get correct balance: got %d want %d", accInfo.Balance, 1337)
	}
	if accInfo.StateNonce != 0 {
		t.Errorf("failed to get correct state nonce: got %d want %d", accInfo.StateNonce, 0)
	}
}

func TestContractRequestHandler(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var recipientWalletAddress = hashing.New(publickey.Encode(&recipientPrivateKey.PublicKey))
	testContract, err := contracts.MakeContract(1, senderPrivateKey, recipientWalletAddress, 25, 1)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}
	testContract.Sign(senderPrivateKey)
	req, err := requests.NewContractRequest("", *testContract)
	if err != nil {
		t.Errorf("failed to create new contract request: %v", err)
	}

	rr := httptest.NewRecorder()
	dbConn, err := sql.Open("sqlite3", constants.AccountsTable)
	if err != nil {
		t.Errorf("failed to open database connection : %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			t.Errorf("failed to close database connection : %v", err)
		}
		if err := os.Remove(constants.AccountsTable); err != nil {
			t.Errorf("failed to remove database : %v", err)
		}
	}()
	statement, _ := dbConn.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()

	contractChan := make(chan contracts.Contract)
	handler := http.HandlerFunc(HandleContractRequest(dbConn, contractChan))
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned with wrong status code: got %v want %v", status, http.StatusBadRequest)
		t.Logf("%s", rr.Body.String())
	}

	var walletAddress = hashing.New(publickey.Encode(&senderPrivateKey.PublicKey))
	if err := accountstable.InsertAccountIntoAccountBalanceTable(dbConn, walletAddress, 1337); err != nil {
		t.Errorf("failed to insert sender account")
	}
	req, err = requests.NewContractRequest("", *testContract)
	if err != nil {
		t.Errorf("failed to create new contract request: %v", err)
	}
	rr = httptest.NewRecorder()

	go handler.ServeHTTP(rr, req)
	channelledContract := <-contractChan
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned with wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !contracts.Equals(*testContract, channelledContract) {
		t.Errorf("contracts do not match: got %+v want %+v", *testContract, channelledContract)
	}
}
