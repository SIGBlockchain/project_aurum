package requests

import (
	"errors"
	"net/http"

	"github.com/SIGBlockchain/project_aurum/internal/endpoints"
)

func NewAccountInfoRequest(host string, walletAddress string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, host+endpoints.AccountInfo, nil)
	if err != nil {
		return nil, errors.New("Failed to make new request:\n" + err.Error())
	}
	values := req.URL.Query()
	values.Add("w", walletAddress)
	req.URL.RawQuery = values.Encode()
	return req, nil
}
