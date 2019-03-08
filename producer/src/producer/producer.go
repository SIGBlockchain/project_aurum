package producer

import (
	"errors"
	"net"
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
func (bp *BlockProducer) WorkLoop() {
	// Creates signal
	signalCh := make(chan os.Signal, 1)
    signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case conn := <-bp.NewConnection:
			go bp.Handle(conn)
		// If an interrupt signal is encountered, exit
		case <-signalCh:
			return
		default:
			// Do other stuff
		}
	}
}