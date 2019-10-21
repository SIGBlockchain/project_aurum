package publickey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
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
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(nil)
	// test that Encoding returns an error for bad input
	if err == nil {
		t.Errorf("Expected err to not be nil, but it was...")
	}
	encoded, err := Encode(&public)
	// test that Encoding does not receive an error for valid input
	if err != nil {
		t.Errorf("Received an error for valid input")
	}
	x509EncodedPub, _ = x509.MarshalPKIXPublicKey(&public)
	x509EncodedPub = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	// test that Encoding results match
	if !reflect.DeepEqual(x509EncodedPub, encoded) {
		t.Errorf("Encoding does not match")
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
	localDecoded, _ := pem.Decode(encoded)
	x509EncodedPub := localDecoded.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	// test that decodings match
	if !reflect.DeepEqual(genericPublicKey, decoded) {
		t.Errorf("Keys do not match after decode")
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
