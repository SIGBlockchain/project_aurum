package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/pendingpool"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
	"github.com/SIGBlockchain/project_aurum/internal/accountinfo"
)

const NOT_FOUND_ERR_MSG = "No entry found for the reqeusted wallet address. Potentially wait until next block is produced to see if address is registered"

// Handler for incoming account info queries
func HandleAccountInfoRequest(dbConn *sql.DB, pMap pendingpool.PendingMap, pendingLock *sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var reqestingWalletAddress = r.URL.Query().Get("w") // assume this is hex-encoded

		var accInfo accountinfo.AccountInfo
		var marshalledStruct []byte

		pendingLock.Lock()
		pendingData, ok := pMap.Sender[reqestingWalletAddress]
		pendingLock.Unlock()

		if ok {
			accInfo = accountinfo.AccountInfo{reqestingWalletAddress, pendingData.PendingBal, pendingData.PendingNonce}
		} else {

			// Query the database
			// TODO: will most likely need a lock on this dbConnection everywhere
			row, err := dbConn.Query(sqlstatements.GET_EVERYTHING_FROM_ACCOUNT_BALANCE_BY_WALLETADDRESS, reqestingWalletAddress)

			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				io.WriteString(w, err.Error())
				return
			}
			defer row.Close()
			// If there is no row with the corresponding wallet address, return not found
			if !row.Next() {
				w.WriteHeader(http.StatusNotFound)
				io.WriteString(w, NOT_FOUND_ERR_MSG)
				return
			}

			// Fill the Account info struct
			err = row.Scan(&accInfo.WalletAddress, &accInfo.Balance, &accInfo.StateNonce)
			if err != nil {
				w.WriteHeader(http.StatusNoContent)
				io.WriteString(w, err.Error())
				return
			}
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
		var requestBody contracts.JSONContract
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if err := json.Unmarshal(buf.Bytes(), &requestBody); err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
		}
		requestedContract, err := requestBody.Unmarshal()
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, err.Error())
			return
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
