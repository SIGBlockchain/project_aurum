package privatekey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

type AurumPrivateKey struct {
	Key *ecdsa.PrivateKey
}

// Returns the PEM-Encoded byte slice from a given private key
func Encode(key *ecdsa.PrivateKey) ([]byte, error) {
	x509EncodedPriv, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return []byte{}, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509EncodedPriv}), nil
}

// Returns the private key from a given PEM-Encoded byte slice representation of the private key
func Decode(key []byte) (*ecdsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode(key)
	x509EncodedPriv := keyBlock.Bytes
	return x509.ParseECPrivateKey(x509EncodedPriv)
}
