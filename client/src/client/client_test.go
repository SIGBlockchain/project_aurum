package client

import (
	"testing"
	"bytes"
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

	var stdin bytes.Buffer
	stdin.Write([]byte("TEST\n"))

	var user_input string
	if GetUserInput(&user_input, &stdin) != nil {
		t.Errorf("User Input Check Failed.")		
	}

	if user_input != "TEST" {
		t.Errorf("User Input Check Failed.")
	}
}