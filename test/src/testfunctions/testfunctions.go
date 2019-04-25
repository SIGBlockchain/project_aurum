package testfunctions

import (
	"errors"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
)

// Generates n random keys and writes them to the file at filename
func GenerateNRandomKeys(filename string, n uint32) error {
	// Opens the file, if it does not exist the O_CREATE flag tells it to create the file otherwise overwrite file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	if n < 1 {
		return errors.New("Must generate at least one private key")
	}
	// Checks if the opening was successful
	if err != nil {
		return err
	}
	// jsonStruct, will contain the information inside the json file
	type jsonStruct struct {
		Privates []ecdsa.PrivateKey
	}

	var keys []string 	// This will hold all pem encoded private key strings
	var i uint32 = 0 	// Iterator, is uint32 to be able to compare with n

	// Create n private keys
	for ; i < n; i++ {
		p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}
		// Encodes the private key
		x509Encoded, _ := x509.MarshalECPrivateKey(p)
		pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
		// Converts the encoded byte string into a string so it can be used in the json struct
		encodedStr := hex.EncodeToString(pemEncoded)
		// Add the encoded byte string to the string slice
		keys = append(keys, encodedStr)
	}

	// Write strings to file
	jbyte, err := json.Marshal(keys)

	// Checks if marshalling was successful
	if err != nil {
		return err
	}
	// Write into the file that was given
	_, err = file.Write(jbyte)

	// Checks if the writing was successful
	if err != nil {
		return err
	}

	return nil
}

/*
Should read from json filename, make N contracts, each with
a `null` sender, and return the contracts as an array
Not feasible to complete this until we're done with accounts
*/
func AirdropNContracts(filename string, n uint32) error {
	return errors.New("Incomplete function")
}

// Not feasible to complete until we are done with accounts
func GenerateGenesisBlock(blockchainFile string) error {
	return errors.New("Incomplete function")
}
