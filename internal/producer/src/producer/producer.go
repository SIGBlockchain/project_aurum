// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/validation"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
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

// This checks if the producer is connected to the internet
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("connectivity check failed")
	}
	conn.Close()
	return nil
}

// This will accept any incoming connections
func (bp *BlockProducer) AcceptConnections() {
	for {
		conn, err := bp.Server.Accept()
		if err != nil {
			return
		}
		bp.NewConnection <- conn
		bp.Logger.Printf("%s connected\n", conn.RemoteAddr())
	}
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
				accInfo, err := accountstable.GetAccountInfo(buf[9:nRcvd])
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

	// Retrieve youngest block header
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledger, metadataTable)
	if err != nil {
		lgr.Fatalf("failed to retrieve youngest block: %s\n", err)
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
					if err := validation.ValidateContract(&newContract); err != nil {
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
				if err := blockchain.AddBlock(newBlock, ledger, metadataTable); err != nil {
					lgr.Fatalf("failed to add block: %s", err.Error())
					os.Exit(1)
				} else {
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

// // This is a security feature for the ledger. If the metadata table gets lost somehow, this function will restore it completely.
// //
// // Another situation is when a producer in a decentralized system joins the network and wants the full ledger.
// func RecoverBlockchainMetadata(ledgerFilename string, metadataFilename string, accountBalanceTable string) error {
// 	//check if metadata file exists
// 	err := emptyFile(metadataFilename)
// 	if err != nil {
// 		log.Printf("Failed to create an empty metadata file: %s", err.Error())
// 		return errors.New("Failed to create an empty metadata file")
// 	}
// 	//check if account balance file exists
// 	err = emptyFile(accountBalanceTable)
// 	if err != nil {
// 		log.Printf("Failed to create an empty account file: %s", err.Error())
// 		return errors.New("Failed to create an empty account file")
// 	}

// 	//set up the two database tables
// 	metaDb, err := sql.Open("sqlite3", metadataFilename)
// 	if err != nil {
// 		return errors.New("Failed to open newly created metadata db")
// 	}
// 	defer metaDb.Close()
// 	_, err = metaDb.Exec("CREATE TABLE metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
// 	if err != nil {
// 		return errors.New("Failed to create metadata table")
// 	}

// 	accDb, err := sql.Open("sqlite3", accountBalanceTable)
// 	if err != nil {
// 		return errors.New("Failed to open newly created accounts db")
// 	}
// 	defer accDb.Close()
// 	_, err = accDb.Exec("CREATE TABLE account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
// 	if err != nil {
// 		return errors.New("Failed to create acount_balances table")
// 	}

// 	// open ledger file
// 	ledgerFile, err := os.OpenFile(ledgerFilename, os.O_RDONLY, 0644)
// 	if err != nil {
// 		return errors.New("Failed to open ledger file")
// 	}
// 	defer ledgerFile.Close()

// 	// loop that adds blocks' metadata into database
// 	bPosition := int64(0)
// 	for {
// 		bOldPos := bPosition
// 		deserializedBlock, bLen, err := extractBlock(ledgerFile, &bPosition)
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			log.Printf("Failed to extract block from ledger: %s", err.Error())
// 			return errors.New("Failed to extract block from ledger")
// 		}

// 		//update the metadata table
// 		err = insertMetadata(metaDb, deserializedBlock, bLen, bOldPos)
// 		if err != nil {
// 			return err
// 		}

// 		//update the account table
// 		err = accountstable.UpdateAccountTable(accDb, deserializedBlock)
// 		if err != nil {
// 			return err
// 		}

// 	}

// 	return err
// }

// //creates an empty file if the file doesn't exist, or clears if the contents of the file if it exists
// func emptyFile(fileName string) error {
// 	_, err := os.Stat(fileName)
// 	if err != nil {
// 		f, err := os.Create(fileName)
// 		if err != nil {
// 			return errors.New("Failed to create " + fileName)
// 		}
// 		f.Close()
// 	} else { //file exits, so clear the file
// 		err = os.Truncate(fileName, 0)
// 		if err != nil {
// 			return errors.New("Failed to truncate " + fileName)
// 		}
// 	}
// 	return nil
// }

// //extract a block from the file, also update file position
// func extractBlock(ledgerFile *os.File, pos *int64) (*block.Block, uint32, error) {
// 	length := make([]byte, 4)

// 	// read 4 bytes for blocks' length
// 	_, err := ledgerFile.Read(length)
// 	if err == io.EOF {
// 		return nil, 0, err
// 	} else if err != nil {
// 		return nil, 0, errors.New("Failed to read ledger file")
// 	}

// 	bLen := binary.LittleEndian.Uint32(length)

// 	// set offset for next read to get to the position of the block
// 	ledgerFile.Seek(*pos+int64(len(length)), 0)
// 	serialized := make([]byte, bLen)

// 	//update file position
// 	*pos += int64(len(length) + len(serialized))

// 	//extract block
// 	_, err = io.ReadAtLeast(ledgerFile, serialized, int(bLen))
// 	if err != nil {
// 		return nil, 0, errors.New("Failed to retrieve serialized block")
// 	}

// 	deserializedBlock := block.Deserialize(serialized)
// 	return &deserializedBlock, bLen, nil
// }

// /*inserts the block metadata into the metadata table
//   NOTE: the db connection passed in should be open
// */
// func insertMetadata(db *sql.DB, b *block.Block, bLen uint32, pos int64) error {
// 	bHeight := b.Height
// 	bHash := block.HashBlock(*b)

// 	sqlQuery := "INSERT INTO metadata (height, position, size, hash) VALUES ($1, $2, $3, $4)"
// 	_, err := db.Exec(sqlQuery, bHeight, pos, bLen, bHash)
// 	if err != nil {
// 		log.Printf("Failed to execute statement: %s", err.Error())
// 		return errors.New("Failed to execute statement")
// 	}

// 	return nil
// }
