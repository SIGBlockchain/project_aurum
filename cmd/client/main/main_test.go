package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
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
