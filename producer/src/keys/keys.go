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

func StoreKey(p *ecdsa.PrivateKey, filename string) error {
	// Opens the file, if it does not exist the O_CREATE flag tells it to create the file otherwise overwrite file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)

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

/*=================================================================================================
* Purpose: Converts encoded strings in a json file into the private key                           *
* Returns:                                                                                        *
=================================================================================================*/
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

/*
EncodePublicKey encodes given public key and returns its
PEM-Encoded byte slice form
*/
func EncodePublicKey(key *ecdsa.PublicKey) []byte {
	return []byte{}
}

/*
DecodePublicKey takes a PEM-Encoded key and returns
an ecdsa PublicKey Struct
*/
func DecodePublicKey(key []byte) ecdsa.PublicKey {
	return ecdsa.PublicKey{}
}
