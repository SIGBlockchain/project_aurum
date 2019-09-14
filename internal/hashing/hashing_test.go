package hashing

import (
	"bytes"
	"reflect"
	"testing"
)

// tests New function
func TestNew(t *testing.T) {
	data := []byte{'s', 'a', 'm'}
	result := New(data)
	var byte32_variable []byte
	// checks if data was hashed by comparing data types
	if reflect.TypeOf(result).Kind() != reflect.TypeOf(byte32_variable).Kind() {
		t.Errorf("Error. Data types do not match.")
	}
	if len(result) != 32 {
		t.Errorf("Error. Data is not 32 bytes long.")
	}
}

func TestGetMerkleRootHashInput(t *testing.T) {
	tests := []struct {
		name string
		input [][]byte
		expected []byte
	}{
		{
			"Empty Slice",
			[][]byte{},
			nil,
		},
		{
			"Size One Slice",
			[][]byte{[]byte("transaction")},
			New(New([]byte("transaction"))),
		},
		{
			"Size Two Slice",
			[][]byte{[]byte("transaction1"), []byte("transaction2")},
			New(New(append(New(New([]byte("transaction1"))), New(New([]byte("transaction2")))...))),

		},
		{
			"Size Three Slice",
			[][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3")},
			New(New(append(New(New(append(New(New([]byte("transaction1"))), New(New([]byte("transaction2")))...))), New(New(append(New(New([]byte("transaction3"))), New(New([]byte("transaction3")))...)))...))),

		},
		{
			"Size Four Slice",
			[][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3"), []byte("transaction4")},
			New(New(append(New(New(append(New(New([]byte("transaction1"))), New(New([]byte("transaction2")))...))), New(New(append(New(New([]byte("transaction3"))), New(New([]byte("transaction4")))...)))...))),

		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !bytes.Equal(GetMerkleRootHash(tt.input),tt.expected) {
				t.Errorf("Failed on %v - results are not equal", tt.name)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	hash1 := SHA256Hash{
		New([]byte{'a'}),
	}

	tests := []struct {
		name string
		h1   SHA256Hash
		h2   []byte
		want bool
	}{
		{
			"Equal",
			hash1,
			[]byte{'a'},
			true,
		},
		{
			"Not equal",
			hash1,
			[]byte{'z'},
			false,
		},
		{
			"Not equal",
			hash1,
			[]byte{},
			false,
		},
		{
			"Not equal",
			hash1,
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.h1.Equals(tt.h2); result != tt.want {
				t.Errorf("Failed to return %v (got %v) for hashes that are: %v", tt.want, result, tt.name)
			}
		})
	}
}
