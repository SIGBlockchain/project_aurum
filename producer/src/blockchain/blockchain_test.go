package blockchain

import (
	"reflect"
	"testing"
	"time"

	block "../block"
)

func TestAddBlockPhaseOne(t *testing.T) {
	// Create a block
	b := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	testFile := "testBlockchain.dat"
	// Add the block
	err := AddBlock(b, testFile)
	if err != nil {
		t.Errorf("Failed to add block")
	}
}

func TestGetBlockByHeight(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	testFile := "testBlockchain.dat"
	// Add the block
	err := AddBlock(expectedBlock, testFile)
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByHeight(0, testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if reflect.DeepEqual(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}
func TestGetBlockPosition(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	testFile := "testBlockchain.dat"
	// Add the block
	err := AddBlock(expectedBlock, testFile)
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByPosition(0, testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if reflect.DeepEqual(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}
func TestGetBlockByHash(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	testFile := "testBlockchain.dat"
	// Add the block
	err := AddBlock(expectedBlock, testFile)
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByHash(block.HashSHA256(expectedBlock.Serialize()), testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if reflect.DeepEqual(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}
