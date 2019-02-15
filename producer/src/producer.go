package producer

import (
	"errors"
	"net"
)

// Purpose: stores communication information
type BlockProducer struct {
	server        net.Listener
	newConnection chan net.Conn
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
		conn, err := bp.server.Accept()
		if err != nil {
			return
		}
		bp.newConnection <- conn
	}
}
