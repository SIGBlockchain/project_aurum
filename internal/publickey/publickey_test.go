package publickey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"testing"
)

// Full test for encoding/decoding public keys
func TestEncoding(t *testing.T) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	encoded := Encode(&public)
	decoded := Decode(encoded)

	if !reflect.DeepEqual(public, *decoded) {
		t.Errorf("Keys do not match")
	}
	reEncoded := Encode(decoded)
	if !reflect.DeepEqual(reEncoded, encoded) {
		t.Errorf("Encoded keys do not match")
	}
}

func TestEquals(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private/public key pair")
	}
	aurumPBKey := AurumPublicKey{
		&privateKey.PublicKey,
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
