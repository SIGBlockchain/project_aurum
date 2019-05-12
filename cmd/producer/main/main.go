package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pborman/getopt"
)

type Flags struct {
	help       *bool
	debug      *bool
	version    *bool
	height     *bool
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
		logs:       getopt.StringLong("log", 'l', "logs.txt", "log file"),
		port:       getopt.StringLong("port", 'p', "13131", "port"),
		interval:   getopt.StringLong("interval", 'i', "0s", "production interval"),
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
