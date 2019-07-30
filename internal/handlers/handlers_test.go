package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/pendingpool"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"

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
	statement, _ := dbConn.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
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
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned with wrong status code: got %v want %v", status, http.StatusOK)
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

func createContractNReq(version uint16, sender *ecdsa.PrivateKey, recip []byte, bal uint64, nonce uint64) (c *contracts.Contract, r *http.Request, e error) {
	returnContract, err := contracts.New(version, sender, recip, bal, nonce)
	if err != nil {
		return nil, nil, errors.New("failed to make contract : " + err.Error())
	}
	returnContract.Sign(sender)
	req, err := requests.NewContractRequest("", *returnContract)
	if err != nil {
		return nil, nil, errors.New("failed to create new contract request: " + err.Error())
	}
	return returnContract, req, nil
}

func TestContractRequestHandler(t *testing.T) {
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
	statement, _ := dbConn.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	statement.Exec()

	pMap := pendingpool.NewPendingMap()
	contractChan := make(chan contracts.Contract)
	pLock := new(sync.Mutex)
	handler := http.HandlerFunc(HandleContractRequest(dbConn, contractChan, pMap, pLock))

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderStr := hex.EncodeToString(hashing.New(publickey.Encode(&senderPrivateKey.PublicKey)))
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var recipientWalletAddress = hashing.New(publickey.Encode(&recipientPrivateKey.PublicKey))

	sender2PrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSender2Str := hex.EncodeToString(hashing.New(publickey.Encode(&sender2PrivateKey.PublicKey)))
	var walletAddress2 = hashing.New(publickey.Encode(&sender2PrivateKey.PublicKey))
	if err := accountstable.InsertAccountIntoAccountBalanceTable(dbConn, walletAddress2, 5000); err != nil {
		t.Errorf("failed to insert sender account")
	}

	testContract, req, err := createContractNReq(1, senderPrivateKey, recipientWalletAddress, 25, 1)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	testContract2, req2, err := createContractNReq(1, senderPrivateKey, recipientWalletAddress, 59, 2)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	invalidNonceContract, invalidNonceReq, err := createContractNReq(1, senderPrivateKey, recipientWalletAddress, 10, 4)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	invalidBalanceContract, invalidBalanceReq, err := createContractNReq(1, senderPrivateKey, recipientWalletAddress, 100000, 3)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	testContract3, req3, err := createContractNReq(1, senderPrivateKey, recipientWalletAddress, 100, 3)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	diffSenderContract, diffSenderReq, err := createContractNReq(1, sender2PrivateKey, recipientWalletAddress, 10, 1)
	if err != nil {
		t.Errorf("failed to make contract : %v", err)
	}

	rr := httptest.NewRecorder()
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

	tests := []struct {
		name      string
		c         *contracts.Contract
		req       *http.Request
		wantBal   uint64
		wantNonce uint64
		key       string
		status    int
	}{
		{
			"valid contract",
			testContract,
			req,
			1312,
			1,
			encodedSenderStr,
			http.StatusOK,
		},
		{
			"valid contract2",
			testContract2,
			req2,
			1337 - 25 - 59,
			2,
			encodedSenderStr,
			http.StatusOK,
		},
		{
			"invalid nonce contract",
			invalidNonceContract,
			invalidNonceReq,
			1337 - 25 - 59,
			2,
			encodedSenderStr,
			http.StatusBadRequest,
		},
		{
			"invalid balance contract",
			invalidBalanceContract,
			invalidBalanceReq,
			1337 - 25 - 59,
			2,
			encodedSenderStr,
			http.StatusBadRequest,
		},
		{
			"valid contract3",
			testContract3,
			req3,
			1337 - 25 - 59 - 100,
			3,
			encodedSenderStr,
			http.StatusOK,
		},
		{
			"Diff sender contract",
			diffSenderContract,
			diffSenderReq,
			5000 - 10,
			1,
			encodedSender2Str,
			http.StatusOK,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr = httptest.NewRecorder()
			go handler.ServeHTTP(rr, tt.req)
			time.Sleep(time.Second)

			status := rr.Code
			if status != tt.status {
				t.Errorf("handler returned with wrong status code: got %v want %v", status, tt.status)
			}
			if status == http.StatusOK && status == tt.status {
				channelledContract := <-contractChan
				if !contracts.Equals(*tt.c, channelledContract) {
					t.Errorf("contracts do not match: got %+v want %+v", *tt.c, channelledContract)
				}
				if pMap.Sender[tt.key].PendingBal != tt.wantBal {
					t.Errorf("balance do not match")
				}
				if pMap.Sender[tt.key].PendingNonce != tt.wantNonce {
					t.Errorf("state nonce do not match")
				}
			}
			if i < 5 {
				if l := len(pMap.Sender); l != 1 {
					t.Errorf("number of key-value pairs in map does not match: got %v want %v", l, 1)
				}
			} else {
				if l := len(pMap.Sender); l != 2 {
					t.Errorf("number of key-value pairs in map does not match: got %v want %v", l, 2)
				}
			}
		})
	}
}
