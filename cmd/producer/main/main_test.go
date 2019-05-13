package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestProducerStartup(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go")
	err := cmd.Start()
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

	err = cmd.Wait()
	if err == nil {
		t.Errorf("wanted panic, did not get one")
	}
	timer.Stop()

	if _, err := os.Stat("aurum_wallet.json"); os.IsNotExist(err) {
		t.Errorf("\"aurum_wallet.json\" does not exist. Error: %s", err)
	}
}
