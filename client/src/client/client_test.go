package client

import (
	"bytes"
	"net"
	"testing"
	"time"

	producer "../../../producer/src/producer"
)

// Test will fail in airplane mode, or just remove wireless connection.
func TestCheckConnection(t *testing.T) {
	err := CheckConnection()
	if err != nil {
		t.Errorf("Internet connection check failed.")
	}
}

// Test will simulate user input and ensure that the function will collect the correct string
func TestGetUserInput(t *testing.T) {

	var testread bytes.Buffer
	testread.Write([]byte("TEST\n"))

	var user_input string
	if GetUserInput(&user_input, &testread) != nil {
		t.Errorf("User Input Check Failed.")
	}

	if user_input != "TEST" {
		t.Errorf("User Input Check Failed.")
	}
}

// Test send to producer with small max length message for one send
func TestSendToProducer(t *testing.T) {
	sz := 1024
	testbuf := make([]byte, sz)
	for i, _ := range testbuf {
		testbuf[i] = 1
	}
	addr := "localhost:8080"
	ln, err := net.Listen("tcp", addr)
	bp := producer.BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
	}
	go bp.AcceptConnections()
	time.Sleep(1)
	if err != nil {
		t.Errorf("Failed to set up listener")
	}
	n, err := SendToProducer(testbuf, addr)
	if err != nil {
		t.Errorf("Failed to send to producer")
	}
	if n != sz {
		t.Errorf("Did not write all bytes to connection")
	}
	ln.Close()
}

// Test send to producer with large message
func TestSendToProducerWithLargeMessage(t *testing.T) {
	sz := 4096
	testbuf := make([]byte, sz)
	for i, _ := range testbuf {
		testbuf[i] = 1
	}
	addr := "localhost:8080"
	ln, err := net.Listen("tcp", addr)
	bp := producer.BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
	}
	go bp.AcceptConnections()
	time.Sleep(1)
	if err != nil {
		t.Errorf("Failed to set up listener")
	}
	n, err := SendToProducer(testbuf, addr)
	if err != nil {
		t.Errorf("Failed to send to producer")
	}
	if n != sz {
		t.Errorf("Did not write all bytes to connection")
	}
	ln.Close()
}