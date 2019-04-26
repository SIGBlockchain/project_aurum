package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"reflect"
	"testing"
)

func TestKeys(t *testing.T) {
	// expected keys
	expectedPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	testFile := "./keys.dat"

	// This should open the file and store the keys in string form
	err := StoreKey(expectedPrivKey, testFile)
	if err != nil {
		t.Errorf("Failed to store keys.")
	}

	// actual key
	actualPrivKey, err := GetKey(testFile)
	if err != nil {
		t.Errorf("Failed to retrieve keys")
	}

	// Gets the Public Keys from the Private Keys
	actualPublicKey := actualPrivKey.PublicKey
	expectedPublicKey := expectedPrivKey.PublicKey

	// Compares the D field of the Private Keys
	if actualPrivKey.D.Cmp(expectedPrivKey.D) != 0 {
		t.Errorf("Private Key from file does not match expected Private Key.")
	}

	// Compares the Big Ints inside of the Public Key field
	if actualPublicKey.X.Cmp(expectedPublicKey.X) != 0 || actualPublicKey.Y.Cmp(expectedPublicKey.Y) != 0 {
		t.Errorf("Public Key from file does not match expected Public Key.")
	}

	// Delete testFile
	os.Remove(testFile)
}

// Full test for encoding/decoding public keys
func TestEncoding(t *testing.T) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	encoded := EncodePublicKey(&public)
	decoded := DecodePublicKey(encoded)

	if !reflect.DeepEqual(public, *decoded) {
		t.Errorf("Keys do not match")
	}
	reEncoded := EncodePublicKey(decoded)
	if !reflect.DeepEqual(reEncoded, encoded) {
		t.Errorf("Encoded keys do not match")
	}
}

func TestDecodePrivateKey(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private key: %s", err)
	}
	encodedPrivateKey, err := EncodePrivateKey(privateKey)
	if err != nil {
		t.Errorf("Failed to encode private key")
	}
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ecdsa.PrivateKey
		wantErr bool
	}{
		{
			args: args{
				key: encodedPrivateKey,
			},
			want:    privateKey,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodePrivateKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodePrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
