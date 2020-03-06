package publickey

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
)

func TestNew(t *testing.T) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	k, ok := private.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Errorf("Cannot extract a ecdsa public key from private key given")
	}
	bytes, err := Encode(k)
	if err != nil {
		t.Errorf("Failed to encode public key inside of private key")
	}
	expected := AurumPublicKey{
		Key:   k,
		Bytes: bytes,
		Hex:   hex.EncodeToString(bytes),
		Hash:  hashing.New(bytes)}

	actual, err := New(private)
	if err != nil {
		t.Errorf("Failed to create new Aurum: %s", err.Error())
	}

	if !reflect.DeepEqual(actual.Bytes, expected.Bytes) {
		t.Errorf("Acutal Bytes does not equal expected Bytes:\nActual: %v\nExpected: %v", actual.Bytes, expected.Bytes)
	}
	if !reflect.DeepEqual(actual.Hash, expected.Hash) {
		t.Errorf("Acutal Hash does not equal expected Hash:\nActual: %v\nExpected: %v", actual.Hash, expected.Hash)
	}
	if actual.Hex != expected.Hex {
		t.Errorf("Acutal Hex does not equal expected Hex:\nActual: %v\nExpected: %v", actual.Hex, expected.Hex)
	}
	if !reflect.DeepEqual(actual.Key, expected.Key) {
		t.Errorf("Acutal Key does not equal expected Key:\nActual: %v\nExpected: %v", actual.Key, expected.Key)
	}

}

// test that public keys can be encoded properly
func TestEncoding(t *testing.T) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	encodedPublic, err := Encode(&public)

	if err != nil {
		t.Errorf("Got error: %s\n", err.Error())
	}
	if encodedPublic[0] != byte(1) {
		t.Errorf("Encoding is missing a byte prefix of 1, instead got: %v\n", encodedPublic[0])
	}
	if bytes.Compare(encodedPublic[1:33], public.X.Bytes()) != 0 {
		t.Errorf("Encoding of X from public key is mssing")
	}
	if bytes.Compare(encodedPublic[33:65], public.Y.Bytes()) != 0 {
		t.Errorf("Encoding of Y from public key is mssing")
	}
	if len(encodedPublic) != 65 {
		t.Errorf("Expected legnth of 65, instead got: %d\n", len(encodedPublic))
	}
}

// test that public keys can be decoded properly
func TestDecoding(t *testing.T) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	encoded, _ := Encode(&public)
	decoded, err := Decode(nil)
	// test that Decoding returns an error
	if err == nil {
		t.Errorf("Expected err to not be nil, but it was...")
	}
	decoded, err = Decode(encoded)
	// test that Decoding does not return an error for valid input
	if err != nil {
		t.Errorf("Expected err to not be nil, but it was...")
	}

	// test that decodings match
	
	if !reflect.DeepEqual(public, decoded) {
		t.Errorf("Keys do not match after decode\n. %v != %v\n", public, decoded)
	}
}

// tests for Equals function for public keys
func TestEquals(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private/public key pair")
	}
	aurumPBKey, err := New(privateKey)
	if err != nil {
		t.Errorf("Failed to create new Aurum public key")
	}

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private/public key pair")
	}

	tests := []struct {
		name    string
		pubKey  AurumPublicKey
		pubKey2 *ecdsa.PublicKey
		want    bool
	}{
		{
			"Equals",
			aurumPBKey,
			&privateKey.PublicKey,
			true,
		},
		{
			"Not Equals",
			aurumPBKey,
			&privateKey2.PublicKey,
			false,
		},
		{
			"Not Equals",
			aurumPBKey,
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.pubKey.Equals(tt.pubKey2); result != tt.want {
				t.Errorf("Failed to return %v (got %v) for public keys that are: %v", tt.want, result, tt.name)
			}
		})
	}
}
