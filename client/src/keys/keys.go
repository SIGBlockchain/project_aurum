package keys

import (
	"crypto/ecdsa"
	"errors"
)

func StoreKey(p *ecdsa.PrivateKey, filename string) error {
	// TODO





	e := errors.New("Error") // change to meaningful text
	return e
}

func GetKey(filename string) *ecdsa.PrivateKey, error {
	// TODO




	e := errors.New("Error") // change to meaningful text
	return privateKey, e
}
