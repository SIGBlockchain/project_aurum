package aurumhttp

import (
	"errors"
	"net/http"
)

func AccountInfoRequest(host string, walletAddress string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", host+"/accountinfo", nil)
	req.Header.Add("walletaddress", walletAddress)
	// resp, err := http.Get(host + "/accountinfo")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Failed to get response from producer:\n" + err.Error())
	}
	return resp, nil
}

func AccountInfoRequestHandler() {}

func ContractRequest(address string) {}

func ContractRequestHandler() {}
