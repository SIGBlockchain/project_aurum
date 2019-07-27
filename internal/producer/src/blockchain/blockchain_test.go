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
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

func setUp(filename string, database string) *sql.DB {
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		panic("Failed to open database")
	}
	statement, _ := conn.Prepare(sqlstatements.CREATE_METADATA_TABLE)
	statement.Exec()

	file, err := os.Create(filename)
	if err != nil {
		panic("Failed to create file.")
	}
	file.Close()

	return conn
}

func tearDown(metadata *sql.DB, filename string, database string) {
	metadata.Close()
	os.Remove(filename)
	os.Remove(database)
}

func addBlockHelper(b block.Block, filename string, metadata *sql.DB) error {
	f, _ := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	defer f.Close()
	return AddBlock(b, f, metadata)
}
func TestPhaseOneAddBlock(t *testing.T) {

	// Create a block
	b := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	b.DataLen = uint16(len(b.Data))

	// Setup
	metadata := setUp("testFile.txt", "testDatabase.db")
	defer tearDown(metadata, "testFile.txt", "testDatabase.db")

	err := addBlockHelper(b, "testFile.txt", metadata)
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
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))

	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block. " + err.Error())
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByHeight(0, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block." + err.Error())
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
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByPosition(0, file, metadata)
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
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByHash(block.HashBlock(expectedBlock), file, metadata)
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
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x', 'o', 'x', 'o'})},
	}
	block0.DataLen = uint16(len(block0.Data))
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block0),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x', 'y', 'z'})},
	}
	block1.DataLen = uint16(len(block1.Data))
	block2 := block.Block{
		Version:        1,
		Height:         2,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block1),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'a', 'b', 'c'})},
	}
	block2.DataLen = uint16(len(block2.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add all the blocks
	err := addBlockHelper(block0, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block0.")
	}
	err = addBlockHelper(block1, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block1.")
	}
	err = addBlockHelper(block2, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block2.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	// Extract all three blocks
	// Block 0 by hash
	actualBlock0, err := GetBlockByHash(block.HashBlock(block0), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by hash).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by hash)")
	}

	// Block 0 by height
	actualBlock0, err = GetBlockByHeight(0, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by height).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by height)")
	}

	// Block 1 by hash
	actualBlock1, err := GetBlockByHash(block.HashBlock(block1), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by hash).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by hash)")
	}

	// Block 1 by height
	actualBlock1, err = GetBlockByHeight(1, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by height).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by height)")
	}

	// Block 2
	actualBlock2, err := GetBlockByHash(block.HashBlock(block2), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 2 by hash).")
	}
	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
		t.Errorf("Blocks do not match (block 2 by hash)")
	}

	// Block 2
	actualBlock2, err = GetBlockByHeight(2, file, metadata)
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
	metadata := setUp(blockchain, table)
	defer tearDown(metadata, blockchain, table)

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	_, err = GetYoungestBlock(file, metadata)
	if err == nil {
		t.Errorf("Should return error if blockchain is empty")
	}
	file.Close()
	block0 := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte("xoxo"))},
	}
	block0.DataLen = uint16(len(block0.Data))
	err = addBlockHelper(block0, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	file, err = os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	actualBlock0, err := GetYoungestBlock(file, metadata)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock0, block0) {
		t.Errorf("Blocks do not match")
	}
	file.Close()
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte("xoxo"))},
	}
	block1.DataLen = uint16(len(block1.Data))
	block1Header := block.BlockHeader{
		Version:        1,
		Height:         1,
		Timestamp:      block1.Timestamp,
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
	}
	err = addBlockHelper(block1, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	file, err = os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	actualBlock1Header, err := GetYoungestBlockHeader(file, metadata)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock1Header, block1Header) {
		t.Errorf("Blocks Headers do not match")
	}
	file.Close()
}
