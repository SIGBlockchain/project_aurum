// This contains the tools to allow a user to send relevant information off to a producer
//
// It also contains the functions for accepting user input and displaying information, a primitive console UI
package client

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/pkg/privatekey"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"
)

var SecretBytes = hashing.New([]byte("aurum"))[8:16]

// This will check if the client is connected to the internet
//
// Will return relevant error if not connected
func CheckConnection() error {
	// Creates a connection conn, and stores any errors in err
	conn, err := net.Dial("tcp", "www.google.com:80")
	// If err is not nil, then there was an error
	if err != nil {
		return errors.New("Connectivity check failed.")
	}
	// Close the connection
	conn.Close()
	return nil
}

// This stores any user input inside of text until a new line is entered
func GetUserInput(text *string, reader io.Reader) error {
	// Creates a reader object, using bufio library
	fmt.Print("[aurum client] >> ")
	// Stores user input until \n, stores into text
	var err error
	newReader := bufio.NewReader(reader)
	*text, err = newReader.ReadString('\n')
	// Ensures no newline characters in input
	*text = strings.Replace(*text, "\n", "", -1)
	return err
}

// Establishes connection to addr with Dial and sends data to address, returns number of bytes written and any errors
func SendToProducer(buf []byte, addr string) (int, error) {
	// Opens a connection, if connection fails, return 0 and error
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		conn.Close()
		return 0, err
	}
	n, err := conn.Write(buf)
	// Close connection, return counter and no error
	conn.Close()
	return n, nil
}

// This clears the console screen of any present text, uses system dependent clear sceen command
//
// This will also place a new seperator at the top of the screen
func ClearScreen() {
	// On non-windows systems, the clear command clears the screen
	cmd := exec.Command("clear")
	// If the operating system is actually windows, change this to cls (clear screen)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	}
	// Sets the output of this command to the command line, and executes
	cmd.Stdout = os.Stdout
	cmd.Run()
	// Prints a seperator at the top of the screen
	fmt.Println("#############################################################################")
	/*==ALTERNATIVE OPTIONS====================================================================
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("=============================================================================")
	fmt.Println("-----------------------------------------------------------------------------")
	fmt.Println("_____________________________________________________________________________")
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	==========================================================================================*/
}

// This will print all relevant commands into the console
func PrintHelp() {
	fmt.Println("\tbalance\t\t\t\tcheck your wallet balance")                         // balance
	fmt.Println("\tclear\t\t\t\tclears the screen of all previous output")            // clear
	fmt.Println("\thelp\t\t\t\tprints all avalible commands and description of each") // help
	fmt.Println("\tmoreinfo\t\t\tprints link to project_aurum github page")           // moreinfo
	fmt.Println("\tsend [recipient] [value]\tsend aurum to using their public key")   // send
	fmt.Println("\tq\t\t\t\tquits the program")                                       // q
}

// This will print a link to the project github page into the console
func PrintGithubLink() {
	fmt.Println("https://github.com/SIGBlockchain/project_aurum for more info")
}

// This will initialize a JSON file called "aurum_wallet.json"
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

// Opens the wallet file and prints out the balance
func CheckBalance() error {
	// Opens the wallet
	file, err := os.Open("aurum_wallet.json")
	if err != nil {
		return errors.New("Failed to open wallet")
	}
	defer file.Close()

	// Reads the json file and stores the data into a byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.New("Failed to read wallet")
	}

	// Json struct for storing the balance from the json file
	type jsonStruct struct {
		Balance uint64
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	err = json.Unmarshal(data, &j)
	if err != nil {
		return err
	}

	// Prints balance
	fmt.Print("Your balance: ", j.Balance, " AUR\n")
	return nil
}

// Opens the wallet and prints out the public key and its hash in hex encoded form
func PrintPublicKeyAndHash() error {
	// Opens the wallet
	file, err := os.Open("aurum_wallet.json")
	if err != nil {
		return errors.New("Failed to open wallet")
	}
	defer file.Close()

	// Reads the json file and stores the data into a byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.New("Failed to read wallet")
	}

	// Json struct for storing the private key from the json file
	type jsonStruct struct {
		PrivateKey string
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	json.Unmarshal(data, &j)

	// Decodes the private key from the jsonStruct
	pemEncoded, _ := hex.DecodeString(j.PrivateKey)
	privateKey, err := privatekey.Decode(pemEncoded)
	if err != nil {
		return err
	}

	// Gets and encodes the public key
	publicKey := privateKey.PublicKey
	pemEncodedPub := publickey.Encode(&publicKey)

	// Encodes the public key into a string and a hash
	publicKeyString := hex.EncodeToString(pemEncodedPub)
	publicKeyHash := hashing.New(pemEncodedPub)

	// Encodes the public key hash into string to print
	publicKeyHashString := hex.EncodeToString(publicKeyHash)

	// Prints public key and its hash
	fmt.Print("Public Key: ", publicKeyString, "\nHashed Key: ", publicKeyHashString, "\n")
	return nil
}

// Opens the wallet and returns the private key
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

// READY TO BE IMPLEMENTED
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

// READY TO BE IMPLEMENTED
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

// READY TO BE IMPLEMENTED
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
	pubKeyEncoded := publickey.Encode(&privKey.PublicKey)
	return hashing.New(pubKeyEncoded), nil
}

func RequestWalletInfo(producerAddr string) (accounts.AccountInfo, error) {
	var accInfo accounts.AccountInfo
	walletAddress, err := GetWalletAddress()
	if err != nil {
		return accInfo, errors.New("failed to get wallet address: " + err.Error())
	}
	var requestInfoMessage []byte
	requestInfoMessage = append(requestInfoMessage, SecretBytes...)
	requestInfoMessage = append(requestInfoMessage, 2)
	requestInfoMessage = append(requestInfoMessage, walletAddress...)
	conn, err := net.Dial("tcp", producerAddr)
	if err != nil {
		return accInfo, errors.New("failed to connect to producer: " + err.Error())
	}
	if _, err := conn.Write(requestInfoMessage); err != nil {
		return accInfo, errors.New("failed to send message to producer: " + err.Error())
	}
	// Should receive Thank you first
	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		return accInfo, errors.New("failed to get thank you message: " + err.Error())
	}
	// Should receive message next
	buf = make([]byte, 1024)
	nRead, err := conn.Read(buf)
	if err != nil {
		return accInfo, errors.New("failed to get response message: " + err.Error())
	}
	if buf[8] == 1 {
		return accInfo, errors.New("got back failure message from producer")
	} else if buf[8] == 0 {
		if err := accInfo.Deserialize(buf[9:nRead]); err != nil {
			return accounts.AccountInfo{}, errors.New("failed to deserialize account info: " + err.Error())
		}
	}
	return accInfo, nil
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
