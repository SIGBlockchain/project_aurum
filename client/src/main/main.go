package main

import (
	"log"
	"os"
	"strings"

	client "../client"
)

// Initializes logger format
func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {
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

		if strings.Compare(userInput, "q") == 0 {
			log.Println("Exiting program.\nGoodbye")
			os.Exit(0)
		}
	}
}
