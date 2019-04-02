package producer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// Purpose: stores communication information
type BlockProducer struct {
	Server        net.Listener
	NewConnection chan net.Conn
	// Add ledger name, metadata name, and contract table name
	// Slice of Contracts representing contract pool
}

// Purpose: Checks to see if there is an internet connection established
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("connectivity check failed")
	}
	conn.Close()
	return nil
}

// Purpose: Accepts incoming connections
func (bp *BlockProducer) AcceptConnections() {
	for {
		conn, err := bp.Server.Accept()
		if err != nil {
			return
		}
		bp.NewConnection <- conn
	}
}

/*
Should contain version, payload type, payload size
*/
type Header struct{}

/*
Messages have headers and payloads
Payloads should correspond to message type
*/
type Message struct{}

// Handles incoming Connections
// Currently this is simply echoing messages back
// In the future this will need to support messages of size > 1024
// This can be done by reading in fragments
func (bp *BlockProducer) Handle(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}
	/*
		Check if keybytes are present.
		If they aren't, close the connection, add IP to greylist

		Check type of message (Balance or Contract)

		If Balance, query the database with the public key
		If no public key exists, send message with `invalid public key`
		Otherwise send balance back

		If Contract, validate contract first
		If validation fails, send message with `invalid contract`
		Otherwise, add contract to contract pool and send verification message
	*/
	conn.Write(buf)
	conn.Close()
}

// The main work loop
// Handles communication, block production, and ledger maintenance
func (bp *BlockProducer) WorkLoop(logger *log.Logger) {
	// Creates signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	// TODO: Extract last block blockhash, timestamp
	// Initialize contract pool
	for {
		select {
		case conn := <-bp.NewConnection:
			go bp.Handle(conn) // TODO: remove `go`
		// If an interrupt signal is encountered, exit
		case <-signalCh:
			// If loop is exited properly, interrupt signal had been recieved
			fmt.Print("\r")
			logger.Println("Interrupt signal encountered, program terminating.")
			return
		default:
			/*
				TODO:
				Check to see if it's time to make a block
				Block interval = (timeNow - timeSinceLastBlock)
				If it is time, make the block from the contract pool
				(Merkle Root, add block)
				Reset last block metadata
			*/
		}
	}
}
