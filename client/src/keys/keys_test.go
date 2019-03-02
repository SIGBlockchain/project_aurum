package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"testing"
)

func TestKeys(t *testing.T) {
	// expected keys
	expectedPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	testFile := "./keys.dat"

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

    actualPublicKey := actualPrivKey.PublicKey
    expectedPublicKey := expectedPrivKey.PublicKey

    if actualPrivKey.D.Cmp(expectedPrivKey.D) != 0 {
    	t.Errorf("Private Key from file does not match expected Private Key.")
    }
    if actualPublicKey.X.Cmp(expectedPublicKey.X) != 0 && actualPublicKey.Y.Cmp(expectedPublicKey.Y) != 0 {
        t.Errorf("Public Key from file does not match expected Public Key.")
    }

	// Delete testFile
	os.Remove(testFile)
}
