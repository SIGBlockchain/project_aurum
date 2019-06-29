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

func TestGetMerkleRootHashEmptyInput(t *testing.T) {
	input := [][]byte{}
	result := GetMerkleRootHash(input)

	if len(input) != len(result) {
		t.Errorf("Error! GetMerkelRootHash does not return an empty slice on input of empty slice")
	}
}

func TestGetMerkleRootHashSinlgeInput(t *testing.T) {
	input := [][]byte{[]byte("transaction")}
	expected := New(New(input[0]))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on single byte slice")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashDoubleInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2")}
	concat := append(New(New(input[0])), New(New(input[1]))...)
	expected := New(New(concat))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on two byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashTripleInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3")}
	concat1 := New(New(append(New(New(input[0])), New(New(input[1]))...)))
	concat2 := New(New(append(New(New(input[2])), New(New(input[2]))...)))
	expected := New(New(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashQuadInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3"), []byte("transaction4")}
	concat1 := New(New(append(New(New(input[0])), New(New(input[1]))...)))
	concat2 := New(New(append(New(New(input[2])), New(New(input[3]))...)))
	expected := New(New(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}
