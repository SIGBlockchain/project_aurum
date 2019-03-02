package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"io/ioutil"
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
	if b_string, err := ioutil.ReadFile(filename); err == nil {
		// Create message variable to hold json data
		type keyStruct struct {
			Private string
		}
		var keys keyStruct
		// Load json data from text file
		err = json.Unmarshal(b_string, &keys)
		// Decodes the private key
		priv_string, err := hex.DecodeString(keys.Private)
		block, _ := pem.Decode(priv_string)
		x509Encoded := block.Bytes
		privateKey, _ = x509.ParseECPrivateKey(x509Encoded)
		// Returns private key
		return privateKey, err
	}
	return privateKey, nil
}