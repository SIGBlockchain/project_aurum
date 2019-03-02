package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

/*=====================================================================
Purpose: Checks if the user has an internet connection, this is done by
	attempting to connect to google, and returning the error message
Returns: Error message if user has faulty internet connection, nil
	otherwise
=====================================================================*/
func CheckConnection() error {
	// Creates a connection conn, and stores any errors in err
	conn, err := net.Dial("tcp", "www.google.com:80")
	// If err is not nil, then there was an error
	if err != nil {
		return errors.New("Connectivity check failed.")
	}
	// Close the connection
	conn.Close()
	return nil
}

/*=================================================================================================
* Purpose: Collects user input from command line, returns as a string                             *
* Returns: A string, holding the user input                                                       *
=================================================================================================*/
func GetUserInput(text *string, reader io.Reader) error {
	// Creates a reader object, using bufio library
	fmt.Print("[aurum client] >> ")
	// Stores user input until \n, stores into text
	var err error
	newReader := bufio.NewReader(reader)
	*text, err = newReader.ReadString('\n')
	// Ensures no newline characters in input
	*text = strings.Replace(*text, "\n", "", -1)
	return err
}
