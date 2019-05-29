package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/pkg/keys"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
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
			if err := SetupWallet(); (err != nil) != tt.wantErr {
				t.Errorf("SetupWallet() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer func() {
				err := os.Remove("aurum_wallet.json")
				if err != nil {
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
		})
	}
}

func TestGetBalance(t *testing.T) {
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
			type balanceData struct {
				Balance    uint64
			}
			wallet, err := os.Open("aurum_wallet.json")
			if err != nil {
				t.Errorf("Failed to open wallet: %s", err)
			}
			defer wallet.Close()
			bytes, _ := ioutil.ReadAll(wallet)
			var bal balanceData
			err = json.Unmarshal(bytes, &bal)
			if err != nil {
				t.Errorf("Failed to unmarshall JSON data: %s", err)
			}
			captureStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Errorf("Pipe function failed: %s", err)
			}
			os.Stdout = w
			if err := getBalance(); (err != nil) != tt.wantErr {
				t.Errorf("getBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
			w.Close()
			out, err := ioutil.ReadAll(r)
			if err != nil {
				t.Errorf("Failed to read from capture file: %s", err)
			}
			os.Stdout = captureStdout
			expected := "0\n"
			if string(out) != expected {
				t.Errorf("Print statement incorrect. Wanted: %s, got %s", expected, string(out))
			}
		})
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
			publicKeyString := hex.EncodeToString(keys.EncodePublicKey(&publicKey))
			publicKeyHash := block.HashSHA256(keys.EncodePublicKey(&publicKey))
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
