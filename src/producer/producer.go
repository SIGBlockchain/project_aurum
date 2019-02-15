package producer

import (
	"log"
	"net"
)

func init() {
	// Initializes logger format
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func CheckConnectivity() {
	// Purpose: Checks to see if there is an internet connection established
	// Parameters: None
	// Returns: Void
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		panic("No internet connection detected.")
	}
	conn.Close()
}

func main() {
	defer func() {
		r := recover()
		if r != nil {
			log.Fatalln("Connection check failed.")
		}
	}()
	CheckConnectivity()
	log.Println("Connection check successful.")
}
