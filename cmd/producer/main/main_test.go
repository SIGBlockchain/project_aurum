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

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

var removeFiles = true

func TestSuite(t *testing.T) {
	producer.GenerateGenesisHashFile(25)
	defer func() { // Function is dangerous, consider only running with flag
		if removeFiles {
			if err := os.Remove("genesis_hashes.txt"); err != nil {
				t.Errorf("failed to remove genesis hashes:\n%s", err.Error())
			}
			if err := os.Remove("blockchain.dat"); err != nil {
				t.Errorf("failed to remove blockchain.dat:\n%s", err.Error())
			}
			if err := os.Remove("metadata.tab"); err != nil {
				t.Errorf("failed to remove metadatata.tab:\n%s", err.Error())
			}
		}
	}()
	type testArg struct {
		name       string
		runCommand []string
	}
	testArgs := []testArg{
		{
			name:       "Genesis",
			runCommand: []string{"go", "run", "main.go", "-d", "-g", "--supply=100", "-t"},
		},
		{
			name:       "Loop",
			runCommand: []string{"go", "run", "main.go", "-d", "--interval=1000ms", "--blocks=3"},
		},
	}
	var stdout bytes.Buffer
	for _, arg := range testArgs {
		t.Run(arg.name, func(t *testing.T) {
			cmd := exec.Command(arg.runCommand[0], arg.runCommand[1:]...)
			cmd.Stdout = &stdout
			if err := cmd.Start(); err != nil {
				t.Errorf("failed to run main command because: %s", err.Error())
			}
			if err := cmd.Wait(); err != nil {
				t.Errorf("main call returned with: %s.", err.Error())
				t.Logf("Stdout: %s", string(stdout.Bytes()))
			}
		})
	}
	dbc, err := sql.Open("sqlite3", "metadata.tab")
	if err != nil {
		t.Errorf("failed to open sqlite database")
	}
	defer func() {
		if err := dbc.Close(); err != nil {
			t.Errorf("failed to close database connection")
		}
	}()
	var count int
	if err := dbc.QueryRow("SELECT COUNT(*) FROM METADATA").Scan(&count); err != nil {
		t.Errorf("failed to query rows")
	}
	if count != 4 {
		t.Errorf("invalid number of blocks: %d != %d", count, 4)
	}
}

func TestRunServer(t *testing.T) {
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
	type testArg struct {
		name            string
		messageToBeSent []byte
		messageToBeRcvd []byte
	}
	testArgs := []testArg{
		{
			name:            "Aurum message",
			messageToBeSent: producer.SecretBytes,
			messageToBeRcvd: []byte("Thank you."),
		},
		{
			name:            "Contract message",
			messageToBeSent: contractMessage,
			messageToBeRcvd: []byte("Thank you."),
		},
	}
	ln, err := net.Listen("tcp", "localhost:13131")
	if err != nil {
		t.Errorf("failed to startup listener")
	}
	byteChan := make(chan []byte)
	buf := make([]byte, 1024)
	go RunServer(ln, byteChan, false)
	for _, arg := range testArgs {
		conn, err := net.Dial("tcp", "localhost:13131")
		if err != nil {
			t.Errorf("failed to connect to server")
		}
		_, err = conn.Write(arg.messageToBeSent)
		if err != nil {
			t.Errorf("failed to send message")
		}
		nRead, err := conn.Read(buf)
		if err != nil {
			t.Errorf("failed to read from connections:\n%s", err.Error())
		}
		if !bytes.Equal(buf[:nRead], arg.messageToBeRcvd) {
			t.Errorf("did not received desired message:\n%s != %s", string(buf[:nRead]), string(arg.messageToBeRcvd))
		}
		res := <-byteChan
		if !bytes.Equal(res, arg.messageToBeSent) {
			t.Errorf("result does not match:\n%s != %s", string(res), string(arg.messageToBeSent))
		}
	}
}

// func TestRunServer(t *testing.T) {
// 	ln, err := net.Listen("tcp", "localhost:13131")
// 	if err != nil {
// 		t.Errorf("failed to start listener:\n%s", err.Error())
// 	}
// 	defer ln.Close()
// 	var byteChan chan []byte
// 	go RunServer(ln, byteChan, false)

// 	buf := make([]byte, 1024)

// 	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
// 	contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
// 	contract.SignContract(senderPrivateKey)
// 	serializedContract, err := contract.Serialize()

// 	var contractMessage []byte
// 	contractMessage = append(contractMessage, producer.SecretBytes...)
// 	contractMessage = append(contractMessage, 1)
// 	contractMessage = append(contractMessage, serializedContract...)

// 	type serverTest struct {
// 		name            string
// 		messageToBeSent []byte
// 		messageToBeRcvd []byte
// 	}
// 	testArgs := []serverTest{
// 		{
// 			name:            "Standard Message",
// 			messageToBeSent: []byte("test message"),
// 			messageToBeRcvd: []byte("test message"),
// 		},
// 		{
// 			name:            "Aurum Message",
// 			messageToBeSent: producer.SecretBytes,
// 			messageToBeRcvd: []byte("aurum client acknowledged"),
// 		},
// 		{
// 			name:            "Contract Message",
// 			messageToBeSent: contractMessage,
// 			messageToBeRcvd: []byte("received contract message"),
// 		},
// 	}

// 	for _, ta := range testArgs {
// 		t.Run(ta.name, func(t *testing.T) {
// 			conn, err := net.Dial("tcp", "localhost:13131")
// 			if err != nil {
// 				t.Logf("failed to connect to server:\n%s", err.Error())
// 				t.FailNow()
// 			}
// 			_, err = conn.Write(ta.messageToBeSent)
// 			if err != nil {
// 				t.Logf("failed to send message to server:\n%s", err.Error())
// 				t.FailNow()
// 			}
// 			nRcvd, err := conn.Read(buf)
// 			if err != nil {
// 				t.Logf("failed to receive bytes from connection:\n%s", err.Error())
// 				t.FailNow()
// 			}
// 			if !bytes.Equal(buf[:nRcvd], ta.messageToBeRcvd) {
// 				t.Logf("messages don't match: %s != %s", string(buf[:nRcvd]), string(ta.messageToBeRcvd))
// 				t.FailNow()
// 			}
// 			// Below code tests the channel communication
// 			// if ta.name == "Contract Message" {
// 			// 	channelledContract := <-byteChan
// 			// 	if !bytes.Equal(channelledContract, serializedContract) {
// 			// 		t.Errorf("channel contents does not match desired slice:\n%v != %v", channelledContract, serializedContract)
// 			// 	}
// 			// }
// 		})

// 	}
// }
