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
	pos = 20

	// check PreviousHash
	blockPrevHash := serial[pos : pos+len(b.PreviousHash)]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("PreviousHashes do not match\n")
	}
	pos = pos + len(b.PreviousHash)

	//check MerkleRootHash
	blockMerkleHash := serial[pos : pos+len(b.MerkleRootHash)]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("MerkleRootHashes do not match\n")
	}

}
