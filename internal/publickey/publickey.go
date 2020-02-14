package publickey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/btcsuite/btcutil/base58"
)

// AurumPublicKey struct holds public key for user
type AurumPublicKey struct {
	Key   *ecdsa.PublicKey
	Bytes []byte
	Hex   string
	Hash  []byte
}

// NewFromPublic returns an AurumPublicKey given a ecdsa.PublicKey
func NewFromPublic(key *ecdsa.PublicKey) (AurumPublicKey, error) {
	bytes, err := Encode(key)
	if err != nil {
		return AurumPublicKey{}, errors.New("Failed to encode public key inside of private key")
	}
	return AurumPublicKey{Key: key, Bytes: bytes, Hex: hex.EncodeToString(bytes), Hash: hashing.New(bytes)}, nil
}

// New returns an AurumPublicKey given a ecdsa.PrivateKey
func New(key *ecdsa.PrivateKey) (AurumPublicKey, error) {
	//try type assertion
	k, ok := key.Public().(*ecdsa.PublicKey)
	if !ok {
		return AurumPublicKey{}, errors.New("Cannot extract a ecdsa public key from private key given")
	}
	return NewFromPublic(k)
}

// Encode returns the PEM-Encoded byte slice from a given public key or a non-nil error if fail
func Encode(key *ecdsa.PublicKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("Could not return the encoded public key - the key value is nil")
	}
	b := append(key.X.Bytes(), key.Y.Bytes()...)
	k := "1" + string(b)
	c:= []byte(k)
	b58EncodedPub := base58.Encode(c)
    b58PrefixEncoded := []byte(b58EncodedPub)
	return b58PrefixEncoded, nil
}

// Decode returns the public key from a given PEM-Encoded byte slice representation of the public key or a non-nil error if fail
func Decode(key []byte) (*ecdsa.PublicKey, error) {
	if key == nil {
		return nil, errors.New("Could not return the decoded public key - the key value is nil")
	}
	blockPub, _ := pem.Decode(key)
	// pem.Decode will return nil for the first value if no PEM data is found. This would be bad
	if blockPub == nil {
		return nil, errors.New("Could not return the public key - failed to PEM decode in preparation x509 encode")
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
