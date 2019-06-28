package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/pkg/keys"

	"github.com/SIGBlockchain/project_aurum/internal/config"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"

	"github.com/SIGBlockchain/project_aurum/internal/endpoints"
	"github.com/SIGBlockchain/project_aurum/internal/handlers"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"

	"github.com/SIGBlockchain/project_aurum/internal/constants"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration : %v", err)
	}

	// If no blockchain.dat, perform airdrop
	if _, err := os.Stat(constants.BlockchainFile); os.IsNotExist(err) {
		log.Println("No blockchain file detected. Executing genesis procedure...")
		addresses, err := producer.ReadGenesisHashes()
		if err != nil {
			log.Fatalf("Failed to read genesis addresses: %v", err)
		}
		genesisBlock, err := producer.BringOnTheGenesis(addresses, cfg.InitialAurumSupply)
		if err != nil {
			log.Fatalf("Failed to create genesis block: %v", err)
		}
		log.Println("Attempting airdrop...")
		if err := producer.Airdrop(constants.BlockchainFile, constants.MetadataTable, constants.AccountsTable, genesisBlock); err != nil {
			log.Fatalf("Failed to perform airdrop: %v", err)
		}
		log.Println("Airdrop complete.")
	}

	accountsDatabaseConnection, err := sql.Open("sqlite3", constants.AccountsTable)
	if err != nil {
		log.Fatalf("Failed to open connection : %v", err)
	}
	defer accountsDatabaseConnection.Close()

	contractChannel := make(chan accounts.Contract)

	var pendingContractPool []accounts.Contract
	intervalChannel := make(chan bool)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(constants.BlockchainFile, constants.MetadataTable)
	if err != nil {
		log.Fatalf("Failed to get youngestBlockHeader")
	}

	var chainHeight = youngestBlockHeader.Height
	var numBlocksGenerated uint64

	var hostname string
	if cfg.Localhost {
		hostname = "localhost:"
	} else {
		hostname = ":"
	}
	hostname += cfg.Port

	http.HandleFunc(endpoints.AccountInfo, handlers.HandleAccountInfoRequest(accountsDatabaseConnection))
	http.HandleFunc(endpoints.Contract, handlers.HandleContractRequest(accountsDatabaseConnection, contractChannel))
	go http.ListenAndServe(hostname, nil)
	log.Printf("Serving requests on port: %s", cfg.Port)

	productionInterval, err := time.ParseDuration(cfg.BlockProductionInterval)
	if err != nil {
		log.Fatalf("Failed to parse production interval: %v", err)
	}
	go triggerInterval(intervalChannel, productionInterval)
	log.Printf("Will produce blocks every %s", cfg.BlockProductionInterval)

	for {
		select {
		case newContract := <-contractChannel:
			pendingContractPool = append(pendingContractPool, newContract)
		case <-intervalChannel:
			log.Printf("Block #%d ready for production.", chainHeight+1)
			if newBlock, err := producer.CreateBlock(cfg.Version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), pendingContractPool); err != nil {
				log.Fatalf("Failed to create block %v", err)
			} else {
				if err := blockchain.AddBlock(newBlock, constants.BlockchainFile, constants.MetadataTable); err != nil {
					log.Fatalf("Failed to add block %v", err)
				} else {
					chainHeight++
					for _, contract := range pendingContractPool {
						senderPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(contract.SenderPubKey))
						if err := accounts.ExchangeBetweenAccountsUpdateAccountBalanceTable(accountsDatabaseConnection, senderPublicKeyHash, contract.RecipPubKeyHash, contract.Value); err != nil {
							log.Printf("Failed to add contact to accounts database : %v", err)
						}
					}
					log.Printf("Block #%d successfully added to blockchain", chainHeight)
					log.Printf("%d contracts confirmed", len(pendingContractPool))
					pendingContractPool = nil
					numBlocksGenerated++
					go triggerInterval(intervalChannel, productionInterval)
				}
			}
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
