package test_functions

import (
	"errors"
)

/*
Should generate N random ecdsa private keys.
JSON marshall them and store them in filename
*/
func GenerateNRandomKeys(filename string, n uint32) error {
	return errors.New("Incomplete function")
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
