package main

import (
	"fmt"
	"log"
	"os"

	testfunctions "github.com/SIGBlockchain/project_aurum/test/src/testfunctions"
	"github.com/pborman/getopt"
)

func main() {

	// The uint32 `n` should be retrieved as an command line argument
	helpFlag := getopt.Bool('?', "display help")
	logFile := getopt.StringLong("log", 'l', "", "log file location")
	n := getopt.Uint32('n', 0, "number of private keys to be generated")
	getopt.CommandLine.Lookup('l').SetOptional()
	getopt.Parse()
	if getopt.CommandLine.Lookup('n').Count() != 1 || *helpFlag == true {
		getopt.Usage()
		os.Exit(0)
	}

	// Setup logger
	logger := log.New(os.Stdout, "LOG:", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if getopt.CommandLine.Lookup('l').Count() > 0 {
		filepath := os.Getenv("GOPATH") + "/src/github.com/SIGBlockchain/project_aurum/test/logs"
		os.Mkdir(filepath, 0777)
		if *logFile == "" {
			filepath += "/logs.txt"
		} else {
			filepath += "/" + *logFile
		}
		f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
		defer f.Close()
		if err != nil {
			logger.Fatalln(fmt.Errorf("Failed to open file: %s", err))
		} else {
			logger = log.New(f, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
		}
	}
	// Run GenerateNRandomKeys
	if err := testfunctions.GenerateNRandomKeys("keys.json", *n); err != nil {
		logger.Fatalln(fmt.Errorf("Failed to generate random private keys: %s", err))
	}
	// Run AirdropNContracts
	// Run GenerateGenesisBlock
	// Run producer package main.go as a goroutine
	/*
		The producer should then recover contract and blockchain metadata
		This requires knowledge of the blockchain file name, which will be an
		environmental argument supplied to that main function
	*/
	// Loop `i` times, making random contracts and sending them to producer
}
