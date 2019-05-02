package producer

import (
	"bytes"
	"log"
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
	var buffer bytes.Buffer
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        log.New(&buffer, "LOG:", log.Ldate),
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

// Sends a message to the listener and checks to see if it echoes back
func TestHandler(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:10000")
	var buffer bytes.Buffer
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        log.New(&buffer, "LOG:", log.Ldate),
	}
	go bp.AcceptConnections()
	go bp.WorkLoop()
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Errorf("Failed to connect to server")
	}
	expected := []byte("This is a test.")
	conn.Write(expected)
	actual := make([]byte, len(expected))
	_, readErr := conn.Read(actual)
	if readErr != nil {
		t.Errorf("Failed to read from socket.")
	}
	if bytes.Equal(expected, actual) == false {
		t.Errorf("Message mismatch")
	}
	conn.Close()
	ln.Close()
}
