package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	producer "github.com/SIGBlockchain/project_aurum/internal/producer/src/producer"
	"github.com/pborman/getopt"
)

func main() {
	// Command line parsing
	help := getopt.Bool('?', "Display Valid Flags")
	debug := getopt.BoolLong("debug", 'd', "Enable Debug Mode")
	globalhost := getopt.BoolLong("global", 'g', "Enable globalhost")
	logFile := getopt.StringLong("log", 'l', "", "Log File Location")
	port := getopt.StringLong("port", 'p', "13131", "Port Number")
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
	// NOTE: If this doesn't work, try deleting `localhost`

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

	// Initialize BP struct with listener and empty map
	bp := producer.BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        logger,
	}

	// Start listening for connections
	logger.Printf("Server listening on port %s.", *port)
	go bp.AcceptConnections()

	// Main loop
	bp.WorkLoop()

	// Close the server
	bp.Server.Close()
}