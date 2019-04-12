// This contains all necessary tools for the producer to accept connections and process the recieved data
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

// This stores connection information for the producer
type BlockProducer struct {
	Server        net.Listener
	NewConnection chan net.Conn
}

// This checks if the producer is connected to the internet
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("Connectivity check failed.")
	}
	conn.Close()
	return nil
}

// This will accept any incoming connections
func (bp *BlockProducer) AcceptConnections() {
	for {
		conn, err := bp.Server.Accept()
		if err != nil {
			return
		}
		bp.NewConnection <- conn
	}
}

// Handles incoming connections, accepting _ of at most 1024 bytes
func (bp *BlockProducer) Handle(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}
	conn.Write(buf)
	conn.Close()
}

// The main work loop which handles communication, block production, and ledger maintenance
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
