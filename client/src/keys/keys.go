package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
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

/*
func GetKey(filename string) *ecdsa.PrivateKey, error {
	// TODO




	e := errors.New("Error") // change to meaningful text
	return privateKey, e
}
*/
