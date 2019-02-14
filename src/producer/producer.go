package main

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// Purpose: stores communication information
type BlockProducer struct {
	connections    map[net.Conn]bool
	server         net.Listener
	newConnection  chan net.Conn
	deadConnection chan net.Conn
}

// Purpose: Checks to see if there is an internet connection established
// Parameters: None
// Returns: Void
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("Connectivity check failed.")
	}
	conn.Close()
	return nil
}

// Purpose: accepts incoming connections
func (bp *BlockProducer) AcceptConnections() {
	for {
		conn, err := bp.server.Accept()
		if err != nil {
			fmt.Println("Client failed to connect.")
		}
		log.Printf("%s connected.\n", conn.LocalAddr())
		bp.newConnection <- conn
	}
}
