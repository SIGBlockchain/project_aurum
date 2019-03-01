package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"
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
		Data:           [][]byte{{12, 3}, {132, 90, 23}, {23}}, // need to add
	}
	// now use the serialize function
	serial := b.Serialize()
	fmt.Println(serial)
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
	//check Data
	testslice := [][]byte{{12, 3}, {132, 90, 23}, {23}}
	blockData := serial[84:90]
	counter := 0
	for i := 0; i < len(testslice); i++ {
		for j := 0; j < len(testslice[i]); j++ {
			if testslice[i][j] != blockData[counter] {
				t.Errorf("Data do not match!")
			}
			counter++
		}
	}

}
