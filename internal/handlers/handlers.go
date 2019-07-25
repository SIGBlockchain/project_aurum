package handlers

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/SIGBlockchain/project_aurum/internal/pendingpool"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/requests"
	"github.com/SIGBlockchain/project_aurum/internal/sqlqueries"
)

// Handler for incoming account info queries
func HandleAccountInfoRequest(dbConn *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var walletAddress = r.URL.Query().Get("w") // assume this is hex-encoded
		// Query the database
		// TODO: will most likely need a lock on this dbConnection everywhere
		// row, err := dbConn.Query(`SELECT * FROM account_balances WHERE public_key_hash = "` + walletAddress + `"`)
		row, err := dbConn.Query(sqlqueries.GET_EVERYTHING_BY_WALLETADDRESS + walletAddress + `"`)

		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, err.Error())
			return
		}
		defer row.Close()
		// If there is no row with the corresponding wallet address, return not found
		if !row.Next() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// TODO: Remake this struct
		type AccountInfo struct {
			WalletAddress string
			Balance       uint64
			StateNonce    uint64
		}
		var accInfo AccountInfo
		// Fill the Account info struct
		err = row.Scan(&accInfo.WalletAddress, &accInfo.Balance, &accInfo.StateNonce)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, err.Error())
			return
		}
		// Marshall the struct into the response body
		marshalledStruct, err := json.Marshal(accInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(marshalledStruct))
	}
}

// Handler for incoming contract requests
func HandleContractRequest(dbConn *sql.DB, contractChannel chan contracts.Contract, pMap pendingpool.PendingMap, pendingLock *sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody requests.JSONContract
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if err := json.Unmarshal(buf.Bytes(), &requestBody); err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
		}
		// TODO: Need a JSON to Contract function
		unhexedRequestPublicKey, err := hex.DecodeString(requestBody.SenderPublicKey)
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
		}
		unhexedRequestSignature, err := hex.DecodeString(requestBody.Signature)
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
		}
		unhexedRequestRecipientHash, err := hex.DecodeString(requestBody.RecipientWalletAddress)
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
		}
		var requestedContract = contracts.Contract{
			requestBody.Version,
			publickey.Decode(unhexedRequestPublicKey),
			requestBody.SignatureLength,
			unhexedRequestSignature,
			unhexedRequestRecipientHash,
			requestBody.Value,
			requestBody.StateNonce,
		}
		pendingLock.Lock()
		if err := pMap.Add(&requestedContract, dbConn); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			contractChannel <- requestedContract
		}
		pendingLock.Unlock()
		return
	}
}
