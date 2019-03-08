package main

import (
	"fmt"
	"log"
	"net"
	"os"

	producer "../producer"
)

// Initializes logger format
func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// Check to see if there is an internet connection
	err := producer.CheckConnectivity()
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
	// NOTE: If this doesn't work, try deleting `localhost`
	ln, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatalln("Failed to start server.")
	}

	// Initialize BP struct with listener and empty map
	bp := producer.BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
	}

	// Start listening for connections
	log.Printf("Server listening on port %s.\n", port)
	go bp.AcceptConnections()

	// Main loop
	bp.WorkLoop()
	// If loop is exited properly, interrupt signal had been recieved
	fmt.Print("\r")
	log.Println("Interrupt signal encountered, program terminating.\n")

	// Close the server
	bp.Server.Close()
}
