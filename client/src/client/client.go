package client

import (
	"errors"
	"net"
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

