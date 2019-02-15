package main

import (
	"log"
	"net"
	"os"
)

// Initializes logger format
func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// Check to see if there is an internet connection
	err := CheckConnectivity()
	if err != nil {
		log.Fatalln("Connectivity check failed.")
	}
	log.Println("Connection check passed.")

	// Default port
	port := "13131"

	// Grabs port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Spin up server
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalln("Failed to start server.")
	}

	// Initialize BP struct with listener and empty map
	bp := BlockProducer{
		server:        ln,
		newConnection: make(chan net.Conn, 128),
	}

	// Start listening for connections
	log.Printf("Server listening on port %s.\n", port)
	go bp.AcceptConnections()

	// Main loop
	for {

	}

	// Close the server
	bp.server.Close()
}
