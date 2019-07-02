package publickey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

type AurumPublicKey struct {
	Key *ecdsa.PublicKey
}

// Returns the PEM-Encoded byte slice from a given public key
func Encode(key *ecdsa.PublicKey) []byte {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
}

// Returns the public key from a given PEM-Encoded byte slice representation of the public key
func Decode(key []byte) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode(key)
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	return genericPublicKey.(*ecdsa.PublicKey)
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
