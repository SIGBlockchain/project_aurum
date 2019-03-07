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

// Test simulates user input, ensures correct processing of command line arguments
func TestProcessCmdLineArgs(t *testing.T) {	
	// Empty Case, just executable call
	err_empty := ProcessCmdLineArgs([]string{})
	if err_empty != nil {
		t.Errorf("No arguments yielded an error, Test Failed.")
	}

	// Other Case
	err_other := ProcessCmdLineArgs([]string{"-test"})
	if err_other == nil {
		t.Errorf("Invalid input yielded no error, Test Failed.")
	}
}