package main

/*
Usage:
Must have genesis-hashes file, or will fail.
Run go run main.go with -g --supply=[x] for genesis
Running go run main.go with --interval=[x]ms will generate a block every x milliseconds (must be milliseconds) indefinitely
Running that command with -t flag will generate only one block and exit.
Running that command with --numBlocks=x will generate only x amount of blocks.
Running that command with --port=x will accept connections on port x.
*/

import (
	"bytes"
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
	help        *bool
	debug       *bool
	version     *bool
	height      *bool
	genesis     *bool
	test        *bool
	globalhost  *bool
	memoryStats *bool
	logs        *string
	port        *string
	interval    *string
	initSupply  *uint64
	numBlocks   *uint64
}

var version = uint16(1)
var ledger = "blockchain.dat"
var metadataTable = "metadata.tab"
var accountsTable = "accounts.tab"

func RunServer(ln net.Listener, bChan chan []byte, debug bool) {
	var lgr = log.New(ioutil.Discard, "SRVR_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if debug {
		lgr.SetOutput(os.Stdout)
	}
	for {
		lgr.Println("Waiting for connection...")
		conn, err := ln.Accept()
		var nRcvd int
		buf := make([]byte, 1024)
		if err != nil {
			lgr.Println("connection failed")
			goto End
		}
		lgr.Printf("%s connected\n", conn.RemoteAddr())
		defer conn.Close()
		nRcvd, err = conn.Read(buf)
		if err != nil {
			goto End
		}
		if nRcvd < 8 && (!bytes.Equal(buf[:8], producer.SecretBytes)) {
			conn.Write([]byte("No thanks.\n"))
			goto End
		} else {
			conn.Write([]byte("Thank you.\n"))
		}
		bChan <- buf[:nRcvd]
		goto End
	End:
		lgr.Println("Closing connection.")
		conn.Close()
	}
}

func ProduceBlocks(byteChan chan []byte, fl Flags, limit bool) {
	var lgr = log.New(ioutil.Discard, "PROD_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if *fl.debug {
		lgr.SetOutput(os.Stdout)
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
	go triggerInterval(intervalChannel, productionInterval)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	var numBlocksGenerated uint64

	var dataPool []accounts.Contract
	var ms runtime.MemStats

	// Main loop
	for {
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
				if err := blockchain.AddBlock(newBlock, ledger, metadataTable); err != nil {
					lgr.Fatalf("failed to add block: %s", err.Error())
					os.Exit(1)
				} else {
					lgr.Printf("block produced: #%d\n", chainHeight+1)
					numBlocksGenerated++
					youngestBlockHeader = newBlock.GetHeader()
					go triggerInterval(intervalChannel, productionInterval)
					dataPool = nil
					if *fl.memoryStats {
						runtime.ReadMemStats(&ms)
						printMemstats(ms)
					}
				}
			}
		case message := <-byteChan:
			lgr.Printf("Main received: %s\n", string(message))
			if message[8] == 1 {
				lgr.Println("Received contract")
				var newContract accounts.Contract
				if err := newContract.Deserialize(message[9:]); err == nil {
					dataPool = append(dataPool, newContract)
				}
			}
		case <-signalCh:
			fmt.Print("\r")
			lgr.Println("Interrupt signal encountered, program terminating.")
			return
		}
		if *fl.test {
			lgr.Printf("Test mode: breaking loop")
			break
		}
		if limit && (numBlocksGenerated >= *fl.numBlocks) {
			lgr.Printf("Limit reached: # blocks generated: %d, blocks desired: %d\n", numBlocksGenerated, *fl.numBlocks)
			break
		}
	}
}

func main() {
	fl := Flags{
		help:        getopt.BoolLong("help", '?', "help"),
		debug:       getopt.BoolLong("debug", 'd', "debug"),
		version:     getopt.BoolLong("version", 'v', "version"),
		height:      getopt.BoolLong("height", 'h', "height"),
		genesis:     getopt.BoolLong("genesis", 'g', "genesis"),
		test:        getopt.BoolLong("test", 't', "test mode"),
		globalhost:  getopt.BoolLong("globalhost", 'o', "global host mode"),
		memoryStats: getopt.BoolLong("memstats", 'm', "gather memory statistics"),
		logs:        getopt.StringLong("log", 'l', "logs.txt", "log file"),
		port:        getopt.StringLong("port", 'p', "13131", "port"),
		interval:    getopt.StringLong("interval", 'i', "", "production interval"),
		initSupply:  getopt.Uint64Long("supply", 'y', 0, "initial supply"),
		numBlocks:   getopt.Uint64Long("blocks", 'b', 0, "number of blocks to generate"),
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
		if err := producer.Airdrop(ledger, metadataTable, genesisBlock); err != nil {
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

	var addr string
	var ln net.Listener

	if *fl.globalhost {
		addr = ":"
		lgr.Println("Listening on all IP addresses")
	} else {
		addr = "localhost:"
		lgr.Println("Listening on local IP addresses")
	}

	var byteChan chan []byte
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
		go RunServer(ln, byteChan, *fl.debug)
		lgr.Printf("Server listening on port %s.", *fl.port)
	}

	ProduceBlocks(byteChan, fl, getopt.IsSet('b'))
}

func printMemstats(ms runtime.MemStats) {
	// useful commands: go run -gcflags='-m -m' main.go <main flags>
	fmt.Printf("Bytes of allocated heap objects: %d", ms.Alloc)
	fmt.Printf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
	fmt.Printf("Count of heap objects allocated: %d", ms.Mallocs)
	fmt.Printf("Count of heap objects freed: %d", ms.Frees)
}

func triggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	time.Sleep(productionInterval)
	intervalChannel <- true
}

// var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
// timeSince := time.Since(lastTimestamp)
// if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
// 	go triggerInterval(intervalChannel, time.Duration(0))
// } else {
// 	diff := productionInterval.Nanoseconds() - timeSince.Nanoseconds()
// 	go triggerInterval(intervalChannel, time.Duration(diff))
// }
