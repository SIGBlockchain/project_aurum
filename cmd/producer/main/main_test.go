package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
)

var removeFiles = false

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
	cmd := exec.Command("go", "run", "main.go", "-d", "-g", "--supply=100", "-t")
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
	cmd := exec.Command("go", "run", "main.go", "-d", "-g", "--supply=100", "-t")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Errorf("failed to run main command because: %s", err.Error())
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf("main call returned with: %s.", err.Error())
		t.Logf("Stderr: %s", string(stderr.Bytes()))
	}
	// cmd = exec.Command("go", "run", "main.go", "-d", "--interval=5000ms", ">", "test.txt", "-t")
	// cmd.Stderr = &stderr
	// if err := cmd.Start(); err != nil {
	// 	t.Errorf("failed to run main command because: %s", err.Error())
	// }
	// if err := cmd.Wait(); err != nil {
	// 	t.Errorf("main call returned with: %s.", err.Error())
	// 	t.Logf("Stderr: %s", string(stderr.Bytes()))
	// }
}

// kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
// 	if err := cmd.Start(); err != nil {
// 		t.Errorf("failed to run main command because: %s", err.Error())
// 	}
// 	time.AfterFunc(10*time.Second, func() {
// 		if err := kill.Run(); err != nil {
// 			t.Errorf("failed to kill process: %s", err.Error())
// 		}
// 	}).Stop()

// timer := time.AfterFunc(3*time.Second, func() {
// 	fmt.Println("Timer off")
// 	err := cmd.Process.Kill()
// 	if err != nil {
// 		t.Errorf("Failed to kill process: %s", err)
// 	}
// })
// err = cmd.Wait()
// timer.Stop()
