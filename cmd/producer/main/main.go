package main

/*
Usage:
Must have genesis_hashes.txt file, or will fail.
Run `go run main.go` with -g --supply=[x] for genesis
Running `go run main.go` with --interval=[x]ms will generate a block every x milliseconds (must be milliseconds) indefinitely
Running that command with -t flag will generate only one block and exit.
Running that command with --numBlocks=x will generate only x number of blocks.
Running that command with --port=x will accept connections on port x.
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/genesis"
	"github.com/SIGBlockchain/project_aurum/internal/producer"

	"github.com/pborman/getopt"
)

var version = uint16(1)
var ledger = "blockchain.dat"
var metadataTable = constants.MetadataTable

func main() {
	// Flag struct parsing
	fl := producer.Flags{
		Help:        getopt.BoolLong("help", '?', "help"),
		Debug:       getopt.BoolLong("debug", 'd', "debug"),
		Version:     getopt.BoolLong("version", 'v', "version"),
		Height:      getopt.BoolLong("height", 'h', "height"),
		Genesis:     getopt.BoolLong("genesis", 'g', "genesis"),
		Test:        getopt.BoolLong("test", 't', "test mode"),
		Globalhost:  getopt.BoolLong("globalhost", 'o', "global host mode"),
		MemoryStats: getopt.BoolLong("memstats", 'm', "gather memory statistics"),
		Logs:        getopt.StringLong("log", 'l', "logs.txt", "log file"),
		Port:        getopt.StringLong("port", 'p', "13131", "port"),
		Interval:    getopt.StringLong("interval", 'i', "", "production interval"),
		InitSupply:  getopt.Uint64Long("supply", 'y', 0, "initial supply"),
		NumBlocks:   getopt.Uint64Long("blocks", 'b', 0, "number of blocks to generate"),
	}
	getopt.Lookup('l').SetOptional()
	getopt.Parse()

	if *fl.Help {
		getopt.Usage()
		os.Exit(0)
	}

	if *fl.Version {
		fmt.Printf("Aurum producer version: %d\n", version)
		os.Exit(0)
	}

	var lgr = log.New(ioutil.Discard, "MAIN_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	if *fl.Debug {
		lgr.SetOutput(os.Stdout)
	}

	if *fl.Genesis {
		genesisHashes, err := genesis.ReadGenesisHashes()
		if err != nil {
			lgr.Fatalf("failed to read in genesis hashes because %s", err.Error())
		}
		if !getopt.IsSet('y') {
			lgr.Fatalln("must set initial aurum supply")
		}
		if *fl.InitSupply < uint64(len(genesisHashes)) {
			lgr.Fatalln("must allocate at least 1 aurum per genesis hash")
		}
		genesisBlock, err := genesis.BringOnTheGenesis(genesisHashes, *fl.InitSupply)
		if err != nil {
			lgr.Fatalf("failed to create genesis block because: %s", err.Error())
		}
		if err := blockchain.Airdrop(ledger, metadataTable, constants.AccountsTable, genesisBlock); err != nil {
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

	if *fl.Globalhost {
		addr = ":"
		lgr.Println("Listening on all IP addresses")
	} else {
		addr = "localhost:"
		lgr.Println("Listening on local IP addresses")
	}

	// Set up byte channel and start listening on port
	// var byteChan chan []byte
	byteChan := make(chan []byte)
	if getopt.IsSet('p') {
		addr += *fl.Port
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			lgr.Fatalf("failed to start tcp listener: %s", err.Error())
		}
		defer func() {
			lgr.Printf("closing listener.")
			ln.Close()
		}()
		go producer.RunServer(ln, byteChan, *fl.Debug)
		lgr.Printf("Server listening on port %s.", *fl.Port)
	}

	// Start producing blocks
	producer.ProduceBlocks(byteChan, fl, getopt.IsSet('b'))
}
