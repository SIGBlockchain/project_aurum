package main

import (
	"log"
	"os"

	client "../client"
)

// Initializes logger format
func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	client.ClearScreen()
	// Check to see if there is an internet connection
	err := client.CheckConnection()
	if err != nil {
		log.Fatalln("Connectivity check failed.")
	}
	log.Println("Connection check passed.")

	var userInput string
	for {
		if client.GetUserInput(&userInput, os.Stdin) != nil {
			log.Fatalln("Error getting input.")
		}

		if userInput == "q" {
			log.Println("Exiting program.\nGoodbye")
			os.Exit(0)
		} else if userInput == "clear" {
			client.ClearScreen()
		} else if userInput == "help" {
			client.PrintHelp()
		} else if userInput == "moreinfo" {
			if client.GoToWebpage() != nil {
				log.Println("Attempt to open github page failed.")
			}
		}
	}
}
