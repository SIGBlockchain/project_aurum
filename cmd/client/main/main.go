package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
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
	value      *string
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
		value:      getopt.StringLong("value", 'v', "", "value to send"),
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
		newContract, err := ContractMessageFromInput(*fl.value, *fl.recipient)
		if err != nil {
			lgr.Fatalf("Failed to create contract message: " + err.Error())
		}
		conn, err := net.Dial("tcp", *fl.producer)
		if err != nil {
			lgr.Fatalf("Failed to connect to producer address: " + *fl.producer + ": " + err.Error())
		}
		if _, err := conn.Write(newContract); err != nil {
			lgr.Fatalf("Failed to send contract to producer: " + err.Error())
		}
		buf := make([]byte, 1024)
		nRcvd, err := conn.Read(buf)
		if err != nil {
			lgr.Fatalf("Failed to read from connection: " + err.Error())
		}
		lgr.Printf("Producer returned with message: " + string(buf[:nRcvd]))
		currentBalance, err := client.GetBalance()
		currentNonce, err := client.GetStateNonce()
		intVal, err := strconv.Atoi(*fl.value)
		if err != nil {
			lgr.Fatalf("Failed to convert value to integer: " + err.Error())
		}
		err = client.UpdateWallet((currentBalance - uint64(intVal)), (currentNonce + 1))
		if err != nil {
			lgr.Fatalf("Failed to update wallet: " + err.Error())
		}
	}

	if *fl.updateInfo {
		if *fl.producer == "" {
			lgr.Fatalf("Producer address is required to update wallet info")
		}

		lgr.Println("Updating wallet info...")
		accountInfo, err := client.RequestWalletInfo(*fl.producer)
		if err != nil {
			lgr.Fatalf("failed to request wallet info: %s", err.Error())
		}

		if err := client.UpdateWallet(accountInfo.Balance, accountInfo.StateNonce); err != nil {
			lgr.Fatalf("failed to update wallet info: %s", err.Error())
		}

		if err := PrintInfo(); err != nil {
			lgr.Fatalf("failed to print wallet info: %s", err.Error())
		}
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
	recipBytes, err := hex.DecodeString(recipient)
	if err != nil {
		return nil, errors.New("Failed to hex decode recipient")
	}
	if len(recipBytes) != 32 {
		return nil, errors.New("Failed to convert recipient to size 32 byte slice")
	}

	senderPubKey, _ := client.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	contract, err := accounts.MakeContract(version, senderPubKey, recipBytes, uint64(intVal), stateNonce+1)
	if err != nil {
		return nil, err
	}
	contract.SignContract(senderPubKey)
	serializedContract, _ := contract.Serialize()

	var contractMessage []byte
	contractMessage = append(contractMessage, producer.SecretBytes...)
	contractMessage = append(contractMessage, 1)
	contractMessage = append(contractMessage, serializedContract...)
	return contractMessage, nil
}
