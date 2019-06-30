package handlers

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/validation"
	"github.com/SIGBlockchain/project_aurum/internal/requests"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"
)

// Handler for incoming account info queries
func HandleAccountInfoRequest(dbConn *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var walletAddress = r.URL.Query().Get("w") // assume this is hex-encoded
		// Query the database
		row, err := dbConn.Query(`SELECT * FROM account_balances WHERE public_key_hash = "` + walletAddress + `"`)
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
		w.WriteHeader(http.StatusFound)
		io.WriteString(w, string(marshalledStruct))
	}
}

// Handler for incoming contract requests
func HandleContractRequest(dbConn *sql.DB, contractChannel chan accounts.Contract) func(w http.ResponseWriter, r *http.Request) {
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
		// TODO: Should use sql connection
		if err := validation.ValidateContract(&requestedContract); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		contractChannel <- requestedContract
		return
	}
}
