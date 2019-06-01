package main

/*
Usage:
Must have genesis-hashes file, or will fail.
Run go run main.go with -g --supply=[x] for genesis
Running go run main.go with --interval=[x]ms will generate
a block every x milliseconds (must be milliseconds) indefinitely
Running the same command with -t flag will generate only
one block and exit.
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
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
	test       *bool
	logs       *string
	port       *string
	interval   *string
	initSupply *uint64
	numBlocks  *uint64
}

var version = uint16(1)
var ledger = "blockchain.dat"
var metadataTable = "metadata.tab"
var accountsTable = "accounts.tab"

func RunServer(ln net.Listener, debug bool) {
	var lgr = log.New(ioutil.Discard, "SRVR_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if debug {
		lgr.SetOutput(os.Stdout)
	}
	for {
		if conn, err := ln.Accept(); err == nil {
			lgr.Printf("%s connected\n", conn.RemoteAddr())
			defer conn.Close()
			buf := make([]byte, 1024)
			if nBytes, err := conn.Read(buf); err == nil {
				lgr.Printf("%s sent: %s", conn.RemoteAddr(), string(buf[:nBytes]))
				// if strings.Contains(string(buf), "aurum") {
				// 	lgr.Printf("Got aurum related message")
				// }
				conn.Write(buf[:nBytes])
			}
		}
	}
}

func main() {
	fl := Flags{
		help:       getopt.BoolLong("help", '?', "help"),
		debug:      getopt.BoolLong("debug", 'd', "debug"),
		version:    getopt.BoolLong("version", 'v', "version"),
		height:     getopt.BoolLong("height", 'h', "height"),
		genesis:    getopt.BoolLong("genesis", 'g', "genesis"),
		test:       getopt.BoolLong("test", 't', "test mode"),
		logs:       getopt.StringLong("log", 'l', "logs.txt", "log file"),
		port:       getopt.StringLong("port", 'p', "13131", "port"),
		interval:   getopt.StringLong("interval", 'i', "", "production interval"),
		initSupply: getopt.Uint64Long("supply", 'y', 0, "initial supply"),
		numBlocks:  getopt.Uint64Long("blocks", 'b', 0, "number of blocks to generate"),
	}
	getopt.Lookup('l').SetOptional()
	getopt.Parse()

	if *fl.help {
		getopt.Usage()
		os.Exit(0)
	}

	if *fl.version {
		fmt.Printf("Aurum producer version: %d\n", version)
		os.Exit(0)
	}

	var lgr = log.New(ioutil.Discard, "MAIN_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	if *fl.debug {
		lgr.SetOutput(os.Stdout)
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
		err = producer.Airdrop(ledger, metadataTable, genesisBlock)
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

	var addr = "localhost:"
	var ln net.Listener
	// var dataLock sync.Mutex
	if getopt.IsSet('p') {
		addr += *fl.port
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			lgr.Fatalf("failed to start tcp listener: %s", err.Error())
		}
		defer func() {
			lgr.Printf("closing listener.")
			ln.Close()
		}()
		lgr.Printf("Server listening on port %s.", *fl.port)
		// contractChan := make(chan []accounts.Contract)
		// TODO: isolate and test this function separately
		go RunServer(ln, *fl.debug)
	}

	productionInterval, err := time.ParseDuration(*fl.interval)
	if err != nil {
		lgr.Fatalln("failed to parse block production interval")
	} else {
		lgr.Println("block production interval: " + *fl.interval)
	}
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledger, metadataTable)
	if err != nil {
		lgr.Fatalf("failed to retrieve youngest block: %s\n", err)
	}
	var intervalChannel = make(chan bool)
	var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
	timeSince := time.Since(lastTimestamp)
	if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
		go triggerInterval(intervalChannel, time.Duration(0))
	} else {
		diff := productionInterval.Nanoseconds() - timeSince.Nanoseconds()
		go triggerInterval(intervalChannel, time.Duration(diff))
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	var numBlocksGenerated uint64

	// Main loop
	for {
		var dataPool []accounts.Contract
		var ms runtime.MemStats
		var chainHeight = youngestBlockHeader.Height
		select {
		case <-intervalChannel:
			lgr.Printf("block ready for production: #%d\n", chainHeight+1)
			if newBlock, err := producer.CreateBlock(version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), dataPool); err != nil {
				lgr.Fatalf("failed to add block %s", err.Error())
				os.Exit(1)
			} else {
				// TODO: make account.Validate only validate the transaction
				// TODO: table should be updated in separate call, after AddBlock
				// TODO: use a sync.Mutex.Lock()/Unlock() for editing tables
				if err := blockchain.AddBlock(newBlock, ledger, metadataTable); err != nil {
					lgr.Fatalf("failed to add block: %s", err.Error())
					os.Exit(1)
				} else {
					lgr.Printf("block produced: #%d\n", chainHeight+1)
					numBlocksGenerated++
					youngestBlockHeader = newBlock.GetHeader()
					go triggerInterval(intervalChannel, productionInterval)
					runtime.ReadMemStats(&ms)
				}
			}
		case <-signalCh:
			fmt.Print("\r")
			lgr.Println("Interrupt signal encountered, program terminating.")
			return
		}
		// useful commands: go run -gcflags='-m -m' main.go -d -t --interval=2000ms
		// lgr.Printf("Bytes of allocated heap objects: %d", ms.Alloc)
		// lgr.Printf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
		// lgr.Printf("Count of heap objects allocated: %d", ms.Mallocs)
		// lgr.Printf("Count of heap objects freed: %d", ms.Frees)
		if *fl.test || (getopt.IsSet('b') && (numBlocksGenerated >= *fl.numBlocks)) {
			break
		}
	}
}

func triggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	time.Sleep(productionInterval)
	intervalChannel <- true
}

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

// 	// Close the server
// 	ln.Close()
// }
