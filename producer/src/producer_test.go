package src

import (
	"net"
	"testing"
)

// Test will fail in airplane mode, or just remove wireless connection.
func TestCheckConnectivity(t *testing.T) {
	err := CheckConnectivity()
	if err != nil {
		t.Errorf("Internet connection check failed.")
	}
}

// Tests a single connection
func TestAcceptConnections(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:10000")
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
	}
	go bp.AcceptConnections()
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Errorf("Failed to connect to server")
	}
	contentsOfChannel := <-bp.NewConnection
	actual := contentsOfChannel.RemoteAddr().String()
	expected := conn.LocalAddr().String()
	if actual != expected {
		t.Errorf("Failed to store connection")
	}
	conn.Close()
	ln.Close()
}
