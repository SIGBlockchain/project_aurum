package block

import (
	"bytes"           // for comparing []bytes
	"encoding/binary" // for encoding/decoding
	"reflect"         // to get data type
	"testing"         // testing
	"time"            // to get time stamp
)

func TestSerialize(t *testing.T) {
	// get time stamp
	ti := time.Now()
	nowTime := ti.UnixNano()

	// create the block
	b := Block{
		Version:        3,
		Height:         300,
		PreviousHash:   []byte{'g', 'u', 'a', 'v', 'a', 'p', 'i', 'n', 'e', 'a', 'p', 'p', 'l', 'e', 'm', 'a', 'n', 'g', 'o', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'a', 'b', 'c'},
		MerkleRootHash: []byte{'g', 'r', 'a', 'p', 'e', 'w', 'a', 't', 'e', 'r', 'm', 'e', 'l', 'o', 'n', 'c', 'o', 'c', 'o', 'n', 'u', 't', 'l', 'e', 'm', 'o', 'n', 's', 'a', 'b', 'c', 'd'},
		Timestamp:      nowTime,
		Data:           [][]byte{}, // need to add
	}
	// now use the serialize function
	serial := b.Serialize()

	// indicies are fixed since we know what the max sizes are going to be

	// check Version
	blockVersion := binary.LittleEndian.Uint32(serial[0:4])
	if blockVersion != b.Version {
		t.Errorf("Version does not match")
	}

	// check Height
	blockHeight := binary.LittleEndian.Uint64(serial[4:12])
	if blockHeight != b.Height {
		t.Errorf("Height does not match")
	}

	//check Timestamp
	blockTimestamp := binary.LittleEndian.Uint64(serial[12:20])
	if int64(blockTimestamp) != b.Timestamp {
		t.Errorf("Timestamps do not match")
	}

	// check PreviousHash
	blockPrevHash := serial[20:52]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("PreviousHashes do not match")
	}

	//check MerkleRootHash
	blockMerkleHash := serial[52:84]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("MerkleRootHashes do not match")
	}

}

// tests HashSHA256 function
func TestHashSHA256(t *testing.T) {
	data := []byte{'s', 'a', 'm'}
	result := HashSHA256(data)
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
	doubleHash := HashSHA256(HashSHA256(input[0]))
	concat := append(doubleHash, doubleHash...)
	expected := HashSHA256(HashSHA256(concat))
	actual := GetMerkleRootHash(input)

	if (!bytes.Equal(expected, actual)){
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on single byte slice")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashDoubleInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2")}
	concat := append(HashSHA256(HashSHA256(input[0])) , HashSHA256(HashSHA256(input[1]))...)
	expected := HashSHA256(HashSHA256(concat))
	actual := GetMerkleRootHash(input)

	if (!bytes.Equal(expected, actual)){
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on two byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashTripleInput(t *testing.T){
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3")}
	concat1 :=  HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[0])) , HashSHA256(HashSHA256(input[1]))...)))
	concat2 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[2])) , HashSHA256(HashSHA256(input[2]))...)))
	expected := HashSHA256(HashSHA256(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if (!bytes.Equal(expected, actual)){
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashQuadInput(t *testing.T){
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3"), []byte("transaction4")}
	concat1 :=  HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[0])) , HashSHA256(HashSHA256(input[1]))...)))
	concat2 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[2])) , HashSHA256(HashSHA256(input[3]))...)))
	expected := HashSHA256(HashSHA256(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if (!bytes.Equal(expected, actual)){
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}