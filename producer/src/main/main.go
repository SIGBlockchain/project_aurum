package main

import (
	"log"
	"net"
	"os"

	producer "github.com/SIGBlockchain/project_aurum/producer/src/producer"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	// Check to see if there is an internet connection
	err := producer.CheckConnectivity()
	if err != nil {
		logger.Fatalln("Connectivity check failed.")

	}
	logger.Println("Connection check passed.")

	// Default port
	port := "13131"

	// Grabs port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Spin up server
	// NOTE: If this doesn't work, try deleting `localhost`
	ln, err := net.Listen("tcp", "localhost:"+port)
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
	logger.Printf("Server listening on port %s.\n", port)
	go bp.AcceptConnections()

	// Main loop
	bp.WorkLoop()

	// Close the server
	bp.Server.Close()
}
