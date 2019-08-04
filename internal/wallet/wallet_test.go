package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/privatekey"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

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
