package integration_testing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/requests"
)

func TestAccountInfoRequestIntegration(t *testing.T) {
	// arrange
	cli := new(http.Client)
	walletAddress := "23aafe84f813bd5093599691ea5731425effe1b8c3f7c1e3c049012558160b8c"
	var balance uint64 = 500000000000000 / 2
	var stateNonce uint64 = 0
	req, err := requests.NewAccountInfoRequest("localhost:35000", walletAddress)
	if err != nil {
		t.Errorf("Failed to create request: " + err.Error())
	}
	accountInfo := struct {
		WalletAddress string
		Balance       uint64
		StateNonce    uint64
	}{
		"", 0, 0,
	}

	// act
	resp, err := cli.Do(req)

	// assert
	if err != nil {
		t.Errorf(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wrong status code. Got %v wanted %v", resp.StatusCode, http.StatusOK)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read body: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &accountInfo); err != nil {
		t.Errorf("Failed to unmarshall body: %v", err)
	}
	if accountInfo.WalletAddress != walletAddress {
		t.Errorf("Wrong wallet address returned. Got %v wanted %v", accountInfo.WalletAddress, walletAddress)
	}
	if accountInfo.Balance != balance {
		t.Errorf("Wrong balance returned. Got %v wanted %v", accountInfo.Balance, balance)
	}
	if accountInfo.StateNonce != stateNonce {
		t.Errorf("Wrong state nonce returned. Got %v wanted %v", accountInfo.StateNonce, stateNonce)
	}
}

func TestContractRequestIntegration(t *testing.T) {

}
