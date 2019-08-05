package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/constants"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/privatekey"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"

	producer "github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
)

// Test will fail in airplane mode, or just remove wireless connection.
func TestCheckConnection(t *testing.T) {
	err := CheckConnection()
	if err != nil {
		t.Errorf("Internet connection check failed.")
	}
}

// Test will simulate user input and ensure that the function will collect the correct string
func TestGetUserInput(t *testing.T) {

	var testread bytes.Buffer
	testread.Write([]byte("TEST\n"))

	var user_input string
	if GetUserInput(&user_input, &testread) != nil {
		t.Errorf("User Input Check Failed.")
	}

	if user_input != "TEST" {
		t.Errorf("User Input Check Failed.")
	}
}

// Test send to producer with small max length message for one send
func TestSendToProducer(t *testing.T) {
	sz := 1024
	testbuf := make([]byte, sz)
	for i, _ := range testbuf {
		testbuf[i] = 1
	}
	addr := "localhost:8080"
	ln, err := net.Listen("tcp", addr)
	var buffer bytes.Buffer
	bp := producer.BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        log.New(&buffer, "LOG:", log.Ldate),
	}
	go bp.AcceptConnections()
	time.Sleep(1)
	if err != nil {
		t.Errorf("Failed to set up listener")
	}
	n, err := SendToProducer(testbuf, addr)
	if err != nil {
		t.Errorf("Failed to send to producer")
	}
	if n != sz {
		t.Errorf("Did not write all bytes to connection")
	}
	ln.Close()
}

func TestSetupWallet(t *testing.T) {
	if err := SetupWallet(); err != nil {
		t.Errorf("SetupWallet() returned error")
	}
	defer func() {
		if err := os.Remove("aurum_wallet.json"); err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	if _, err := os.Stat("aurum_wallet.json"); os.IsNotExist(err) {
		t.Errorf("\"aurum_wallet.json\" does not exist. Error: %s", err)
	}
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	wallet, err := os.Open("aurum_wallet.json")
	if err != nil {
		t.Errorf("Failed to open wallet: %s", err)
	}
	defer wallet.Close()
	bytes, _ := ioutil.ReadAll(wallet)
	var wd walletData
	err = json.Unmarshal(bytes, &wd)
	if err != nil {
		t.Errorf("Failed to unmarshall JSON data: %s", err)
	}
	if wd.Balance != 0 {
		t.Errorf("Incorrect balance. Want %d, got %d", 0, wd.Balance)
	}
	if wd.Nonce != 0 {
		t.Errorf("Incorrect nonce. Want %d, got %d", 0, wd.Nonce)

	}
	privateKeyString, err := hex.DecodeString(wd.PrivateKey)
	if err != nil {
		t.Errorf("Failed to decode private key: %s", err)
	}
	pemDecodedKey, _ := pem.Decode(privateKeyString)
	x509Encoded := pemDecodedKey.Bytes
	_, err = x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		t.Errorf("Failed to parse private key: %s", err)
	}
	if err := SetupWallet(); err == nil {
		t.Errorf("supposed to cause error when attempting to call SetupWallet() when wallet already exists")
	}
}

func TestCheckBalance(t *testing.T) {
	if err := SetupWallet(); err != nil {
		t.Errorf("Failed to setup wallet: %s", err)
	}
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type walletData struct {
				PrivateKey string
				Balance    uint64
				Nonce      uint64
			}
			wallet, err := os.Open("aurum_wallet.json")
			if err != nil {
				t.Errorf("Failed to open wallet: %s", err)
			}
			defer wallet.Close()
			bytes, _ := ioutil.ReadAll(wallet)
			var wd walletData
			err = json.Unmarshal(bytes, &wd)
			if err != nil {
				t.Errorf("Failed to unmarshall JSON data: %s", err)
			}
			captureStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Errorf("Pipe function failed: %s", err)
			}
			os.Stdout = w
			if err := CheckBalance(); (err != nil) != tt.wantErr {
				t.Errorf("CheckBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
			w.Close()
			out, err := ioutil.ReadAll(r)
			if err != nil {
				t.Errorf("Failed to read from capture file: %s", err)
			}
			os.Stdout = captureStdout
			expected := "Your balance: 0 AUR\n"
			if string(out) != expected {
				t.Errorf("Print statement incorrect. Wanted: %s, got %s", expected, string(out))
			}
		})
	}
}

func TestPrintPublicKeyAndHash(t *testing.T) {
	if err := SetupWallet(); err != nil {
		t.Errorf("Failed to setup wallet: %s", err)
	}
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type walletData struct {
				PrivateKey string
				Balance    uint64
				Nonce      uint64
			}
			wallet, err := os.Open("aurum_wallet.json")
			if err != nil {
				t.Errorf("Failed to open wallet: %s", err)
			}
			defer wallet.Close()
			bytes, _ := ioutil.ReadAll(wallet)
			var wd walletData
			err = json.Unmarshal(bytes, &wd)
			if err != nil {
				t.Errorf("Failed to unmarshall JSON data: %s", err)
			}
			privateKeyString, err := hex.DecodeString(wd.PrivateKey)
			if err != nil {
				t.Errorf("Failed to decode private key: %s", err)
			}
			pemDecodedKey, _ := pem.Decode(privateKeyString)
			x509Encoded := pemDecodedKey.Bytes
			privateKey, err := x509.ParseECPrivateKey(x509Encoded)
			publicKey := privateKey.PublicKey
			publicKeyString := hex.EncodeToString(publickey.Encode(&publicKey))
			publicKeyHash := hashing.New(publickey.Encode(&publicKey))
			publicKeyHashString := hex.EncodeToString(publicKeyHash)
			if err != nil {
				t.Errorf("Failed to parse private key: %s", err)
			}
			captureStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Errorf("Pipe function failed: %s", err)
			}
			os.Stdout = w
			if err := PrintPublicKeyAndHash(); (err != nil) != tt.wantErr {
				t.Errorf("PrintPublicKeyAndHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			w.Close()
			out, err := ioutil.ReadAll(r)
			if err != nil {
				t.Errorf("Failed to read from capture file: %s", err)
			}
			os.Stdout = captureStdout
			expected := fmt.Sprintf("Public Key: %s\nHashed Key: %s\n", publicKeyString, publicKeyHashString)
			if string(out) != expected {
				t.Errorf("Print statement incorrect. Wanted: %s, got %s", expected, string(out))
			}
		})
	}
}

func TestGetPrivateKey(t *testing.T) {
	SetupWallet()
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	wallet, err := os.Open("aurum_wallet.json")
	if err != nil {
		t.Errorf("Failed to open wallet: %s", err)
	}
	defer wallet.Close()
	bytes, _ := ioutil.ReadAll(wallet)
	var wd walletData
	err = json.Unmarshal(bytes, &wd)
	if err != nil {
		t.Errorf("Failed to unmarshall JSON data: %s", err)
	}
	privateKeyString, err := hex.DecodeString(wd.PrivateKey)
	if err != nil {
		t.Errorf("Failed to decode private key: %s", err)
	}
	pemDecodedKey, _ := pem.Decode(privateKeyString)
	x509Encoded := pemDecodedKey.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	tests := []struct {
		name    string
		want    *ecdsa.PrivateKey
		wantErr bool
	}{
		{
			want:    privateKey,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPrivateKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWalletAddress(t *testing.T) {
	SetupWallet()
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	wallet, err := os.Open("aurum_wallet.json")
	if err != nil {
		t.Errorf("Failed to open wallet: %s", err)
	}
	defer wallet.Close()
	myBytes, _ := ioutil.ReadAll(wallet)
	var wd walletData
	err = json.Unmarshal(myBytes, &wd)
	if err != nil {
		t.Errorf("Failed to unmarshall JSON data: %s", err)
	}
	privateKeyString, err := hex.DecodeString(wd.PrivateKey)
	if err != nil {
		t.Errorf("Failed to decode private key: %s", err)
	}
	pemDecodedKey, _ := pem.Decode(privateKeyString)
	x509Encoded := pemDecodedKey.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	publicKey := privateKey.PublicKey
	publicKeyHash := hashing.New(publickey.Encode(&publicKey))
	if err != nil {
		t.Errorf("Failed to parse private key: %s", err)
	}
	addr, err := GetWalletAddress()
	if (err != nil) != false {
		t.Errorf("GetWalletAddress() error = %v, wantErr %v", err, false)
	}
	var expected = publicKeyHash
	if !bytes.Equal(expected, addr) {
		t.Logf(hex.EncodeToString(expected))
		t.Errorf("Values fail to match. Wanted: %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(addr))
	}
}

func TestGetStateNonce(t *testing.T) {
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	// Create JSON file for wallet
	file, err := os.Create("aurum_wallet.json")
	if err != nil {
		t.Errorf("Failed to create \"aurum_wallet.json\". Error: %s", err)
	}
	defer file.Close()
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	var wd walletData
	wd.Nonce = rand.Uint64()
	// Marshall the jsonStruct
	jsonEncoded, err := json.Marshal(wd)
	if err != nil {

		t.Errorf("Failed to marshal the wallet for the test. Error: %s", err)
	}
	// Write into the json file
	_, err = file.Write(jsonEncoded)
	if err != nil {
		t.Errorf("Failed to write into the json file. Error: %s", err)
	}

	myNonce, err := GetStateNonce()
	if err != nil {
		t.Errorf("getNonce() error = %v, wantErr %v", err, false)
	}
	var expected = wd.Nonce
	if !reflect.DeepEqual(expected, myNonce) {
		t.Errorf("Values fail to match. Wanted: %v, got %v", expected, myNonce)
	}
}

func TestGetBalance(t *testing.T) {
	defer func() {
		err := os.Remove("aurum_wallet.json")
		if err != nil {
			t.Errorf("Failed to remove \"aurum_wallet.json\". Error: %s", err)
		}
	}()
	// Create JSON file for wallet
	file, err := os.Create("aurum_wallet.json")
	if err != nil {
		t.Errorf("Failed to create \"aurum_wallet.json\". Error: %s", err)
	}
	defer file.Close()
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	var wd walletData
	wd.Balance = rand.Uint64()

	// Marshall the jsonStruct
	jsonEncoded, err := json.Marshal(wd)
	if err != nil {

		t.Errorf("Failed to marshal the wallet for the test. Error: %s", err)
	}
	// Write into the json file
	_, err = file.Write(jsonEncoded)
	if err != nil {
		t.Errorf("Failed to write into the json file. Error: %s", err)
	}

	myBal, err := GetBalance()
	if err != nil {
		t.Errorf("getBalance() error = %v, wantErr %v", err, false)
	}
	var expected = wd.Balance
	if !reflect.DeepEqual(expected, myBal) {
		t.Errorf("Values fail to match. Wanted: %v, got %v", expected, myBal)
	}
}

func TestUpdateWallet(t *testing.T) {
	wallet := "aurum_wallet.json"
	if err := SetupWallet(); err != nil {
		t.Errorf("failed to set up wallet: %s", err.Error())
	}
	defer func() {
		if err := os.Remove(wallet); err != nil {
			t.Errorf("failed to remove wallet: %s", err.Error())
		}
	}()
	private, err := GetPrivateKey()
	if err != nil {
		t.Errorf("failed to get private key: %s", err.Error())
	}
	currentBalance, err := GetBalance()
	if err != nil {
		t.Errorf("failed to get initial balance: %s", err.Error())
	}
	if currentBalance != 0 {
		t.Errorf("current balance not what was expected: %d != %d", currentBalance, 0)
	}
	currentNonce, err := GetStateNonce()
	if err != nil {
		t.Errorf("failed to get initial nonce: %s", err.Error())
	}
	if currentNonce != 0 {
		t.Errorf("current nonce not what was expected: %d != %d", currentNonce, 0)
	}
	if err := UpdateWallet(100000, 15); err != nil {
		t.Errorf("failed to update wallet: %s", err.Error())
	}
	newBalance, err := GetBalance()
	if err != nil {
		t.Errorf("failed to get new balance: %s", err.Error())
	}
	if newBalance != 100000 {
		t.Errorf("new balance not what was expected: %d != %d", newBalance, 100000)
	}
	newNonce, err := GetStateNonce()
	if err != nil {
		t.Errorf("failed to get initial nonce: %s", err.Error())
	}
	if newNonce != 15 {
		t.Errorf("new nonce not what was expected: %d != %d", currentNonce, 15)
	}
	// Check if private key was unchanged
	unchangedPrivate, err := GetPrivateKey()
	if err != nil {
		t.Errorf("failed to get private key second time: %s", err.Error())
	}
	encPrivExpected, err := privatekey.Encode(private)
	if err != nil {
		t.Errorf("failed to encode private key: %s", err.Error())
	}
	encPrivActual, err := privatekey.Encode(unchangedPrivate)
	if err != nil {
		t.Errorf("failed to encode private key: %s", err.Error())
	}
	if !bytes.Equal(encPrivActual, encPrivExpected) {
		t.Errorf("private keys do not match: %v != %v", encPrivActual, encPrivExpected)
	}
}

func TestRequestWalletInfo(t *testing.T) {
	if err := SetupWallet(); err != nil {
		t.Errorf("failed to setup wallet:\n%s", err.Error())
	}
	defer func() {
		if err := os.Remove("aurum_wallet.json"); err != nil {
			t.Errorf("failed to remove aurum_wallet.json:\n%s", err.Error())
		}
	}()

	dbName := constants.AccountsTable
	dbc, _ := sql.Open("sqlite3", dbName)
	defer func() {
		err := dbc.Close()
		if err != nil {
			t.Errorf("Failed to close database: %s", err)
		}
		err = os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	_, err := dbc.Exec("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	if err != nil {
		t.Errorf("Failed to create table in database: %s", err)
	}

	walletAddress, err := GetWalletAddress()
	if err != nil {
		t.Errorf("failed to retrieve wallet address:\n%s", err.Error())
	}

	ln, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		t.Errorf("failed to start server:\n%s", err.Error())
	}
	byteChan := make(chan []byte)
	debug := false

	go producer.RunServer(ln, byteChan, debug)

	tests := []struct {
		name          string
		expectedBal   uint64
		expectedNonce uint64
		wantErr       bool
	}{
		{
			name:          "Wallet address not in table",
			expectedBal:   0,
			expectedNonce: 0,
			wantErr:       true,
		},
		{
			name:          "Wallet address in table",
			expectedBal:   15,
			expectedNonce: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Wallet address in table" {
				err = accountstable.InsertAccountIntoAccountBalanceTable(dbc, walletAddress, tt.expectedBal)
				if err != nil {
					t.Errorf("failed to insert account into account balance table")
				}
			}

			accountInfo, err := RequestWalletInfo("localhost:10000")
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestWalletInfo() error:\nWantErr: %v\nActualErr: %v", tt.wantErr, err.Error())
			}

			if accountInfo.Balance != tt.expectedBal {
				t.Errorf("Account balance from producer does not match")
			}
			if accountInfo.StateNonce != tt.expectedNonce {
				t.Errorf("Account state nonce from producer does not match")
			}

		})
	}
}
