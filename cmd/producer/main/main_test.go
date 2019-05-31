package main

import (
	"bytes"
	"database/sql"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

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
		// t.SkipNow()
		dbc, _ := sql.Open("sqlite3", "metadata.tab")
		defer func() {
			if err := dbc.Close(); err != nil {
				t.Errorf("failed to close database connection")
			}
		}()
		for i := 0; i < 2; i++ {
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
	})
	t.Run("Communication", func(t *testing.T) {
		// t.SkipNow()
		cmd := exec.Command("go", "run", "main.go", "-d", "--interval=2000ms", "--blocks=2", "--port=9001")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Start(); err != nil {
			t.Errorf("failed to run main command: %s", err.Error())
		}
		dur, _ := time.ParseDuration("1000ms")
		time.Sleep(dur)
		conn, err := net.Dial("tcp", "localhost:9001")
		if err != nil {
			t.Errorf("failed to connect to producer: %s", err.Error())
		}
		defer conn.Close()
		s := []byte("testing communication")
		if _, err := conn.Write(s); err != nil {
			t.Errorf("failed to send message: %s", err.Error())
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Errorf("failed to read from connections: %s", err.Error())
		}
		if err := cmd.Wait(); err != nil {
			t.Errorf("main call returned with: %s", err.Error())
			t.Logf("Stdout: %s", string(stdout.Bytes()))
		}
		if !bytes.Equal(buf[:n], s) {
			t.Errorf("buffer does not match: %v != %v", buf[:n], s)
		}
	})
}
