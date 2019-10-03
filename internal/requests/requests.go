package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
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

func NewContractRequest(host string, newContract contracts.Contract) (*http.Request, error) {
	newJSONContract, err := newContract.Marshal()
	if err != nil {
		return nil, errors.New("Failed to convert contract to JSONContract: " + err.Error())
	}
	marshalledContract, err := json.Marshal(newJSONContract)
	if err != nil {
		return nil, errors.New("Failed to marshall contract: " + err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, host+endpoints.Contract, bytes.NewBuffer(marshalledContract))
	if err != nil {
		return nil, errors.New("Failed to create request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil

}

func AddPeerToDiscoveryRequest(ip string, port string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, endpoints.AddPeer, nil)
	if err != nil {
		return nil, errors.New("Failed to make new request:\n" + err.Error())
	}
	values := req.URL.Query()
	values.Add("ip", ip)
	values.Add("port", port)
	req.URL.RawQuery = values.Encode()
	return req, nil
}

func GetBlockByHeightRequest(blockHeight string) (*http.Request, error) {
	return nil, errors.New("failed to make new request")
}