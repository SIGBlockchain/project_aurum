package blockchain

import (
	"bytes"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"

	block "github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
)

func setUp(filename string, database string) {
	conn, _ := sql.Open("sqlite3", database)
	statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	statement.Exec()
	conn.Close()

	file, err := os.Create(filename)
	if err != nil {
		panic("Failed to create file.")
	}
	file.Close()
}

func tearDown(filename string, database string) {
	os.Remove(filename)
	os.Remove(database)
}

func TestPhaseOneAddBlock(t *testing.T) {

	// Create a block
	b := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	b.DataLen = uint16(len(b.Data))

	// Setup
	defer tearDown("testFile.txt", "testDatabase.db")
	setUp("testFile.txt", "testDatabase.db")

	// Add the block
	err := AddBlock(b, "testFile.txt", "testDatabase.db")
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestPhaseTwoGetBlockByHeight(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))

	// Setup
	setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown("testBlockchain.dat", "testDatabase.db")
	// Add the block
	err := AddBlock(expectedBlock, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByHeight(0, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}
func TestPhaseTwoGetBlockPosition(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown("testBlockchain.dat", "testDatabase.db")
	// Add the block
	err := AddBlock(expectedBlock, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByPosition(0, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}

func TestPhaseTwoGetBlockByHash(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown("testBlockchain.dat", "testDatabase.db")
	// Add the block
	err := AddBlock(expectedBlock, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block.")
	}
	actualBlock, err := GetBlockByHash(block.HashBlock(expectedBlock), "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}

func TestPhaseTwoMultiple(t *testing.T) {
	// Create a bunch of blocks
	block0 := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x', 'o', 'x', 'o'})},
	}
	block0.DataLen = uint16(len(block0.Data))
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block0),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'x', 'y', 'z'})},
	}
	block1.DataLen = uint16(len(block1.Data))
	block2 := block.Block{
		Version:        1,
		Height:         2,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block1),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte{'a', 'b', 'c'})},
	}
	block2.DataLen = uint16(len(block2.Data))
	// Setup
	setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown("testBlockchain.dat", "testDatabase.db")
	// Add all the blocks
	err := AddBlock(block0, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block0.")
	}
	err = AddBlock(block1, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block1.")
	}
	err = AddBlock(block2, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to add block2.")
	}

	// Extract all three blocks
	// Block 0 by hash
	actualBlock0, err := GetBlockByHash(block.HashBlock(block0), "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by hash).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by hash)")
	}

	// Block 0 by height
	actualBlock0, err = GetBlockByHeight(0, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by height).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by height)")
	}

	// Block 1 by hash
	actualBlock1, err := GetBlockByHash(block.HashBlock(block1), "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by hash).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by hash)")
	}

	// Block 1 by height
	actualBlock1, err = GetBlockByHeight(1, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by height).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by height)")
	}

	// Block 2
	actualBlock2, err := GetBlockByHash(block.HashBlock(block2), "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 2 by hash).")
	}
	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
		t.Errorf("Blocks do not match (block 2 by hash)")
	}

	// Block 2
	actualBlock2, err = GetBlockByHeight(2, "testBlockchain.dat", "testDatabase.db")
	if err != nil {
		t.Errorf("Failed to extract block (block 2 by height).")
	}
	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
		t.Errorf("Blocks do not match (block 2 by height)")
	}
}

func TestGetYoungestBlockAndBlockHeader(t *testing.T) {
	blockchain := "testBlockchain.dat"
	table := "testTable.dat"
	setUp(blockchain, table)
	defer tearDown(blockchain, table)
	_, err := GetYoungestBlock(blockchain, table)
	if err == nil {
		t.Errorf("Should return error if blockchain is empty")
	}
	block0 := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte("xoxo"))},
	}
	block0.DataLen = uint16(len(block0.Data))
	err = AddBlock(block0, blockchain, table)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	actualBlock0, err := GetYoungestBlock(blockchain, table)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock0, block0) {
		t.Errorf("Blocks do not match")
	}
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
		Data:           [][]byte{block.HashSHA256([]byte("xoxo"))},
	}
	block1.DataLen = uint16(len(block1.Data))
	block1Header := block.BlockHeader{
		Version:        1,
		Height:         1,
		Timestamp:      block1.Timestamp,
		PreviousHash:   block.HashSHA256([]byte{'0'}),
		MerkleRootHash: block.HashSHA256([]byte{'1'}),
	}
	err = AddBlock(block1, blockchain, table)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	actualBlock1Header, err := GetYoungestBlockHeader(blockchain, table)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock1Header, block1Header) {
		t.Errorf("Blocks Headers do not match")
	}
}
