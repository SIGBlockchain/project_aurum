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
