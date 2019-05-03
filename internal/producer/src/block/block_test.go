package block

import (
	"bytes"           // for comparing []bytes
	"encoding/binary" // for encoding/decoding
	"reflect"         // to get data type
	"testing"         // testing
	"time"            // to get time stamp

	"github.com/google/go-cmp/cmp"
)

func TestSerialize(t *testing.T) {
	// get time stamp
	ti := time.Now()
	nowTime := ti.UnixNano()

	// create the block
	b := Block{
		Version:        3,
		Height:         300,
		PreviousHash:   []byte("guavapineapplemango1234567890abc"),
		MerkleRootHash: []byte("grapewatermeloncoconut1emonsabcd"),
		Timestamp:      nowTime,
		Data:           [][]byte{{12, 3}, {132, 90, 23}, {23}},
	}
	// set data length
	b.DataLen = uint16(len(b.Data))
	// now use the serialize function
	serial := b.Serialize()
	// indicies are fixed since we know what the max sizes are going to be

	// check Version
	blockVersion := binary.LittleEndian.Uint16(serial[0:2])
	if blockVersion != b.Version {
		t.Errorf("Version does not match")
	}

	// check Height
	blockHeight := binary.LittleEndian.Uint64(serial[2:10])
	if blockHeight != b.Height {
		t.Errorf("Height does not match")
	}

	// check Timestamp
	blockTimestamp := binary.LittleEndian.Uint64(serial[10:18])
	if int64(blockTimestamp) != b.Timestamp {
		t.Errorf("Timestamps do not match")
	}

	// check PreviousHash
	blockPrevHash := serial[18:50]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("PreviousHashes do not match")
	}

	// check MerkleRootHash
	blockMerkleHash := serial[50:82]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("MerkleRootHashes do not match")
	}

	// check DataLen
	blockDataLen := binary.LittleEndian.Uint16(serial[82:84])
	if blockDataLen != b.DataLen {
		t.Errorf("DataLen does not match")
	}

	// check Data
	testslice := [][]byte{{12, 3}, {132, 90, 23}, {23}}
	dataLen := int(blockDataLen)
	blockData := make([][]byte, dataLen)
	index := 84

	for i := 0; i < dataLen; i++ {
		elementLen := int(serial[index])
		index += 2
		blockData[i] = serial[index : index+elementLen]
		index += elementLen
	}

	for i := 0; i < dataLen; i++ {
		if bytes.Compare(testslice[i], blockData[i]) != 0 {
			t.Errorf("Data does not match")
		}
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
	expected := HashSHA256(HashSHA256(input[0]))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on single byte slice")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashDoubleInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2")}
	concat := append(HashSHA256(HashSHA256(input[0])), HashSHA256(HashSHA256(input[1]))...)
	expected := HashSHA256(HashSHA256(concat))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on two byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashTripleInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3")}
	concat1 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[0])), HashSHA256(HashSHA256(input[1]))...)))
	concat2 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[2])), HashSHA256(HashSHA256(input[2]))...)))
	expected := HashSHA256(HashSHA256(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestGetMerkleRootHashQuadInput(t *testing.T) {
	input := [][]byte{[]byte("transaction1"), []byte("transaction2"), []byte("transaction3"), []byte("transaction4")}
	concat1 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[0])), HashSHA256(HashSHA256(input[1]))...)))
	concat2 := HashSHA256(HashSHA256(append(HashSHA256(HashSHA256(input[2])), HashSHA256(HashSHA256(input[3]))...)))
	expected := HashSHA256(HashSHA256(append(concat1, concat2...)))
	actual := GetMerkleRootHash(input)

	if !bytes.Equal(expected, actual) {
		t.Errorf("Error! GetMerkelRootHash does not produce correct result on three byte slices")
		t.Errorf("Expected != Actual")
		t.Errorf("%v != %v", expected, actual)
	}
}

func TestDeserialize(t *testing.T) {
	expected := Block{
		Version:        1,
		Height:         0,
		PreviousHash:   HashSHA256([]byte{'x'}),
		MerkleRootHash: HashSHA256([]byte{'q'}),
		Timestamp:      time.Now().UnixNano(),
		Data:           [][]byte{HashSHA256([]byte{'r'})},
	}
	expected.DataLen = uint16(len(expected.Data))
	intermed := expected.Serialize()
	actual := Deserialize(intermed)
	if !cmp.Equal(expected, actual) {
		t.Errorf("Blocks do not match")
	}

	//change itermed to see if that changes the deserialized block
	intermed[18] = uint8(21)
	intermed[54] = uint8(21)
	if !cmp.Equal(expected, actual) {
		t.Errorf("Blocks do not match")
	}

}