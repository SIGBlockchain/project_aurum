package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/privatekey"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

// SetupWallet initializes a JSON file called "aurum_wallet.json"
// with the hex encoded privatekey, balance, and nonce
func SetupWallet() error {
	// if the JSON file already exists, return error
	_, err := os.Stat("aurum_wallet.json")
	if err == nil {
		return errors.New("JSON file for aurum_wallet already exists")
	}

	// Create JSON file for wallet
	file, err := os.Create("aurum_wallet.json")
	if err != nil {
		return err
	}
	defer file.Close()

	// Json structure that will be used to store information into the json file
	type jsonStruct struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}

	// Generate ecdsa key pairs
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// Encodes private key
	pemEncoded, err := privatekey.Encode(privateKey)
	if err != nil {
		return err
	}

	// Encodes the pem encoded private key into string and stores it into a jsonStruct
	hexKey := hex.EncodeToString(pemEncoded)
	j := jsonStruct{PrivateKey: hexKey}

	// Marshall the jsonStruct
	jsonEncoded, err := json.Marshal(j)
	if err != nil {
		return err
	}

	// Write into the json file
	_, err = file.Write(jsonEncoded)
	if err != nil {
		return err
	}

	return nil
}

// GetPrivateKey opens the wallet and returns the private key
func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	// Opens the wallet
	file, err := os.Open("aurum_wallet.json")
	if err != nil {
		return nil, errors.New("Failed to open wallet")
	}
	defer file.Close()

	// Reads the file and stores the data into a byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Failed to read wallet")
	}

	// Json struct for storing the private key from the json file
	type jsonStruct struct {
		PrivateKey string
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	err = json.Unmarshal(data, &j)
	if err != nil {
		return nil, err
	}

	// Decodes the private key from the jsonStruct
	pemEncoded, _ := hex.DecodeString(j.PrivateKey)
	privateKey, err := privatekey.Decode(pemEncoded)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// GetBalance opens the wallet and returns the balance
func GetBalance() (uint64, error) {
	fwallet, err := os.Open("aurum_wallet.json")
	if err != nil {
		return 0, errors.New("Failed to open wallet file: " + err.Error())
	}
	defer fwallet.Close()

	jsonEncoded, err := ioutil.ReadAll(fwallet)
	if err != nil {
		return 0, errors.New("Failed to read wallet file: " + err.Error())
	}

	type jsonStruct struct {
		Balance uint64
	}

	var j jsonStruct
	err = json.Unmarshal(jsonEncoded, &j)
	if err != nil {
		return 0, errors.New("Failed to parse data from json file: " + err.Error())
	}

	return j.Balance, nil
}

// GetStateNonce opens the wallet and returns the state nonce
func GetStateNonce() (uint64, error) {
	// Opens the wallet
	file, err := os.Open("aurum_wallet.json")
	if err != nil {
		return 0, errors.New("Failed to open wallet")
	}
	defer file.Close()

	// Reads the json file and stores the data into the data byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, errors.New("Failed to read wallet")
	}

	// Json struct for storing the nonce from the json file
	type jsonStruct struct {
		Nonce uint64
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	err = json.Unmarshal(data, &j)
	if err != nil {
		return 0, errors.New("Failed to parse data from json file: " + err.Error())
	}

	return j.Nonce, nil
}

// GetWalletAddress opens the wallet and returns the wallet address
func GetWalletAddress() ([]byte, error) {
	// Opens the wallet
	file, err := os.Open("aurum_wallet.json")
	if err != nil {
		return nil, errors.New("Failed to open wallet")
	}
	defer file.Close()

	// Reads the json file and stores the data into a byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Failed to read wallet")
	}

	// Json struct for storing the data from the json file
	type jsonStruct struct {
		PrivateKey string
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	err = json.Unmarshal(data, &j)
	if err != nil {
		return nil, errors.New("Failed to parse data from json file")
	}

	// Get the private key hash
	privKeyHash, err := hex.DecodeString(j.PrivateKey)
	if err != nil {
		return nil, errors.New("Failed to decode private key string")
	}

	// Get the private key
	privKey, err := privatekey.Decode(privKeyHash)
	if err != nil {
		return nil, errors.New("Failed to decode private key hash")
	}

	// Get the PEM encoded public key
	pubKeyEncoded, err := publickey.Encode(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}
	return hashing.New(pubKeyEncoded), nil
}

func UpdateWallet(balance, stateNonce uint64) error {
	wallet := "aurum_wallet.json"
	if _, err := os.Stat(wallet); os.IsNotExist(err) {
		return errors.New("wallet file not detected: " + err.Error())
	}
	type walletData struct {
		PrivateKey string
		Balance    uint64
		Nonce      uint64
	}
	f, err := os.Open(wallet)
	if err != nil {
		return errors.New("failed to open wallet: " + err.Error())
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	var jsonData walletData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return errors.New("failed to unmarshall data: %s" + err.Error())
	}
	jsonData.Balance = balance
	jsonData.Nonce = stateNonce

	dumpData, err := json.Marshal(jsonData)
	if err != nil {
		return errors.New("failed to marshal dump data: " + err.Error())
	}
	if err := ioutil.WriteFile(wallet, dumpData, 0644); err != nil {
		return errors.New("failed to write to file: " + err.Error())
	}

	return nil
}
