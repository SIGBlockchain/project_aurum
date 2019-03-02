package keys

import (
	"testing"
	"os"
	"math/big"
)

func TestGetKeys(t *testing.T) {
	// Checks if fatal exit when file does not exist. WILL CRASH PROGRAM
	//priv_key, pub_key := GetKeys("./.keys.test.dat.json")

	// Creates test file, inserts test values
	f, create_err := os.Create("./.keys.test.dat.json")
    if create_err != nil {
        t.Errorf("Failure in creating test file.")
    }
    _, write_err := f.WriteString(`{"private" : "0000000000000008","public" : "000000000000000f"}`)
    if write_err != nil {
        t.Errorf("Failure in writing test file.")
    }
    // Fills expected private/public values
    expectedPrivate := new(big.Int)
    expectedPublic := new(big.Int)
    expectedPrivate.SetString("0000000000000008", 16)
    expectedPublic.SetString("000000000000000f", 16)

    // Gets actual private/public keys
    priv_key, pub_key := GetKeys("./.keys.test.dat.json")
    // Asserts if public and private keys are equal to their expected values
    if pub_key.Cmp(expectedPublic) != 0 {
    	t.Errorf("Public Key %d does not match expected key %d.", pub_key, expectedPublic)
    }
    if priv_key.Cmp(expectedPrivate) != 0{
    	t.Errorf("Private Key %d does not match expected key %d.", priv_key, expectedPrivate)
    }
    os.Remove("./.keys.test.dat.json")
}