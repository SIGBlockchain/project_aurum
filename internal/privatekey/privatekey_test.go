package privatekey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"
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

func TestGenerateNRandomKeys(t *testing.T) {
	type args struct {
		filename string
		n        uint32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test with N = 0",
			args: args{
				filename: "testfile.json",
				n:        0,
			},
			wantErr: true,
			// Error should say "must generate at least 1 key"
		},
		{
			name: "Test with N = 100",
			args: args{
				filename: "testfile.json",
				n:        100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateNRandomKeys(tt.args.filename, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNRandomKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, err := os.Stat(tt.args.filename); os.IsNotExist(err) {
				t.Errorf("Test file for keys not detected: %s", err)
			}
			if tt.args.n > 0 {
				jsonFile, err := os.Open(tt.args.filename)
				if err != nil {
					t.Errorf("Failed to open json file: %s", err)
				}
				defer jsonFile.Close()
				var keys []string
				byteKeys, err := ioutil.ReadAll(jsonFile)
				if err != nil {
					t.Errorf("Failed to read in keys from json file: %s", err)
				}
				err = json.Unmarshal(byteKeys, &keys)
				if err != nil {
					t.Errorf("Failed to unmarshall keys because: %s", err)
				}
				if uint32(len(keys)) != tt.args.n {
					t.Errorf("Number of private keys do not match n: %d", len(keys))
				}

			}
		})
	}
	if err := os.Remove("testfile.json"); err != nil {
		t.Errorf("Failed to remove file: %s because: %s", "testfile.json", err)
	}
}
