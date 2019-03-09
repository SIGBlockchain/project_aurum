package producer

import (
	"fmt"
	"log"
	"net"
	"os"
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

// Sends a message to the listener and checks to see if it echoes back
func TestHandler(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:10000")
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
	}
	go bp.AcceptConnections()
	go bp.WorkLoop(log.New(os.Stderr, "", log.Lshortfile))
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Errorf("Failed to connect to server")
	}
	expected := []byte("This is a test.")
	conn.Write(expected)
	fmt.Println("Made it")
	actual := make([]byte, 128)
	_, readErr := conn.Read(actual)
	fmt.Println("Made it")
	if readErr != nil {
		t.Errorf("Failed to read from socket.")
	}
	for i, b := range expected {
		if b != actual[i] {
			t.Errorf("Message mismatch.")
		}
	}
	conn.Close()
	ln.Close()
}
