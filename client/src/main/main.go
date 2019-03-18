package main

import (
	"log"
	"os"
	"github.com/pborman/getopt/v2"
	"bytes"
	"fmt"
	"strings"

	"github.com/SIGBlockchain/project_aurum/client/src/client"
)

func main() {
	// List of Options
	helpFlag := getopt.Bool('?', "Display Valid Flags")
	debugFlag := getopt.BoolLong("debug", 'd', "Enable Debug Mode")
	logFile := getopt.StringLong("logfile", 'l', "", "Log File Destination")
	getopt.CommandLine.Lookup('l').SetOptional()
	getopt.Parse()
	// If the help flag is on, print usage to os.Stdout
	if *helpFlag == true {
		getopt.Usage()
		os.Exit(0)
	}
	logger := log.New(os.Stdout, "LOG: ", log.Ldate | log.Lmicroseconds | log.Lshortfile)
	
	// If the debug flag is not on, the logger is set to a dummy buffer, which stores the input
	if *debugFlag == false {
		var buffer bytes.Buffer
		logger = log.New(&buffer, "LOG: ", log.Ldate | log.Lmicroseconds | log.Lshortfile)
	}
	// If the log flag is on, it will send the logs to a file in client/logs
	if getopt.CommandLine.Lookup('l').Count() > 0 {
		filepath := os.Getenv("GOPATH") + "/src/github.com/SIGBlockchain/project_aurum/client/logs"
		os.Mkdir(filepath, 0777)
		// If no filename is given, logs.txt
		if *logFile == "" {
			filepath += "/logs.txt"
		// Otherwise the custom filename is used
		} else {
			filepath += "/" + *logFile
		}
		f, err := os.OpenFile(filepath, os.O_RDWR | os.O_CREATE, 0666)
		defer f.Close()

		// If there is any error, do not set the logger. Log an error messgae
		if err != nil {
			logger.Fatalln(err)
		} else {
			logger = log.New(f, "LOG: ", log.Ldate | log.Lmicroseconds | log.Lshortfile)	
		}
	}
	// Clears the screen before program execution
	client.ClearScreen()

	// Check to see if there is an internet connection
	err := client.CheckConnection()
	//err := error(nil) // This is used for offline testing
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Println("Connection check passed.")

	// This string contains the entire input
	var userInput string
	for {
		// If there are any errors in getting input, end program execution
		if client.GetUserInput(&userInput, os.Stdin) != nil {
			logger.Fatalln("Error getting input.")
		}
		// inputs holds the individual arguments of a command
		inputs := strings.Split(userInput, " ")
		// Switch checks the first argument of a command
		switch inputs[0] {
      
		// 'q' command exits the program
		case "q":
			logger.Println("Exiting program.\nGoodbye")
			os.Exit(0)
		// 'clear' command clears the command window of all previous text, adds upper divider
		case "clear":
			client.ClearScreen()
		// 'help' command prints all the avalible commands, and a brief description
		case "help":
			client.PrintHelp()
		// 'moreinfo' command prints the link to the github page
		case "moreinfo":
			client.PrintGithubLink()
		// 'send' command sends a given value to a given recipient
		//		send [recipient] [value]
		case "send":
			// 'send' requires 3 arguments at a minimum, otherwise ignore command
			if len(inputs) != 3 {
				fmt.Println("ERROR: Improper Use of send Command\n\tsend [recipient] [value]")
				break
			}
			// This 64 bit integer holds the recipient's public key
			var recipient string
			_, err := fmt.Sscanf(inputs[1], "%s", &recipient)
			if err != nil {
				logger.Println("ERROR: Attempt to collect recipient string failed")
				break
			}
			// This 64 bit integer holds the value for the contract
			var value int64
			_, err = fmt.Sscanf(inputs[2], "%d", &value)
			if err != nil {
				logger.Println("ERROR: Attempt to collect value integer failed")
				break
			}
			logger.Println("Accepted send as valid input")
			// Pass recipient, value to function
		case "balance":
			// Insert function to get balance
			logger.Println("Accepted send as valid input")
		// If first argument of a command matches no valid command, print an error message
		default:
			fmt.Println("Invalid command \"" + userInput + "\" rejected\n\tUse \"help\" to see available commands")
			logger.Println("Invalid command \"" + userInput + "\" rejected")
	}
}
