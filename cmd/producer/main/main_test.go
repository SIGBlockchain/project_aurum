package main

import (
	"bytes"
	"database/sql"
	"net"
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
}

func TestRunServer(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:13131")
	if err != nil {
		t.Errorf("failed to start listener:\n%s", err.Error())
	}
	defer ln.Close()
	go RunServer(ln, false)

	buf := make([]byte, 1024)
	type serverTest struct {
		name            string
		messageToBeSent []byte
		messageToBeRcvd []byte
	}
	testArgs := []serverTest{
		{
			name:            "Standard Message",
			messageToBeSent: []byte("test message"),
			messageToBeRcvd: []byte("test message"),
		},
		{
			name:            "Aurum Message",
			messageToBeSent: producer.SecretBytes,
			messageToBeRcvd: []byte("aurum client acknowledged"),
		},
	}

	for _, ta := range testArgs {
		t.Run(ta.name, func(t *testing.T) {
			conn, err := net.Dial("tcp", "localhost:13131")
			if err != nil {
				t.Logf("failed to connect to server:\n%s", err.Error())
				t.FailNow()
			}
			_, err = conn.Write(ta.messageToBeSent)
			if err != nil {
				t.Logf("failed to send message to server:\n%s", err.Error())
				t.FailNow()
			}
			nRcvd, err := conn.Read(buf)
			if err != nil {
				t.Logf("failed to receive bytes from connection:\n%s", err.Error())
				t.FailNow()
			}
			if !bytes.Equal(buf[:nRcvd], ta.messageToBeRcvd) {
				t.Logf("messages don't match: %s != %s", string(buf[:nRcvd]), string(ta.messageToBeRcvd))
				t.FailNow()
			}
		})

	}
}
