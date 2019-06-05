package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
			runCommand: []string{"go", "run", "main.go", "-d", "-g", "--supply=100"},
		},
		// {
		// 	name:       "Loop",
		// 	runCommand: []string{"go", "run", "main.go", "-d", "--interval=1000ms", "--blocks=3"},
		// },
		{
			name:       "Bytechannel",
			runCommand: []string{"go", "run", "main.go", "-d", "--interval=3000ms", "-t", "--port=13131"},
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
			if arg.name == "Bytechannel" {
				// senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				// recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				// recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
				// contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
				// contract.SignContract(senderPrivateKey)
				// serializedContract, _ := contract.Serialize()

				// var contractMessage []byte
				// contractMessage = append(contractMessage, producer.SecretBytes...)
				// contractMessage = append(contractMessage, 1)
				// contractMessage = append(contractMessage, serializedContract...)
				_, err := net.Dial("tcp", "localhost:13131")
				if err != nil {
					t.Errorf("failed to connect to server:\n%s", err.Error())
				}

				// _, err = conn.Write(contractMessage)
				// if err != nil {
				// 	t.Errorf("failed to send message")
				// }
				// buf := make([]byte, 1024)
				// nRead, err := conn.Read(buf)
				// if err != nil {
				// 	t.Errorf("failed to read from connections:\n%s", err.Error())
				// } else {
				// 	t.Logf("Buff: %s", string(buf))
				// }
				// if !bytes.Equal(buf[:nRead], []byte("Thank you.\n")) {
				// 	t.Errorf("did not received desired message:\n%s != %s", string(buf[:nRead]), string([]byte("Thank you.\n")))
				// }
			}
			if err := cmd.Wait(); err != nil {
				t.Errorf("main call returned with: %s.", err.Error())
				t.Logf("Stdout: %s", string(stdout.Bytes()))
			}
		})
	}
	t.Logf("%s", string(stdout.Bytes()))

	// dbc, err := sql.Open("sqlite3", "metadata.tab")
	// if err != nil {
	// 	t.Errorf("failed to open sqlite database")
	// }
	// defer func() {
	// 	if err := dbc.Close(); err != nil {
	// 		t.Errorf("failed to close database connection")
	// 	}
	// }()
	// var count int
	// var expectedCount = 5
	// if err := dbc.QueryRow("SELECT COUNT(*) FROM METADATA").Scan(&count); err != nil {
	// 	t.Errorf("failed to query rows")
	// }
	// if count != expectedCount {
	// 	t.Errorf("invalid number of blocks: %d != %d", count, expectedCount)
	// }
	// youngestBlock, err := blockchain.GetYoungestBlock(ledger, metadataTable)
	// if err != nil {
	// 	t.Errorf("failed to get youngest block:\n%s", err.Error())
	// }
	// t.Logf("%v", youngestBlock)
	// data := youngestBlock.Data[0]
	// var compContract accounts.Contract
	// if err := compContract.Deserialize(data); err != nil {
	// 	t.Errorf("failed to deserialize data:\n%s", err.Error())
	// }
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
			name:            "Regular message",
			messageToBeSent: []byte("hello\n"),
			messageToBeRcvd: []byte("No thanks.\n"),
		},
		{
			name:            "Aurum message",
			messageToBeSent: producer.SecretBytes,
			messageToBeRcvd: []byte("Thank you.\n"),
		},
		{
			name:            "Contract message",
			messageToBeSent: contractMessage,
			messageToBeRcvd: []byte("Thank you.\n"),
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
		if arg.name != "Regular message" {
			res := <-byteChan
			if !bytes.Equal(res, arg.messageToBeSent) {
				t.Errorf("result does not match:\n%s != %s", string(res), string(arg.messageToBeSent))
			}
			if arg.name == "Contract message" {
				var contract accounts.Contract
				if err := contract.Deserialize(res[9:]); err != nil {
					t.Errorf("failed to deserialize contract:\n%s", err.Error())
				}
				if !bytes.Equal(res[9:], serializedContract) {
					t.Errorf("serialized contracts do not match:\n%v != %v", res[9:], serializedContract)
				}
			}
		}
	}
}
