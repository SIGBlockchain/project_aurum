package main

import (
	"log"
	"os"
	"github.com/pborman/getopt/v2"
	"bytes"
	//"fmt"
	//"strings"

	client "../client"
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
		filepath := os.Getenv("GOPATH") + "/src/project_aurum/client/logs"
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
			logger.Println(filepath)
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

	var userInput string
	for {
		if client.GetUserInput(&userInput, os.Stdin) != nil {
			logger.Fatalln("Error getting input.")
		}

		if userInput == "q" {
			logger.Println("Exiting program.\nGoodbye")
			os.Exit(0)
		} else if userInput == "clear" {
			client.ClearScreen()
		} else if userInput == "help" {
			client.PrintHelp()
		} else if userInput == "moreinfo" {
			client.PrintGithubLink()
		}
	}
}
