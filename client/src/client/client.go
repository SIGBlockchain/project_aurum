package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"os"
	"os/exec"
	"runtime"
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
	fmt.Println("\tclear\t\tclears the screen of all previous output") // clear
	fmt.Println("\thelp\t\tprints all avalible commands and description of each") // help
	fmt.Println("\tmoreinfo\topens a browser and opens up project_aurum github page") // moreinfo
	fmt.Println("\tq\t\tquits the program") // q
}

/*=================================================================================================
* Purpose: Sends user to project github page                                                      *
* Returns: An Error message if failed, nil otherwise                                              *
=================================================================================================*/
func GoToWebpage() error{
	// On non-windows systems, the open command opens a URL with default browser
	cmd := exec.Command("xdg-open", "https://github.com/SIGBlockchain/project_aurum")
	// If the operating system is actually windows, change this to start
	if runtime.GOOS == "windows" {
		cmd = exec.Command("start", "https://github.com/SIGBlockchain/project_aurum")
	}
	// Sets the output of this command to the command line, and executes
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}
