package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

func TestKeys(t *testing.T) {
	// expected keys
	expectedPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	testFile := "keys.dat"

	// This should open the file and store the keys in string form
	err := StoreKey(expectedPrivKey, testFile)
	if err != nil {
		t.Errorf("Failed to store keys.")
	}

	// actual key
	actualPrivKey, err := GetKey(testFile)
	if err != nil {
		t.Errorf("Failed to retrieve keys")
	}

	if expectedPrivKey != actualPrivKey {
		t.Errorf("Contents of file does not match the ")
	}
}
