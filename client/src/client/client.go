// This contains the tools to allow a user to send relevant information off to a producer
//
// It also contains the functions for accepting user input and displaying information, a primitive console UI
package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// This will check if the client is connected to the internet
//
// Will return relevant error if not connected
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

// This stores any user input inside of text until a new line is entered
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

// Establishes connection to addr with Dial and sends data to address, returns number of bytes written and any errors
func SendToProducer(buf []byte, addr string) (int, error) {
	// Opens a connection, if connection fails, return 0 and error
	conn, err:= net.Dial("tcp", addr)
	if err != nil {
		conn.Close()
		return 0, err
	}
	n, err := conn.Write(buf)
	// Close connection, return counter and no error
	conn.Close()
	return n, nil
}

// This clears the console screen of any present text, uses system dependent clear sceen command
//
// This will also place a new seperator at the top of the screen
func ClearScreen() {
	// On non-windows systems, the clear command clears the screen
	cmd := exec.Command("clear")
	// If the operating system is actually windows, change this to cls (clear screen)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	}
	// Sets the output of this command to the command line, and executes
	cmd.Stdout = os.Stdout
	cmd.Run()
	// Prints a seperator at the top of the screen
	fmt.Println("#############################################################################")
	/*==ALTERNATIVE OPTIONS====================================================================
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	fmt.Println("=============================================================================")
	fmt.Println("-----------------------------------------------------------------------------")
	fmt.Println("_____________________________________________________________________________")
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	==========================================================================================*/
}

// This will print all relevant commands into the console
func PrintHelp() {
	fmt.Println("\tbalance\t\t\t\tcheck your wallet balance")								// balance
	fmt.Println("\tclear\t\t\t\tclears the screen of all previous output")            	// clear
	fmt.Println("\thelp\t\t\t\tprints all avalible commands and description of each") 	// help
	fmt.Println("\tmoreinfo\t\t\tprints link to project_aurum github page")           	// moreinfo
	fmt.Println("\tsend [recipient] [value]\tsend aurum to using their public key")	// send
	fmt.Println("\tq\t\t\t\tquits the program")                                       	// q
}

// This will print a link to the project github page into the console
func PrintGithubLink() {
	fmt.Println("https://github.com/SIGBlockchain/project_aurum for more info")
}