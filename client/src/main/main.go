package main

import (
	"log"
	"fmt"
	"strings"
	"os"

	client "project_aurum/client/src/client"
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

	var user_input string
	for {
		if client.GetUserInput(&user_input, os.Stdin) != nil {
			break;
		}

		if strings.Compare(user_input, "q") == 0 {
			fmt.Println("goodbye")
			break;
		}
	}
}