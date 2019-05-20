package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"

	"github.com/pborman/getopt"
)

type Flags struct {
	help       *bool
	debug      *bool
	version    *bool
	height     *bool
	genesis    *bool
	logs       *string
	port       *string
	interval   *string
	initSupply *uint64
}

var version = uint16(1)
var ledger = "blockchain.dat"
var metadata = "metadata.tab"
var accounts = "accounts.tab"

func main() {
	fl := Flags{
		help:       getopt.BoolLong("help", '?', "help"),
		debug:      getopt.BoolLong("debug", 'd', "debug"),
		version:    getopt.BoolLong("version", 'v', "version"),
		height:     getopt.BoolLong("height", 'h', "height"),
		genesis:    getopt.BoolLong("genesis", 'g', "genesis"),
		logs:       getopt.StringLong("log", 'l', "logs.txt", "log file"),
		port:       getopt.StringLong("port", 'p', "13131", "port"),
		interval:   getopt.StringLong("interval", 'i', "", "production interval"),
		initSupply: getopt.Uint64Long("supply", 'y', 0, "initial supply"),
	}
	getopt.Lookup('l').SetOptional()
	getopt.Parse()

	if *fl.help {
		getopt.Usage()
		os.Exit(0)
	}

	if *fl.version {
		fmt.Printf("Aurum producer version: %d\n", version)
	}

	var lgr = log.New(ioutil.Discard, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	if *fl.debug {
		lgr.SetOutput(os.Stderr)
	}

	if *fl.genesis {
		genesisHashes, err := producer.ReadGenesisHashes()
		if err != nil {
			lgr.Fatalf("failed to read in genesis hashes because %s", err.Error())
		}
		if !getopt.IsSet('y') {
			lgr.Fatalln("must set initial aurum supply")
		}
		if *fl.initSupply < uint64(len(genesisHashes)) {
			lgr.Fatalln("must allocate at least 1 aurum per genesis hash")
		}
		genesisBlock, err := producer.BringOnTheGenesis(genesisHashes, *fl.initSupply)
		if err != nil {
			lgr.Fatalf("failed to create genesis block because: %s", err.Error())
		}
		err = producer.Airdrop(ledger, metadata, genesisBlock)
		if err != nil {
			lgr.Fatalf("failed to execute airdrop because: %s", err.Error())
		} else {
			lgr.Println("airdrop successful.")
			os.Exit(0)
		}
	}

	_, err := os.Stat(ledger)
	if err != nil {
		lgr.Fatalf("failed to load ledger %s\n", err.Error())
	}

	// Main loop
	productionInterval, err := time.ParseDuration(*fl.interval)
	if err != nil {
		lgr.Fatalln("failed to parse block production interval")
	} else {
		lgr.Println("block production interval: " + *fl.interval)
	}
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledger, metadata)
	if err != nil {
		lgr.Fatalf("failed to retrieve youngest block: %s\n", err)
	}
	var intervalChannel = make(chan bool)
	var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
	timeSince := time.Since(lastTimestamp)
	if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
		go func() { intervalChannel <- true }()
	}
	var chainHeight = youngestBlockHeader.Height
	var dataPool []producer.Data

	select {
	case <-intervalChannel:
		lgr.Printf("block ready for production: #%d\n", chainHeight+1)
		if newBlock, err := producer.CreateBlock(version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), dataPool); err != nil {
			lgr.Fatalf("failed to add block %s", err.Error())
		} else {
			// TODO: make account.Validate only validate the transaction
			// TODO: table should be updated in separate call, after AddBlock
			// TODO: use a sync.Mutex.Lock()/Unlock() for editing tables
			if err := blockchain.AddBlock(newBlock, ledger, metadata); err != nil {
				lgr.Fatalf("failed to add block: %s", err.Error())
			} else {
				chainHeight++
				lgr.Printf("block produced: #%d\n", chainHeight)
				youngestBlockHeader = newBlock.GetHeader()
				go func() {
					time.AfterFunc(productionInterval, func() {
						intervalChannel <- true
					})
				}()
			}
		}
	}
}

// func init() {
// 	if *fl.debug {
// 		lgr.SetOutput(os.Stderr)
// 	}
// 	if getopt.IsSet('l') || getopt.IsSet("log") {
// 		os.Mkdir(filepath, 0777)
// 		if *fl.logs == "" {
// 			filepath += "logs.txt"
// 		} else {
// 			filepath += *fl.logs
// 		}
// 		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatalln("failed to open file" + err.Error())
// 		}
// 		defer func() {
// 			if err := f.Close(); err != nil {
// 				log.Fatalln("failed to close file")
// 			}
// 		}()
// 		lgr.SetOutput(io.Writer(f))
// 	}
// }

// func init() {
// 	if *fl.globalhost {
// 		addr = ":"
// 		lgr.Println("Listening on all IP addresses")
// 	} else {
// 		addr = fmt.Sprintf("localhost:")
// 		lgr.Println("Listening on local IP addresses")
// 	}
// }

// func main() {
// 	ln, err := net.Listen("tcp", addr+*fl.port)
// 	if err != nil {
// 		lgr.Fatalln("Failed to start server.")
// 	}
// 	lgr.Printf("Server listening on port %s.", *fl.port)
// 	newDataChan := make(chan producer.Data)
// 	go func() {
// 		for {
// 			conn, err := ln.Accept()
// 			if err != nil {
// 				continue
// 			}
// 			lgr.Printf("%s connection\n", conn.RemoteAddr())
// 			go func() {
// 				defer conn.Close()
// 				buf := make([]byte, 1024)
// 				_, err := conn.Read(buf)
// 				if err != nil {
// 					return
// 				}
// 				// Handle message
// 				conn.Write(buf)
// 			}()
// 		}
// 	}()

// 	// Main loop
// 	timerChan := make(chan bool)
// 	// var chainHeight uint64
// 	var dataPool []producer.Data
// 	// productionInterval, err := time.ParseDuration(*interval)
// 	// if err != nil {
// 	// 	lgr.Fatalln("failed to parse interval")
// 	// }
// 	// youngestBlock, err := blockchain.GetYoungestBlock(ledger, metadata)
// 	// if err != nil {
// 	// 	lgr.Fatalf("failed to retrieve youngest block header: %s\n", err)
// 	// }
// 	for {
// 		select {
// 		case newData := <-newDataChan:
// 			dataPool = append(dataPool, newData)
// 		case <-timerChan:
// 			// newBlock, _ := producer.CreateBlock(version, chainHeight+1, block.HashBlock(youngestBlock), dataPool)
// 			// blockchain.AddBlock(newBlock, ledger, metadata)
// 			dataPool = nil
// 			// go func() {
// 			// 	time.AfterFunc(productionInterval, func() {
// 			// 		<-timerChan
// 			// 	})
// 			// }()
// 		}
// 	}

// 	// Close the server
// 	ln.Close()
// }

// var filepath = os.Getenv("GOPATH") + "/src/github.com/SIGBlockchain/project_aurum/producer/logs/"
// 	var lgr = log.New(ioutil.Discard, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

// 	if *fl.debug {
// 		lgr.SetOutput(os.Stderr)
// 	}
// 	if getopt.IsSet('l') || getopt.IsSet("log") {
// 		os.Mkdir(filepath, 0777)
// 		if *fl.logs == "" {
// 			filepath += "logs.txt"
// 		} else {
// 			filepath += *fl.logs
// 		}
// 		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatalln("failed to open file" + err.Error())
// 		}
// 		defer func() {
// 			if err := f.Close(); err != nil {
// 				log.Fatalln("failed to close file")
// 			}
// 		}()
// 		lgr.SetOutput(io.Writer(f))
// 	}
