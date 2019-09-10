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
	input := [][]byte{}
	expected := GetMerkleRootHash(input)

	if len(input) != len(expected) {
		t.Errorf("Error! GetMerkelRootHash does not return an empty slice on input of empty slice")
	}

	input = [][]byte{[]byte("transaction")}
	expected = New(New(input[0]))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on single byte slice")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}

	input = [][]byte{[]byte("transaction1"), []byte("transaction2")}
	concat := append(New(New(input[0])), New(New(input[1]))...)
	expected = New(New(concat))
	actual = GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on two byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}

	input = [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3")}
	concat1 := New(New(append(New(New(input[0])), New(New(input[1]))...)))
	concat2 := New(New(append(New(New(input[2])), New(New(input[2]))...)))
	expected = New(New(append(concat1, concat2...)))
	actual = GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}

	input = [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3"), []byte("transaction4")}
	concat1 = New(New(append(New(New(input[0])), New(New(input[1]))...)))
	concat2 = New(New(append(New(New(input[2])), New(New(input[3]))...)))
	expected = New(New(append(concat1, concat2...)))
	actual = GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
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
