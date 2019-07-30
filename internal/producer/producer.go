// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/validation"
)

type Flags struct {
	Help        *bool
	Debug       *bool
	Version     *bool
	Height      *bool
	Genesis     *bool
	Test        *bool
	Globalhost  *bool
	MemoryStats *bool
	Logs        *string
	Port        *string
	Interval    *string
	InitSupply  *uint64
	NumBlocks   *uint64
}

var version = uint16(1)
var ledger = "blockchain.dat"
var metadataTable = constants.MetadataTable

// TODO: Will need to change the name to support get functions
var accountsTable = constants.AccountsTable

var SecretBytes = hashing.New([]byte("aurum"))[8:16]

// This stores connection information for the producer
type BlockProducer struct {
	Server        net.Listener
	NewConnection chan net.Conn
	Logger        *log.Logger
	// Add ledger name, metadata name, and contract table name
	// Slice of Contracts representing contract pool
}

// Should contain version, payload type, payload size
type Header struct{}

// Messages have headers and payloads
// Payloads should correspond to message type
type Message struct{}

func RunServer(ln net.Listener, bChan chan []byte, debug bool) {
	// Set logger
	var lgr = log.New(ioutil.Discard, "SRVR_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if debug {
		lgr.SetOutput(os.Stdout)
	}
	for {
		lgr.Println("Waiting for connection...")

		// Block for connection
		conn, err := ln.Accept()
		var nRcvd int
		buf := make([]byte, 1024)
		if err != nil {
			lgr.Println("connection failed")
			goto End
		}
		lgr.Printf("%s connected\n", conn.RemoteAddr())
		defer conn.Close()

		// Block for receiving message
		nRcvd, err = conn.Read(buf)
		if err != nil {
			goto End
		}

		lgr.Println("Received message from", conn.RemoteAddr())

		// Determine the type of message
		if nRcvd < 8 || (!bytes.Equal(buf[:8], SecretBytes)) {
			conn.Write([]byte("No thanks.\n"))
			goto End
		} else {
			conn.Write([]byte("Thank you.\n"))
			if buf[8] == 2 {
				lgr.Println("Received account info request")
				// TODO: Will require a sync.Mutex lock here eventually
				// Open connection to account table
				dbConnection, err := sql.Open("sqlite3", constants.AccountsTable)
				if err != nil {
					lgr.Fatalf("Failed to open account table: %s\n", err)
				}
				accInfo, err := accountstable.GetAccountInfo(dbConnection, buf[9:nRcvd])
				if err := dbConnection.Close(); err != nil {
					lgr.Fatalf("Failed to close account table: %s\n", err)
				}
				var responseMessage []byte
				responseMessage = append(responseMessage, SecretBytes...)
				if err != nil {
					lgr.Printf("Failed to get account info for %s: %s", hex.EncodeToString(buf[9:nRcvd]), err.Error())
					responseMessage = append(responseMessage, 1)
				} else {
					responseMessage = append(responseMessage, 0)
					if serializedAccInfo, err := accInfo.Serialize(); err == nil {
						responseMessage = append(responseMessage, serializedAccInfo...)
					}
				}

				time.Sleep(3 * time.Second)
				conn.Write(responseMessage)
				goto End
			}
		}

		lgr.Println("Sending to channel")
		// Send to channel if aurum-related message
		bChan <- buf[:nRcvd]

		lgr.Println("Message successfully sent to main")
		goto End
	End:
		lgr.Println("Closing connection.")
		conn.Close()
	}
}

func ProduceBlocks(byteChan chan []byte, fl Flags, limit bool) {
	// Set logger
	var lgr = log.New(ioutil.Discard, "PROD_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if *fl.Debug {
		lgr.SetOutput(os.Stdout)
	}

	// Open connection to metadata database
	metadataConn, err := sql.Open("sqlite3", metadataTable)
	if err != nil {
		lgr.Fatalf("failed to open metadata table: %s\n", err)
	}
	defer metadataConn.Close()

	// Open connection to account database
	dbConnection, err := sql.Open("sqlite3", constants.AccountsTable)
	if err != nil {
		lgr.Fatalf("Failed to open account table: %s\n", err)
	}
	defer func() {
		if err := dbConnection.Close(); err != nil {
			lgr.Fatalf("Failed to close account table: %v", err)
		}
	}()

	// Retrieve youngest block header
	ledgerFile, err := os.OpenFile(ledger, os.O_RDONLY, 0644)
	if err != nil {
		lgr.Fatalf("failed to open ledger file: %s\n", err)
	}
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledgerFile, metadataConn)
	if err != nil {
		lgr.Fatalf("failed to retrieve youngest block: %s\n", err)
	}
	if err := ledgerFile.Close(); err != nil {
		lgr.Fatalf("Failed to close blockchain file: %v", err)
	}

	// Set up SIGINT channel
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Initialize other variables
	var numBlocksGenerated uint64
	var dataPool []contracts.Contract
	var ms runtime.MemStats

	// Determine production interval and start trigger goroutine
	productionInterval, err := time.ParseDuration(*fl.Interval)
	if err != nil {
		lgr.Fatalln("failed to parse block production interval")
	} else {
		lgr.Println("block production interval: " + *fl.Interval)
	}
	var intervalChannel = make(chan bool)
	go triggerInterval(intervalChannel, productionInterval)

	// Main loop
	for {
		var chainHeight = youngestBlockHeader.Height
		select {
		case message := <-byteChan:

			// If it's a contract, add it to the contract pool
			switch message[8] {
			case 1:
				lgr.Println("Received contract")
				var newContract contracts.Contract
				if err := newContract.Deserialize(message[9:]); err == nil {
					// TODO: Validate the contract prior to adding
					if err := validation.ValidateContract(dbConnection, &newContract); err != nil {
						lgr.Println("Invalid contract because: " + err.Error())
					} else {
						dataPool = append(dataPool, newContract)
						lgr.Println("Valid contract")
					}
				}
				break
			}
		case <-intervalChannel:
			// Triggered if it's time to produce a block
			lgr.Printf("block ready for production: #%d\n", chainHeight+1)
			// lgr.Printf("Production block dataPool: %v", dataPool)
			if newBlock, err := block.New(version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), dataPool); err != nil {
				lgr.Fatalf("failed to add block %s", err.Error())
				os.Exit(1)
			} else {

				// Add the block
				ledgerFile, err := os.OpenFile(constants.BlockchainFile, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					lgr.Fatalf("failed to open ledger file: %s\n", err)
				}
				err = blockchain.AddBlock(newBlock, ledgerFile, metadataConn)
				if err != nil {
					lgr.Fatalf("failed to add block: %s", err.Error())
					os.Exit(1)
				} else {
					if err := ledgerFile.Close(); err != nil {
						log.Fatalf("Failed to close blockchain file: %v", err)
					}
					lgr.Printf("block produced: #%d\n", chainHeight+1)
					numBlocksGenerated++
					youngestBlockHeader = newBlock.GetHeader()
					go triggerInterval(intervalChannel, productionInterval)

					// TODO: for each contract in the dataPool, update the accounts table
					// TODO: will require a sync.Mutex for the accounts table
					dbConn, err := sql.Open("sqlite3", constants.AccountsTable)
					if err != nil {
						lgr.Fatalf("Failed to connect to accounts database: %v", err)
					}
					for _, contract := range dataPool {
						senderPKH := hashing.New(publickey.Encode(contract.SenderPubKey))
						err := accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(dbConn, senderPKH, contract.RecipPubKeyHash, contract.Value)
						if err != nil {
							lgr.Printf("Failed to add contract to accounts database: %v", err)
						}
					}
					dbConn.Close()
					dataPool = nil

					// Memory stats
					if *fl.MemoryStats {
						runtime.ReadMemStats(&ms)
						printMemstats(ms)
					}
					// If in test mode, break the loop
					if *fl.Test {
						lgr.Printf("Test mode: breaking loop")
						return
					}

					// If reached limit of blocks desired to be generated, break the loop
					if limit && (numBlocksGenerated >= *fl.NumBlocks) {
						lgr.Printf("Limit reached: # blocks generated: %d, blocks desired: %d\n", numBlocksGenerated, *fl.NumBlocks)
						return
					}
				}

			}

		case <-signalCh:

			// If you receive a SIGINT, exit the loop
			fmt.Print("\r")
			lgr.Println("Interrupt signal encountered, program terminating.")
			return
		}

	}
}

func printMemstats(ms runtime.MemStats) {
	// useful commands: go run -gcflags='-m -m' main.go <main flags>
	fmt.Printf("Bytes of allocated heap objects: %d", ms.Alloc)
	fmt.Printf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
	fmt.Printf("Count of heap objects allocated: %d", ms.Mallocs)
	fmt.Printf("Count of heap objects freed: %d", ms.Frees)
}

func triggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	// Triggers block production case
	time.Sleep(productionInterval)
	intervalChannel <- true
}

func calculateInterval(youngestBlockHeader block.BlockHeader, productionInterval time.Duration, intervalChannel chan bool) {
	var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
	timeSince := time.Since(lastTimestamp)
	if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
		go triggerInterval(intervalChannel, time.Duration(0))
	} else {
		diff := productionInterval.Nanoseconds() - timeSince.Nanoseconds()
		go triggerInterval(intervalChannel, time.Duration(diff))
	}
}
