package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

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
		if err := os.Remove(wallet); err != nil {
			t.Errorf("failed to remove wallet: %s", err.Error())
		}
	}()

	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPKH := block.HashSHA256(keys.EncodePublicKey(&recipient.PublicKey))
	testValue := 9000

	type testArg struct {
		name      string
		value     string
		recipient string
		wantErr   bool
	}
	testArgs := []testArg{
		{
			name:      "case where aurum_wallet.json does not exist",
			value:     string(testValue),
			recipient: string(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "case where value is negative",
			value:     string(testValue * -1),
			recipient: string(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "case where value is zero",
			value:     string(testValue - testValue),
			recipient: string(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "case where value is greater than wallet balance",
			value:     string(testValue),
			recipient: string(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "case where recipient cannot be converted to size 32 byte",
			value:     string(testValue),
			recipient: string(recipientPKH),
			wantErr:   false,
		},
		{
			name:      "check to make sure there's secret bytes, uint8(1), serialized signed contract",
			value:     string(testValue),
			recipient: string(recipientPKH),
			wantErr:   true,
		},
	}

	for _, tt := range testArgs {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ContractMessageFromInput(tt.recipient, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContractMessageFromInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
