package privatekey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private key: %s", err)
	}
	encodedPrivateKey, err := Encode(privateKey)
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
			got, err := Decode(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private/public key pair")
	}
	aurumPVKey := AurumPrivateKey{
		privateKey,
	}

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate private/public key pair")
	}

	tests := []struct {
		name     string
		privKey  AurumPrivateKey
		privKey2 *ecdsa.PrivateKey
		want     bool
	}{
		{
			"Equals",
			aurumPVKey,
			privateKey,
			true,
		},
		{
			"Not Equals",
			aurumPVKey,
			privateKey2,
			false,
		},
		{
			"Not Equals",
			aurumPVKey,
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.privKey.Equals(tt.privKey2); result != tt.want {
				t.Errorf("Failed to return %v (got %v) for private keys that are: %v", tt.want, result, tt.name)
			}
		})
	}
}
