package publickey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// AurumPublicKey struct holds public key for user
type AurumPublicKey struct {
	Key *ecdsa.PublicKey
}

// Encode returns the PEM-Encoded byte slice from a given public key or a non-nil error if fail
func Encode(key *ecdsa.PublicKey) ([]byte, error) {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub}), nil
}

// Decode returns the public key from a given PEM-Encoded byte slice representation of the public key or a non-nil error if fail
func Decode(key []byte) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode(key)
	// pem.Decode will return nil for the first value if no PEM data is found. This would be bad
	if blockPub == nil {
		return nil, errors.New("Could not return the public key - the key value is nil")
	}

	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}
	return genericPublicKey.(*ecdsa.PublicKey), nil
}

// Equals returns true if the given two *ecdsa.PublicKey are equal
func (pbKey AurumPublicKey) Equals(pbKey2 *ecdsa.PublicKey) bool {
	if pbKey.Key == nil || pbKey2 == nil {
		return false
	}

	if pbKey.Key.X.Cmp(pbKey2.X) != 0 ||
		pbKey.Key.Y.Cmp(pbKey2.Y) != 0 ||
		pbKey.Key.Curve != pbKey2.Curve {
		return false
	}

	return true
}
