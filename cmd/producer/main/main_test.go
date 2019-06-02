package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"net"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"

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
	var byteChan chan []byte
	go RunServer(ln, byteChan, false)

	buf := make([]byte, 1024)

	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
	contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.SignContract(senderPrivateKey)
	serializedContract, err := contract.Serialize()

	var contractMessage []byte
	contractMessage = append(contractMessage, producer.SecretBytes...)
	contractMessage = append(contractMessage, 1)
	contractMessage = append(contractMessage, serializedContract...)

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
		{
			name:            "Contract Message",
			messageToBeSent: contractMessage,
			messageToBeRcvd: []byte("received contract message"),
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
			// Below code tests the channel communication
			// if ta.name == "Contract Message" {
			// 	channelledContract := <-byteChan
			// 	if !bytes.Equal(channelledContract, serializedContract) {
			// 		t.Errorf("channel contents does not match desired slice:\n%v != %v", channelledContract, serializedContract)
			// 	}
			// }
		})

	}
}

func TestRunServer2(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:13131")
	if err != nil {
		t.Errorf("failed to startup listener")
	}
	byteChan := make(chan []byte)
	go RunServer2(ln, byteChan, false)
	conn, err := net.Dial("tcp", "localhost:13131")
	if err != nil {
		t.Errorf("failed to connect to server")
	}
	_, err = conn.Write(producer.SecretBytes)
	if err != nil {
		t.Errorf("failed to send message")
	}
	res := <-byteChan
	if !bytes.Equal(res, producer.SecretBytes) {
		t.Errorf("result does not match:\n%s != %s", string(res), string(producer.SecretBytes))
	}
}
