package block

import (
	"bytes"           // for comparing []bytes
	"encoding/binary" // for encoding/decoding
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
		PreviousHash:   []byte{'g', 'u', 'a', 'v', 'a'},
		MerkleRootHash: []byte{'g', 'r', 'a', 'p', 'e'},
		Timestamp:      nowTime,
		Data:           [][]byte{}, // need to add
	}
	// now use the serialize function
	serial := b.Serialize()
	// will keep track of current position in the slice
	pos := 0

	// check version
	blockVersion := binary.LittleEndian.Uint32(serial[0:4])
	if blockVersion != b.Version {
		t.Errorf("version does not match")
	}

	// check height
	blockHeight := binary.LittleEndian.Uint64(serial[4:12])
	if blockHeight != b.Height {
		t.Errorf("height does not match")
	}

	//check Timestamp
	blockTimestamp := binary.LittleEndian.Uint64(serial[12:20])
	if int64(blockTimestamp) != b.Timestamp {
		t.Errorf("time stamps do not match")
	}
	pos = 20

	// check previousHash
	blockPrevHash := serial[pos : pos+len(b.PreviousHash)]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("Previous hashes do not match\n")
	}
	pos = pos + len(b.PreviousHash)

	//check merkleRootHash
	blockMerkleHash := serial[pos : pos+len(b.MerkleRootHash)]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("Merkle hashes do not match\n")
	}

}
