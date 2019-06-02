package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
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

	type testArg struct {
		name      string
		value     string
		recipient string
		wantErr   bool
	}
	testArgs := []testArg{
		{
			// case where value is negative
		},
		{
			// case where value is zero
		},
		{
			// case where value is greater than wallet balance
		},
		{
			// case where recipient cannot be converted to size 32 byte
		},
		{
			// valid case,
			//check to make sure there's secret bytes, uint8(1), serialized signed contract
		},
	}
	t.Logf("%v", testArgs) // delete this when for loop is complete

	// for _, arg := range testArgs {}
}
