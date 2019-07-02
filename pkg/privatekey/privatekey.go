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

// Equals returns true if the given two *ecdsa.PrivateKey are equal
func (pvKey AurumPrivateKey) Equals(pvKey2 *ecdsa.PrivateKey) bool {
	if pvKey.Key == nil || pvKey2 == nil {
		return false
	}

	if pvKey.Key.PublicKey.X.Cmp(pvKey2.PublicKey.X) != 0 ||
		pvKey.Key.PublicKey.Y.Cmp(pvKey2.PublicKey.Y) != 0 ||
		pvKey.Key.PublicKey.Curve != pvKey2.PublicKey.Curve ||
		pvKey.Key.D.Cmp(pvKey2.D) != 0 {
		return false
	}

	return true
}
