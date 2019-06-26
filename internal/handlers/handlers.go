package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
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
