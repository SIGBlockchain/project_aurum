package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
)

var removeFiles = true

func TestProducerGenesisFlag(t *testing.T) {
	producer.GenerateGenesisHashFile(25)
	defer func() { // Function is dangerous, consider only running with flag
		if removeFiles {
			if err := os.Remove("genesis_hashes.txt"); err != nil {
				t.Errorf("failed to remove genesis hashes")
			}
			if err := os.Remove("blockchain.dat"); err != nil {
				t.Errorf("failed to remove blockchain.dat")
			}
			if err := os.Remove("metadata.tab"); err != nil {
				t.Errorf("failed to remove metadatata.tab")
			}
		}
	}()
	cmd := exec.Command("go", "run", "main.go", "-d", "-g", "--supply=100")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		t.Errorf("failed to run main command because: %s", err.Error())
	}
	err = cmd.Wait()
	if err != nil {
		t.Errorf("main call returned with: %s.", err.Error())
		t.Logf("Stderr: %s", string(stderr.Bytes()))
	}
}

func TestLoop(t *testing.T) {
	producer.GenerateGenesisHashFile(25)
	cmd := exec.Command("go", "run", "-d", "-g", "--supply=100")
	err := cmd.Start()
	if err != nil {
		t.Errorf("failed to run main command: %s", err.Error())
	}
}
