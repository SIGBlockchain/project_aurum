package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

func TestProducerStartup(t *testing.T) {
	var pkhashes [][]byte
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := producer.BringOnTheGenesis(pkhashes, 1000)
	err := producer.Airdrop(ledger, metadata, genny)
	cmd := exec.Command("go", "run", "main.go")
	err = cmd.Start()
	if err != nil {
		t.Errorf("Failed to run main command: %s", err)
	}

	timer := time.AfterFunc(5*time.Second, func() {
		fmt.Println("Timer off")
		err := cmd.Process.Kill()
		if err != nil {
			t.Errorf("Failed to kill process: %s", err)
		}
	})

	cmd.Wait()
	timer.Stop()
}
