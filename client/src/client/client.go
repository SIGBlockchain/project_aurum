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

// Establishes connection to addr with Dial
// Return 0 and err if Dial fails
// Get the length of buf
// Write buf to the connection IN 1024 BYTE CHUNKS
// if any conn write call fails, return how many bytes you wrote and an error
// if everything works out fine, return how many bytes you wrote and a nil error
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

/*=================================================================================================
* Purpose: Clears the terminal of all previous text and adds a seperator to the top of the screen *
* Returns: Nothing                                                                                *
=================================================================================================*/
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

/*=================================================================================================
* Purpose: Prints all avalible commands into the terminal, with a brief description for the usage *
* of each. In alphabetical order                                                                  *
* Returns: Nothing																				  *
=================================================================================================*/
func PrintHelp() {
	fmt.Println("\tclear\t\tclears the screen of all previous output")            // clear
	fmt.Println("\thelp\t\tprints all avalible commands and description of each") // help
	fmt.Println("\tmoreinfo\tprints link to project_aurum github page")           // moreinfo
	fmt.Println("\tq\t\tquits the program")                                       // q
}

/*=================================================================================================
* Purpose: Prints link to project github page                                                     *
=================================================================================================*/
func PrintGithubLink() {
	fmt.Println("https://github.com/SIGBlockchain/project_aurum for more info")
}