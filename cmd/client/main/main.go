package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	producer "github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/pborman/getopt"
)

type Flags struct {
	help       *bool
	version    *bool
	setup      *bool
	info       *bool
	updateInfo *bool
	contract   *bool
	recipient  *string
	value      *uint64
	producer   *string
}

var version = uint16(1)
var wallet = "aurum_wallet.json"

func PrintInfo() error {
	walletAddr, err := client.GetWalletAddress()
	if err != nil {
		return err
	}
	stateNonce, err := client.GetStateNonce()
	if err != nil {
		return err
	}
	balance, err := client.GetBalance()
	if err != nil {
		return err
	}
	fmt.Printf("Wallet Address: %s\n", hex.EncodeToString(walletAddr))
	fmt.Printf("State nonce: %d\n", stateNonce)
	fmt.Printf("Balance: %d\n", balance)
	return nil
}

func main() {
	fl := Flags{
		help:       getopt.BoolLong("help", '?', "help"),
		version:    getopt.BoolLong("version", 'w', "version"),
		setup:      getopt.BoolLong("setup", 's', "set up client"),
		info:       getopt.BoolLong("info", 'i', "wallet info"),
		updateInfo: getopt.BoolLong("update", 'u', "update wallet info"),
		contract:   getopt.BoolLong("contract", 'c', "make contract"),
		recipient:  getopt.StringLong("recipient", 'r', "recipient"),
		value:      getopt.Uint64Long("value", 'v', 0, "value to send"),
		producer:   getopt.StringLong("producer", 'p', "", "producer address"),
	}
	getopt.Parse()

	if *fl.help {
		getopt.Usage()
		os.Exit(0)
	}

	if *fl.version {
		fmt.Printf("Aurum client version: %d\n", version)
		os.Exit(0)
	}

	var lgr = log.New(os.Stderr, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	if *fl.setup {
		fmt.Println("Initializing Aurum wallet...")
		if err := client.SetupWallet(); err != nil {
			lgr.Fatalf("failed to setup wallet: %s", err.Error())
		} else {
			fmt.Println("Wallet setup completed successfully.")
			if err := PrintInfo(); err != nil {
				lgr.Fatalf("failed to print wallet info: %s", err.Error())
			}
			os.Exit(0)
		}
	}

	if *fl.info {
		fmt.Println("Wallet setup completed successfully.")
		if err := PrintInfo(); err != nil {
			lgr.Fatalf("failed to print wallet info: %s", err.Error())
		}
	}

	if *fl.contract {
		// TODO: Check for *fl.recipient, *fl.value ( x > 0 ), and *fl.producer address; if any are missing, lgr.Fatal()
		// TODO: Call ContractMessageFromInput(...) and Send to producer
		// TODO: Output success of sending to producer (with response)
	}

	if *fl.updateInfo {
		// TODO: Send a update info message, receive the response and update the wallet balance/nonce
	}

}

// Convert value to uint64; if unsuccessful output an error
// If value is zero, output error
// GetBalance(), if value is > than wallet balance, output an error
// GetStateNonce(), GetPrivateKey()
// Convert recipient to []byte; if unsuccessful output an error
// MakeContract(...) (use version global), SignContract(...)
// Output a contract message, with the following structure:
// producer.SecretBytes + uint8(1) + serializedContract
// NOTE: The uint8(1) here will let the producer know that this is a contract message
func ContractMessageFromInput(value string, recipient string) ([]byte, error) {
	intVal, err := strconv.Atoi(value) // convert value (string) to int
	if err != nil {
		return nil, errors.New("Unable to convert input to int " + err.Error())
	}

	// case input is zero or less
	if intVal <= 0 {
		return nil, errors.New("Input value is less than or equal to zero")
	}

	// case balance < input
	balance, err := client.GetBalance()
	if err != nil {
		return nil, errors.New("Failed to get balance: " + err.Error())
	}
	if balance < uint64(intVal) {
		return nil, errors.New("Input is greater than available balance")
	}

	stateNonce, err := client.GetStateNonce()
	if err != nil {
		return nil, errors.New("Failed to get stateNonce: " + err.Error())
	}

	// case recipBytes != 32
	recipBytes := []byte(recipient)
	if len(recipBytes) != 32 {
		return nil, errors.New("Failed to convert recipient to size 32 byte slice")
	}

	senderPubKey, _ := client.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	contract, err := accounts.MakeContract(version, senderPubKey, recipBytes, uint64(intVal), stateNonce)
	if err != nil {
		return nil, err
	}

	contract.SignContract(senderPubKey)

	totalSize := (2 + 178 + 1 + int(contract.SigLen) + 32 + 8 + 8)
	contractMessage := make([]byte, 10+totalSize)
	copy(contractMessage[0:8], producer.SecretBytes)
	binary.LittleEndian.PutUint16(contractMessage[8:10], version)
	serializedContract, err := contract.Serialize()
	copy(contractMessage[10:], serializedContract)
	return contractMessage, err
}

// ---------------------------------------------------------------------------
// func main() {
// 	// List of Options
// 	helpFlag := getopt.Bool('?', "Display Valid Flags")
// 	debugFlag := getopt.BoolLong("debug", 'd', "Enable Debug Mode")
// 	logFile := getopt.StringLong("logfile", 'l', "", "Log File Destination")
// 	getopt.CommandLine.Lookup('l').SetOptional()
// 	getopt.Parse()
// 	// If the help flag is on, print usage to os.Stdout
// 	if *helpFlag == true {
// 		getopt.Usage()
// 		os.Exit(0)
// 	}
// 	logger := log.New(os.Stdout, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

// 	// If the debug flag is not on, the logger is set to a dummy buffer, which stores the input
// 	if *debugFlag == false {
// 		var buffer bytes.Buffer
// 		logger = log.New(&buffer, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
// 	}
// 	// If the log flag is on, it will send the logs to a file in client/logs
// 	if getopt.CommandLine.Lookup('l').Count() > 0 {
// 		filepath := os.Getenv("GOPATH") + "/src/github.com/SIGBlockchain/project_aurum/logs"
// 		os.Mkdir(filepath, 0777)
// 		// If no filename is given, logs.txt
// 		if *logFile == "" {
// 			filepath += "/client_logs.txt"
// 			// Otherwise the custom filename is used
// 		} else {
// 			filepath += "/" + *logFile
// 		}
// 		f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
// 		defer f.Close()

// 		// If there is any error, do not set the logger. Log an error messgae
// 		if err != nil {
// 			logger.Fatalln(err)
// 		} else {
// 			logger = log.New(f, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
// 		}
// 	}
// 	// Clears the screen before program execution
// 	client.ClearScreen()

// 	// Check to see if there is an internet connection
// 	err := client.CheckConnection()
// 	//err := error(nil) // This is used for offline testing
// 	if err != nil {
// 		logger.Fatalln(err)
// 	}
// 	logger.Println("Connection check passed.")

// 	// This string contains the entire input
// 	var userInput string
// 	for {
// 		// If there are any errors in getting input, end program execution
// 		if client.GetUserInput(&userInput, os.Stdin) != nil {
// 			logger.Fatalln("Error getting input.")
// 		}
// 		// inputs holds the individual arguments of a command
// 		inputs := strings.Split(userInput, " ")
// 		// Switch checks the first argument of a command
// 		switch inputs[0] {

// 		// 'q' command exits the program
// 		case "q":
// 			logger.Println("Exiting program.\nGoodbye")
// 			os.Exit(0)
// 		// 'clear' command clears the command window of all previous text, adds upper divider
// 		case "clear":
// 			client.ClearScreen()
// 		// 'help' command prints all the avalible commands, and a brief description
// 		case "help":
// 			client.PrintHelp()
// 		// 'moreinfo' command prints the link to the github page
// 		case "moreinfo":
// 			client.PrintGithubLink()
// 		// 'send' command sends a given value to a given recipient
// 		//		send [recipient] [value]
// 		case "send":
// 			// 'send' requires 3 arguments at a minimum, otherwise ignore command
// 			if len(inputs) != 3 {
// 				fmt.Println("ERROR: Improper Use of send Command\n\tsend [recipient] [value]")
// 				break
// 			}
// 			// This 64 bit integer holds the recipient's public key
// 			var recipient string
// 			_, err := fmt.Sscanf(inputs[1], "%s", &recipient)
// 			if err != nil {
// 				logger.Println("ERROR: Attempt to collect recipient string failed")
// 				break
// 			}
// 			// This 64 bit integer holds the value for the contract
// 			var value int64
// 			_, err = fmt.Sscanf(inputs[2], "%d", &value)
// 			if err != nil {
// 				logger.Println("ERROR: Attempt to collect value integer failed")
// 				break
// 			}
// 			logger.Println("Accepted send as valid input")
// 			// Pass recipient, value to function
// 		case "balance":
// 			// Insert function to get balance
// 			logger.Println("Accepted send as valid input")
// 		// If first argument of a command matches no valid command, print an error message
// 		default:
// 			fmt.Println("Invalid command \"" + userInput + "\" rejected\n\tUse \"help\" to see available commands")
// 			logger.Println("Invalid command \"" + userInput + "\" rejected")
// 		}
// 	}
// }
