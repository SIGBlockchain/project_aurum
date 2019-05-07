package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	producer "github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/pborman/getopt"
)

var version = uint16(1)
var ledger = "blockchain.dat"
var metadata = "metadata.tab"

func main() {
	// Command line parsing
	help := getopt.Bool('?', "help")
	debug := getopt.BoolLong("debug", 'd', "debug mode")
	globalhost := getopt.BoolLong("global", 'g', "enable global host")
	logFile := getopt.StringLong("log", 'l', "", "log file")
	port := getopt.StringLong("port", 'p', "13131", "port")
	interval := getopt.StringLong("interval", 'i', "", "production interval")
	// initialAurumSupply := getopt.Uint64Long("supply", 's', 0, "initial aurum supply")
	getopt.CommandLine.Lookup('l').SetOptional()
	getopt.Parse()

	// If the help flag is on, print usage to os.Stdout
	if *help == true {
		getopt.Usage()
		os.Exit(0)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if *debug == false {
		var buffer bytes.Buffer
		logger = log.New(&buffer, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	}

	if getopt.CommandLine.Lookup('l').Count() > 0 {
		filepath := os.Getenv("GOPATH") + "/src/github.com/SIGBlockchain/project_aurum/logs"
		os.Mkdir(filepath, 0777)
		// If no filename is given, logs.txt
		if *logFile == "" {
			filepath += "/producer_logs.txt"
			// Otherwise the custom filename is used
		} else {
			filepath += "/" + *logFile
		}
		f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
		defer f.Close()

		// If there is any error, do not set the logger. Log an error messgae
		if err != nil {
			logger.Fatalln(err)
		} else {
			logger = log.New(f, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
		}
	}

	// Check to see if there is an internet connection
	err := producer.CheckConnectivity()
	if err != nil {
		logger.Fatalln("Connectivity check failed.")
	}
	logger.Println("Connection check passed.")

	// Spin up server
	var addr string
	if *globalhost {
		addr = fmt.Sprintf(":")
		logger.Println("Listening on all IP addresses")
	} else {
		addr = fmt.Sprintf("localhost:")
		logger.Println("Listening on local IP addresses")

	}
	ln, err := net.Listen("tcp", addr+*port)
	if err != nil {
		logger.Fatalln("Failed to start server.")
	}

	// Start listening for connections
	logger.Printf("Server listening on port %s.", *port)
	newDataChan := make(chan producer.Data)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			logger.Printf("%s connection\n", conn.RemoteAddr())
			go func() {
				buf := make([]byte, 1024)
				_, err := conn.Read(buf)
				if err != nil {
					return
				}
				// Handle message
			}()
		}
	}()

	// Main loop
	timerChan := make(chan bool)
	// var chainHeight uint64
	var dataPool []producer.Data
	productionInterval, err := time.ParseDuration(*interval)
	if err != nil {
		logger.Fatalln("failed to parse interval")
	}
	// youngestBlock, err := blockchain.GetYoungestBlock(ledger, metadata)
	// if err != nil {
	// 	logger.Fatalf("failed to retrieve youngest block header: %s\n", err)
	// }
	for {
		select {
		case newData := <-newDataChan:
			dataPool = append(dataPool, newData)
		case <-timerChan:
			// newBlock, _ := producer.CreateBlock(version, chainHeight+1, block.HashBlock(youngestBlock), dataPool)
			// blockchain.AddBlock(newBlock, ledger, metadata)
			dataPool = nil
			go func() {
				time.AfterFunc(productionInterval, func() {
					<-timerChan
				})
			}()
		}
	}

	// Close the server
	ln.Close()
}
