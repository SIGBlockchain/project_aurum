package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/config"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/pendingpool"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"

	"github.com/SIGBlockchain/project_aurum/internal/endpoints"
	"github.com/SIGBlockchain/project_aurum/internal/handlers"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/genesis"
)

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

	// Load configuration
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration : %v", err)
	}

	// If no blockchain.dat, perform airdrop
	if _, err := os.Stat(constants.BlockchainFile); os.IsNotExist(err) {
		log.Println("No blockchain file detected. Executing genesis procedure...")
		addresses, err := genesis.ReadGenesisHashes()
		if err != nil {
			log.Fatalf("Failed to read genesis addresses: %v", err)
		}
		genesisBlock, err := genesis.BringOnTheGenesis(addresses, cfg.InitialAurumSupply)
		if err != nil {
			log.Fatalf("Failed to create genesis block: %v", err)
		}
		log.Println("Attempting airdrop...")
		if err := blockchain.Airdrop(constants.DockerVolumeDir+constants.BlockchainFile, constants.DockerVolumeDir+constants.MetadataTable, constants.DockerVolumeDir+constants.AccountsTable, genesisBlock); err != nil {
			log.Fatalf("Failed to perform airdrop: %v", err)
		}
		log.Println("Airdrop complete.")
	}

	// TODO: If we did have a blockchain.dat but no table(s), we could execute a recovery here

	// Open connection to accounts database
	accountsDatabaseConnection, err := sql.Open("sqlite3", constants.DockerVolumeDir+constants.AccountsTable)
	if err != nil {
		log.Fatalf("Failed to open connection : %v", err)
	}
	defer accountsDatabaseConnection.Close()

	// Open connection to metadata database
	metadataDatabaseConnection, err := sql.Open("sqlite3", constants.DockerVolumeDir+constants.MetadataTable)
	if err != nil {
		log.Fatalf("Failed to open connection : %v", err)
	}
	defer metadataDatabaseConnection.Close()

	// Declare channel for new contracts
	contractChannel := make(chan contracts.Contract)

	// Declare pool of contracts pending block production
	var pendingContractPool []contracts.Contract

	// Signal channel for interrupts
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// Extract youngest block header from blockchain
	ledgerFile, err := os.OpenFile(constants.DockerVolumeDir+constants.BlockchainFile, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open ledger file")
	}
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledgerFile, metadataDatabaseConnection)
	if err != nil {
		log.Fatalf("Failed to get youngestBlockHeader")
	}
	if err := ledgerFile.Close(); err != nil {
		log.Fatalf("Failed to close ledger file: %v", err)
	}
	// Metadata about block production
	var chainHeight = youngestBlockHeader.Height
	var numBlocksGenerated uint64

	// Define hostname address
	var hostname string
	if cfg.Localhost {
		hostname = "localhost:"
	} else {
		hostname = ":"
	}
	hostname += cfg.Port

	// Lock for pending map
	pendingLock := new(sync.Mutex)

	pendingMap := pendingpool.NewPendingMap()

	// Set handlers for endpoints and run server
	http.HandleFunc(endpoints.AccountInfo, handlers.HandleAccountInfoRequest(accountsDatabaseConnection, pendingMap, pendingLock))

	http.HandleFunc(endpoints.Contract, handlers.HandleContractRequest(accountsDatabaseConnection, contractChannel, pendingMap, pendingLock))
	go http.ListenAndServe(hostname, nil)
	log.Printf("Serving requests on port %s", cfg.Port)

	// Declare channel for triggering block production
	intervalChannel := make(chan bool)
	productionInterval, err := time.ParseDuration(cfg.BlockProductionInterval)
	if err != nil {
		log.Fatalf("Failed to parse production interval: %v", err)
	}

	// Trigger block production after interval has elapsed
	go triggerInterval(intervalChannel, productionInterval)
	log.Printf("Will produce block every %s", cfg.BlockProductionInterval)
	log.Printf("Current chain height is %d", chainHeight)

	for {
		select {
		// New valid contract received is added to pending pool
		case newContract := <-contractChannel:
			newContractEncodedSenderPubKey, err := publickey.Encode(newContract.SenderPubKey)
			if err != nil {
				log.Fatalf("Failed to encode new contract sender public key")
			}
			pendingContractPool = append(pendingContractPool, newContract)
			log.Printf("Added new contract to pool:\n(%s) ->|%d aurum|-> (%s) ",
				hex.EncodeToString(hashing.New(newContractEncodedSenderPubKey)),
				newContract.Value, hex.EncodeToString(newContract.RecipPubKeyHash))

		// New block is ready to be produced
		case <-intervalChannel:
			log.Printf("Block #%d ready for production.", chainHeight+1)
			pendingLock.Lock()
			if newBlock, err := block.New(cfg.Version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), pendingContractPool); err != nil {
				log.Fatalf("Failed to create block %v", err)
			} else {

				// Add block to blockchain
				blockchainFile, err := os.OpenFile(constants.BlockchainFile, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatalf("failed to open ledger file: %s\n", err)
				}
				err = blockchain.AddBlock(newBlock, blockchainFile, metadataDatabaseConnection)
				if err != nil {
					log.Fatalf("Failed to add block %v", err)
				} else {
					if err := blockchainFile.Close(); err != nil {
						log.Fatalf("Failed to close blockchain file: %v", err)
					}
					chainHeight++
					// Reset pending pool map to empty
					for k := range pendingMap.Sender {
						delete(pendingMap.Sender, k)
					}
					// Update accounts table with all contracts in pool
					for _, contract := range pendingContractPool {
						if err = accountstable.ExchangeAndUpdateAccounts(accountsDatabaseConnection, &contract); err != nil {
							log.Printf("Failed to add contract %+v to accounts database : %v", contract, err)
						}
					}

					log.Printf("Block #%d successfully added to blockchain", chainHeight)
					log.Printf("%d contracts confirmed in block #%d", len(pendingContractPool), chainHeight)

					// Reset pool
					pendingContractPool = nil

					// Reset youngest block header
					youngestBlockHeader = newBlock.GetHeader()

					// Reset production interval
					go triggerInterval(intervalChannel, productionInterval)

					numBlocksGenerated++

				}
			}
			pendingLock.Unlock()
		// Signal interrupt detected
		case <-signalChannel:
			fmt.Print("\r")
			log.Println("Interrupt signal encountered, terminating...")
			log.Printf("Number of blocks generated: %d", numBlocksGenerated)
			return
		}

	}
}

func triggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	// Triggers block production case
	time.Sleep(productionInterval)
	intervalChannel <- true
}
