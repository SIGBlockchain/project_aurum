package block

import (
	"bytes"           // for comparing []bytes
	"encoding/binary" // for encoding/decoding
	"fmt"             // printing
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
		TimeStamp:      nowTime,
		Data:           [][]byte{}, // need to add
	}
	// now use the serialize function
	serial := b.Serialize()
	// will keep track of current position in the slice
	pos := 0
	fmt.Printf("serial... %x\n", serial)

	// check version
	fmt.Printf("b.Version %x\n", b.Version)
	fmt.Printf("spot for version %d\n", binary.LittleEndian.Uint32(serial[0:4]))
	blockVersion := binary.LittleEndian.Uint32(serial[0:4])

	if blockVersion != b.Version {
		t.Errorf("version does not match")
	}

	// check height
	fmt.Printf("b.Height %d\n", b.Height)
	fmt.Printf("spot for height bytes %d\n", binary.LittleEndian.Uint32(serial[4:12]))
	blockHeight := binary.LittleEndian.Uint64(serial[4:12])
	if blockHeight != b.Height {
		t.Errorf("height does not match")
	}
	pos = 12

	// check previousHash
	fmt.Printf("b.previousHash %x\n", b.PreviousHash)
	fmt.Printf("spot for previousHash bytes %x\n", serial[12:pos+len(b.PreviousHash)])
	blockPrevHash := serial[12 : pos+len(b.PreviousHash)]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("Previous hashes do not match\n")
	}
	pos = pos + len(b.PreviousHash)

	//check merkleRootHash
	fmt.Printf("b.merkleRootHash %x\n", b.MerkleRootHash)
	fmt.Printf("spot for merkleRootHash bytes %x\n", serial[pos:pos+len(b.MerkleRootHash)])
	blockMerkleHash := serial[pos : pos+len(b.MerkleRootHash)]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("Merkle hashes do not match\n")
	}
	pos = pos + len(b.MerkleRootHash)

	//check timeStamp
	fmt.Printf("b.TimeStamp %x\n", b.TimeStamp)
	fmt.Printf("spot for timestamp bytes %x\n", binary.LittleEndian.Uint64(serial[pos:]))
	blockTimeStamp := binary.LittleEndian.Uint64(serial[pos:])
	if int64(blockTimeStamp) != b.TimeStamp {
		t.Errorf("time stamps do not match")
	}
}
