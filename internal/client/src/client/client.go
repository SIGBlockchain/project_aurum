// This contains the tools to allow a user to send relevant information off to a producer
//
// It also contains the functions for accepting user input and displaying information, a primitive console UI
package client

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
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
	"unsafe"

    "github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"

	keys "github.com/SIGBlockchain/project_aurum/pkg/keys"
)

var secretBytes = block.HashSHA256([]byte("aurum"))[8:16]

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
	pemEncoded, err := keys.EncodePrivateKey(privateKey)
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
	privateKey, err := keys.DecodePrivateKey(pemEncoded)
	if err != nil {
		return err
	}

	// Gets and encodes the public key
	publicKey := privateKey.PublicKey
	pemEncodedPub := keys.EncodePublicKey(&publicKey)

	// Encodes the public key into a string and a hash
	publicKeyString := hex.EncodeToString(pemEncodedPub)
	publicKeyHash := block.HashSHA256(pemEncodedPub)

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
	privateKey, err := keys.DecodePrivateKey(pemEncoded)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// NOT READY TO BE IMPLEMENTED YET
func UpdateWallet() error { return errors.New("Not ready to be implemented yet") }

type ContractRequest struct {
	SecBytes    []byte
	Version     uint16
	MessageType uint16
	Request     *producer.Data
}

func (conReq *ContractRequest) Serialize() ([]byte, error) {
	lengthSecBytes := int(len(conReq.SecBytes))
	lengthVersion := int(unsafe.Sizeof(conReq.Version))
	lengthType := int(unsafe.Sizeof(conReq.MessageType))
	incomingConReqDataBdy, err := conReq.Request.Serialize() //TODO should this be conReq.Request.Bdy.Serialize() instead?
	if err != nil {
		return nil, errors.New("Failed to serialize Contract Request")
	}
	lengthMinusBdy := lengthSecBytes + lengthVersion + lengthType
	serializedConReq := make([]byte, lengthMinusBdy)
	copy(serializedConReq[0:lengthSecBytes], conReq.SecBytes)
	binary.LittleEndian.PutUint16(serializedConReq[lengthSecBytes:lengthSecBytes+lengthVersion], conReq.Version)
	binary.LittleEndian.PutUint16(serializedConReq[lengthSecBytes+lengthVersion:lengthSecBytes+lengthVersion+lengthType], conReq.MessageType)

	serializedConReq = append(serializedConReq, incomingConReqDataBdy...)
	return serializedConReq, nil
}

func (conReq *ContractRequest) Deserialize(serializedRequest []byte) error {
	conReq.SecBytes = serializedRequest[:8]
	conReq.Version = binary.LittleEndian.Uint16(serializedRequest[8:10])
	conReq.MessageType = binary.LittleEndian.Uint16(serializedRequest[10:12])
	conReq.Request = &producer.Data{}
	conReq.Request.Deserialize(serializedRequest[12:])
	
    return nil

}

// Open aurum_wallet.json for private key and nonce
func SendAurum(producerAddr string, clientPrivateKey *ecdsa.PrivateKey, recipientPublicKeyHash []byte, value uint64) error {
    // get nonce TODO functionize into GetNonce()
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
		Nonce      uint64
	}

	// Parse the data from the json file into a jsonStruct
	var j jsonStruct
	err = json.Unmarshal(data, &j)
	if err != nil {
		return err
	}
    // end get nonce

    newContract, err := accounts.MakeContract(1, clientPrivateKey, recipientPublicKeyHash, value, 1)

    if err != nil {
        return err
    }
    newContract.SignContract(clientPrivateKey)
    contractReq := &ContractRequest{
        SecBytes    : secretBytes,
        Version     : 1,
        MessageType : 0,
        Request     : &producer.Data{
                Hdr     : producer.DataHeader {
                    Version : 1,
                    Type    : 0,
                },
                Bdy     : newContract,
        },
    }

    serializedReq, err := contractReq.Serialize()
    if err != nil {
        return err
    }

    _, err = SendToProducer(serializedReq, producerAddr)
    if err != nil {
        return err
    }

    return nil
}
