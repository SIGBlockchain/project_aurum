package testfunctions

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

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
//			if tt.args.n == 0 && err != errors.New("Must generate at least one private key") {
//				t.Errorf("Wrong error message generated. Should say: %s, instead says: %s", "\"Must generate at least one private key\"", err)
//			}
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
