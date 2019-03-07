package client

import (
	"testing"
	"bytes"
	"reflect"
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

// Test simulates user input, ensures correct collection of command line arguments
func TestReadCmdLineArgs(t *testing.T) {
	var testread bytes.Buffer

	// Empty Case
	testread.Write([]byte(""))
	slice_empty, err_empty := readCmdLineArgs(&testread)
	if len(slice_empty) > 0 {
		t.Errorf("Empty input yielded nonempty arguments, Test Failed.")	
	}
	if err_empty == nil {
		t.Errorf("Empty input yielded no errors, Test Failed.")
	}

	// Singleton, just executable call, Case
	testread.Reset()
	testread.Write([]byte("./client"))
	slice_single, err_single := readCmdLineArgs(&testread)
	if len(slice_single) > 0 {
		t.Errorf("Single input yielded nonempty arguments, Test Failed.")	
	}
	if err_single != nil {
		t.Errorf("Single input yielded an error, Test Failed.")
	}

	// Other Case
	testread.Reset()
	testread.Write([]byte("./client -test -arguments"))
	slice_other, err_other := readCmdLineArgs(&testread)
	expected_args := []string{"-test", "-arguments"}
	if reflect.DeepEqual(expected_args, slice_other) {
		t.Errorf("Normal input yielded incorrect arguments, Test Failed.")	
	}
	if err_other != nil {
		t.Errorf("Normal input yielded an error, Test Failed.")
	}
}

// Test simulates user input, ensures correct processing of command line arguments
func TestProcessCmdLineArgs(t *testing.T) {
	var testread bytes.Buffer

	// Empty Case
	testread.Write([]byte(""))
	err_empty := ProcessCmdLineArgs(&testread)
	if err_empty == nil {
		t.Errorf("Empty input yielded no errors, Test Failed.")
	}

	// Singleton, just executable call, Case
	testread.Reset()
	testread.Write([]byte("./client"))
	err_single := ProcessCmdLineArgs(&testread)
	if err_single != nil {
		t.Errorf("Single input yielded an error, Test Failed.")
	}

	// Other Case
	testread.Reset()
	testread.Write([]byte("./client -test"))
	err_other := ProcessCmdLineArgs(&testread)
	if err_other == nil {
		t.Errorf("Invalid input yielded no error, Test Failed.")
	}
}