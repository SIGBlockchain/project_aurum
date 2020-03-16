package publickey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"math/big"
 

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
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
	b := make([]byte, 1+len(key.X.Bytes())+len(key.Y.Bytes()))
	b[0] = byte(1)
	copy(b[1:33], key.X.Bytes())
	copy(b[33:65], key.Y.Bytes())

	return b, nil
}

// Decode returns the public key from a given PEM-Encoded byte slice representation of the public key or a non-nil error if fail
func Decode(key []byte) (*ecdsa.PublicKey, error) {
	if key == nil {
		return nil, errors.New("Could not return the decoded public key - the key value is nil")
	}
    if(key[0]!=1){
		return nil, errors.New("Could not return the decoded public key - the key is not uncompressed public key")
	}

	n := new(big.Int)
	m := new(big.Int)


	xByte, err := n.SetString(hex.EncodeToString(key[1:33]), 16)

	if !err {
		return nil, errors.New("Could not convert string to int")
	}

	yByte, ok := m.SetString(hex.EncodeToString(key[33:65]), 16)
	if !ok {
		return nil, errors.New("Could not convert string to int")
	}
	
	priv := ecdsa.PublicKey{Curve: elliptic.P256(), X: xByte, Y: yByte}
	return &priv, nil
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
