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

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

var removeFiles = true

func TestSuite(t *testing.T) {
	// producer.GenerateGenesisHashFile(25)
	defer func() { // Function is dangerous, consider only running with flag
		if removeFiles {
			// if err := os.Remove("genesis_hashes.txt"); err != nil {
			// 	t.Errorf("failed to remove genesis hashes:\n%s", err.Error())
			// }
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
	var expectedCount = 4
	if err := dbc.QueryRow("SELECT COUNT(*) FROM METADATA").Scan(&count); err != nil {
		t.Errorf("failed to query rows")
	}
	if count != expectedCount {
		t.Errorf("invalid number of blocks: %d != %d", count, expectedCount)
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

func TestByteChannel(t *testing.T) {
	genesisHashes, err := producer.ReadGenesisHashes()
	if err != nil {
		t.Errorf("failed to read genesis hashes:\n%s", err.Error())
	}
	genesisBlock, err := producer.BringOnTheGenesis(genesisHashes, 1000)
	if err != nil {
		t.Errorf("failed to create genesis block:\n%s", err.Error())
	}
	if err := producer.Airdrop(ledger, metadataTable, genesisBlock); err != nil {
		t.Errorf("failed to perform air drop:\n%s", err.Error())
	}
	defer func() {
		if removeFiles {
			if err := os.Remove("blockchain.dat"); err != nil {
				t.Errorf("failed to remove blockchain.dat:\n%s", err.Error())
			}
			if err := os.Remove("metadata.tab"); err != nil {
				t.Errorf("failed to remove metadatata.tab:\n%s", err.Error())
			}
		}
	}()
	ln, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to start server:\n%s", err.Error())
	}
	byteChan := make(chan []byte)
	debug := false

	go RunServer(ln, byteChan, debug)
	testMode := true
	prodInterval := "2000ms"
	memStats := false
	fl := Flags{
		debug:       &debug,
		interval:    &prodInterval,
		test:        &testMode,
		memoryStats: &memStats,
	}
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
	contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.SignContract(senderPrivateKey)
	serializedContract, _ := contract.Serialize()

	var contractMessage []byte
	contractMessage = append(contractMessage, producer.SecretBytes...)
	contractMessage = append(contractMessage, 1)
	contractMessage = append(contractMessage, serializedContract...)

	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	_, err = conn.Write(contractMessage)
	if err != nil {
		t.Errorf("failed to send message")
	}
	ProduceBlocks(byteChan, fl, true)

	youngestBlock, err := blockchain.GetYoungestBlock(ledger, metadataTable)
	if err != nil {
		t.Errorf("failed to get youngest block:\n%s", err.Error())
	}
	data := youngestBlock.Data[0]
	var compContract accounts.Contract
	if err := compContract.Deserialize(data); err != nil {
		t.Errorf("failed to deserialize data:\n%s", err.Error())
	}
	if !bytes.Equal(serializedContract, data) {
		t.Errorf("data does not match:\n%s != %s", string(serializedContract), string(data))
	}
}

func TestResponseToAccountInfoRequest(t *testing.T) {
	if err := client.SetupWallet(); err != nil {
		t.Errorf("failed to setup wallet:\n%s", err.Error())
	}
	defer func() {
		if err := os.Remove("aurum_wallet.json"); err != nil {
			t.Errorf("failed to remove aurum_wallet.json:\n%s", err.Error())
		}
	}()
	dbName := "accountBalanceTable.tab"
	dbc, _ := sql.Open("sqlite3", dbName)
	defer func() {
		err := dbc.Close()
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
		err = os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()
	walletAddress, err := client.GetWalletAddress()
	// t.Logf("Wallet address: %v", walletAddress)
	if err != nil {
		t.Errorf("failed to retrieve wallet address:\n%s", err.Error())
	}
	ln, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to start server:\n%s", err.Error())
	}
	byteChan := make(chan []byte)
	debug := false

	go RunServer(ln, byteChan, debug)

	// Request
	var requestInfoMessage []byte
	requestInfoMessage = append(requestInfoMessage, producer.SecretBytes...)
	requestInfoMessage = append(requestInfoMessage, 2)
	requestInfoMessage = append(requestInfoMessage, walletAddress...)
	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	// t.Logf("Sending message: %v", requestInfoMessage)
	if _, err := conn.Write(requestInfoMessage); err != nil {
		t.Errorf("failed to send request info message:\n%s", err.Error())
	}
	buf := make([]byte, 1024)
	nRead, err := conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if !bytes.Equal(buf[:nRead], []byte("Thank you.\n")) {
		t.Errorf("expected different response: %v != %v", string(buf[:nRead]), string([]byte("Thank you\n")))
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if buf[8] != 1 {
		t.Errorf("failed to get errored response from producer")
	}
	conn.Close()

	// Check for successful insertion
	if err := accounts.InsertAccountIntoAccountBalanceTable(dbc, walletAddress, 1000); err != nil {
		t.Errorf("failed to insert sender account")
	}
	_, err = accounts.GetAccountInfo(walletAddress)
	if err != nil {
		t.Errorf("failed to retrieve account info:\n%s", err.Error())
	}
	// t.Logf("account info: %v", accInfo)
	dbc.Close()

	// New request
	conn, err = net.Dial("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	// t.Logf("Sending message: %v", requestInfoMessage)
	if _, err := conn.Write(requestInfoMessage); err != nil {
		t.Errorf("failed to send request info message:\n%s", err.Error())
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if !bytes.Equal(buf[:nRead], []byte("Thank you.\n")) {
		t.Errorf("expected different response: %v != %v", string(buf[:nRead]), string([]byte("Thank you\n")))
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}

	if buf[8] != 0 {
		t.Errorf("failed to get success response from producer: %d != %d", buf[8], 0)
	}
	var accInfo accounts.AccountInfo
	if err := accInfo.Deserialize(buf[9:nRead]); err != nil {
		t.Errorf("failed to deserialize account info:\n%s", err.Error())
	}
}
