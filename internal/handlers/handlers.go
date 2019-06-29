package handlers

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/requests"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"
)

func HandleAccountInfoRequest(dbConn *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var walletAddress = r.URL.Query().Get("w") // assume this is hex-encoded
		row, err := dbConn.Query(`SELECT * FROM account_balances WHERE public_key_hash = "` + walletAddress + `"`)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, err.Error())
			return
		}
		defer row.Close()
		if !row.Next() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		type AccountInfo struct {
			WalletAddress string
			Balance       uint64
			StateNonce    uint64
		}
		var accInfo AccountInfo
		err = row.Scan(&accInfo.WalletAddress, &accInfo.Balance, &accInfo.StateNonce)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, err.Error())
			return
		}
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
		var requestedContract = accounts.Contract{
			requestBody.Version,
			publickey.Decode(unhexedRequestPublicKey),
			requestBody.SignatureLength,
			unhexedRequestSignature,
			unhexedRequestRecipientHash,
			requestBody.Value,
			requestBody.StateNonce,
		}
		// TODO: Should use sql connection
		if err := accounts.ValidateContract(&requestedContract); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		contractChannel <- requestedContract
		return
	}
}
