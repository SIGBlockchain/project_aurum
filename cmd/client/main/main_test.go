package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

var removeWallet = true

func TestSetupFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "-s")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		t.Errorf("failed to run main command: %s", err.Error())
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf("main call returned with: %s", err.Error())
		t.Logf("Stdout: %s", string(stdout.Bytes()))
	}
	if _, err := os.Stat(wallet); os.IsNotExist(err) {
		t.Errorf("failed to generate aurum_wallet: %s", err.Error())
	}
	defer func() {
		if err := os.Remove(wallet); err != nil {
			t.Errorf("failed to remove wallet: %s", err.Error())
		}
	}()
}

func TestContractMessageFromInput(t *testing.T) {
	if err := client.SetupWallet(); err != nil {
		t.Errorf("failed to setup wallet: %s", err.Error())
	}
	defer func() {
		if removeWallet {
			os.Remove(wallet)
		}
	}()

	file, err := os.Open(wallet)
	if err != nil {
		t.Errorf("failed to open wallet: %s", err.Error())
	}
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	var wd walletData
	jBytes, _ := ioutil.ReadAll(file)
	json.Unmarshal(jBytes, &wd)
	wd.Balance = 50 // change the balance for the test
	jsonEncoded, _ := json.Marshal(wd)
	err = ioutil.WriteFile(wallet, jsonEncoded, 0644)
	if err != nil {
		t.Errorf("Failed to write into wallet: %s", err.Error())
	}
	file.Close()

	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPKH := hashing.New(publickey.Encode(&recipient.PublicKey))
	testValue := 9000

	type testArg struct {
		name      string
		value     string
		recipient string
		wantErr   bool
	}
	testArgs := []testArg{
		{
			name:      "case where value is negative",
			value:     strconv.Itoa(testValue * -1),
			recipient: hex.EncodeToString(recipientPKH),
			wantErr:   true,
		},
		{
			name:      "case where value is zero",
			value:     strconv.Itoa(testValue - testValue),
			recipient: hex.EncodeToString(recipientPKH),
			wantErr:   true,
		},
		{
			name:      "case where value is greater than wallet balance",
			value:     strconv.Itoa(testValue),
			recipient: hex.EncodeToString(recipientPKH),
			wantErr:   true,
		},
		{
			name:      "case where recipient cannot be converted to size 32 byte",
			value:     strconv.Itoa(50),
			recipient: hex.EncodeToString([]byte("hello")),
			wantErr:   true,
		},
		{
			name:      "check to make sure there's secret bytes, uint8(1), serialized signed contract",
			value:     strconv.Itoa(50),
			recipient: hex.EncodeToString(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "case where aurum_wallet.json does not exist",
			value:     strconv.Itoa(testValue),
			recipient: hex.EncodeToString(recipientPKH),
			wantErr:   true,
		},
	}

	for _, tt := range testArgs {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "case where aurum_wallet.json does not exist" {
				err := os.Remove(wallet)
				if err != nil {
					t.Errorf("failed to remove wallet: %s", err.Error())
				}
			}
			_, err := ContractMessageFromInput(tt.value, tt.recipient)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContractMessageFromInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
