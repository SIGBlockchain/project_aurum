package main

import (
	"bytes"
	"database/sql"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
)

var removeFiles = true

func TestSuite(t *testing.T) {
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
	t.Run("Genesis", func(t *testing.T) {
		cmd := exec.Command("go", "run", "main.go", "-d", "-g", "--supply=100", "-t")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		err := cmd.Start()
		if err != nil {
			t.Errorf("failed to run main command because: %s", err.Error())
		}
		err = cmd.Wait()
		if err != nil {
			t.Errorf("main call returned with: %s.", err.Error())
			t.Logf("Stdout: %s", string(stdout.Bytes()))
		}
	})
	t.Run("Loop", func(t *testing.T) {
		dbc, _ := sql.Open("sqlite3", "metadata.tab")
		defer func() {
			if err := dbc.Close(); err != nil {
				t.Errorf("failed to close database connection")
			}
		}()
		// var ms runtime.MemStats
		for i := 0; i < 10; i++ {
			var count int
			if err := dbc.QueryRow("SELECT COUNT(*) FROM METADATA").Scan(&count); err != nil {
				t.Errorf("failed to query rows")
			}
			if count != i+1 {
				t.Errorf("invalid number of blocks: %d != %d", i+1, count)
			}
			cmd := exec.Command("go", "run", "main.go", "-d", "--interval=1000ms", "-t")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Start(); err != nil {
				t.Errorf("failed to run main command because: %s", err.Error())
			}
			if err := cmd.Wait(); err != nil {
				t.Errorf("main call returned with: %s.", err.Error())
				t.Logf("Stdout: %s", string(stdout.Bytes()))
			}
		}
		// runtime.ReadMemStats(&ms)
		// t.Logf("Bytes of allocated heap objects: %d", ms.Alloc)
		// t.Logf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
		// t.Logf("Count of heap objects allocated: %d", ms.Mallocs)
		// t.Logf("Count of heap objects freed: %d", ms.Frees)
	})
}

func TestCommunication(t *testing.T) {

}
