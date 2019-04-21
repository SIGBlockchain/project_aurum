// Package contains all the necessary tools to interact with and store keys
package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"os"
)

// Stores a key into a given file
func StoreKey(p *ecdsa.PrivateKey, filename string) error {
	// Opens the file, if it does not exist the O_CREATE flag tells it to create the file otherwise overwrite file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	// Checks if the opening was successful
	if err != nil {
		return err
	}

	// Encodes the private key
	x509Encoded, _ := x509.MarshalECPrivateKey(p)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	// jsonStruct, will contain the information inside the json file
	type jsonStruct struct {
		Private string
	}

	// Converts the encoded byte string into a string so it can be used in the json struct
	encodedStr := hex.EncodeToString(pemEncoded)
	j := jsonStruct{
		Private: encodedStr,
	}

	jbyte, err := json.Marshal(j)

	// Write into the file that was given
	_, err = file.Write(jbyte)

	// Checks if the writing was successful
	if err != nil {
		return err
	}

	return nil
}

// Gets the key held in a given file
func GetKey(filename string) (*ecdsa.PrivateKey, error) {
	privateKey := new(ecdsa.PrivateKey)
	// Reads json file into b_string, if any errors occur, abort
	bString, err := ioutil.ReadFile(filename)
	if err != nil {
		return privateKey, err
	}
	// Create message variable to hold json data
	type keyStruct struct {
		Private string
	}
	var keys keyStruct
	// Load json data from text file
	err = json.Unmarshal(bString, &keys)
	// Decodes the private key
	privString, err := hex.DecodeString(keys.Private)
	block, _ := pem.Decode(privString)
	x509Encoded := block.Bytes
	privateKey, _ = x509.ParseECPrivateKey(x509Encoded)
	// Returns private key
	return privateKey, err
}

// Returns the PEM-Encoded byte slice from a given public key
func EncodePublicKey(key *ecdsa.PublicKey) []byte {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
}

// Returns the public key from a given PEM-Encoded byte slice representation of the public key
func DecodePublicKey(key []byte) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode(key)
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	return genericPublicKey.(*ecdsa.PublicKey)
}

// Returns the PEM-Encoded byte slice from a given private key
func EncodePrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	x509EncodedPriv, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return []byte{}, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509EncodedPriv}), nil
}

// Returns the private key from a given PEM-Encoded byte slice representation of the private key
func DecodePrivateKey(key []byte) (*ecdsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode(key)
	x509EncodedPriv := keyBlock.Bytes
	return x509.ParseECPrivateKey(x509EncodedPriv)
}
