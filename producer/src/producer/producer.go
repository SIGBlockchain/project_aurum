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
}

// Purpose: Checks to see if there is an internet connection established
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("Connectivity check failed.")
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

// Handles incoming Connections
// Currently this is simply echoing messages back
// In the future this will need to support messages of size > 1024
// This can be done by reading in fragments

/*
Two types of messages:
Balance:
Given a public key, query the database for the current account balance
** successful key find -> 2 fields, true and value
** failed key find -> 1 field, false + greylist
Sent a Contract:
Validate contract, add to contract pool
*/
func (bp *BlockProducer) Handle(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}
	conn.Write(buf)
	conn.Close()
}

// The main work loop
// Handles communication, block production, and ledger maintenance

/*
cases:
new connection -> should call Handle but not as a separate thread
sig int -> should gracefully exit
default -> checks to see if it's time to update the ledger
** if it is time, then create a block and add it
*** it is time if (time now - time of last block) = some time interval
*** Creating a block should involve extracting from the contract pool
*** May require knowing what the last block is
*/
func (bp *BlockProducer) WorkLoop(logger *log.Logger) {
	// Creates signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case conn := <-bp.NewConnection:
			go bp.Handle(conn)
		// If an interrupt signal is encountered, exit
		case <-signalCh:
			// If loop is exited properly, interrupt signal had been recieved
			fmt.Print("\r")
			logger.Println("Interrupt signal encountered, program terminating.")
			return
		default:
			// Do other stuff
		}
	}
}
