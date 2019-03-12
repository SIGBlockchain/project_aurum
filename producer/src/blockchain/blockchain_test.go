package blockchain

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"

	block "../block"
)

func setUp(filename string, database string) {
	fmt.Println("Setting up test")
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
	fmt.Println("Tearing down test")
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

// func TestPhaseTwoGetBlockByHeight(t *testing.T) {
// 	// Create a block
// 	expectedBlock := block.Block{
// 		Version:        1,
// 		Height:         0,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashSHA256([]byte{'0'}),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
// 	}
// 	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
// 	// Blockchain datafile
// 	testFile := "testBlockchain.dat"
// 	// Create database
// 	testDB := "testDatabase.db"
// 	conn, _ := sql.Open("sqlite3", testDB)
// 	statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
// 	statement.Exec()
// 	conn.Close()
// 	// Add the block
// 	err := AddBlock(expectedBlock, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	actualBlock, err := GetBlockByHeight(0, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
// 		t.Errorf("Blocks do not match")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	os.Remove(testFile)
// 	os.Remove(testDB)
// }
// func TestPhaseTwoGetBlockPosition(t *testing.T) {
// 	// Create a block
// 	expectedBlock := block.Block{
// 		Version:        1,
// 		Height:         0,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashSHA256([]byte{'0'}),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
// 	}
// 	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
// 	// Blockchain datafile
// 	testFile := "testBlockchain.dat"
// 	// Create database
// 	testDB := "testDatabase.db"
// 	conn, _ := sql.Open("sqlite3", testDB)
// 	statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
// 	statement.Exec()
// 	conn.Close()
// 	// Add the block
// 	err := AddBlock(expectedBlock, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	actualBlock, err := GetBlockByPosition(0, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
// 		t.Errorf("Blocks do not match")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	os.Remove(testFile)
// 	os.Remove(testDB)
// }
// func TestPhaseTwoGetBlockByHash(t *testing.T) {
// 	// Create a block
// 	expectedBlock := block.Block{
// 		Version:        1,
// 		Height:         0,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashSHA256([]byte{'0'}),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x'})},
// 	}
// 	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
// 	// Blockchain datafile
// 	testFile := "testBlockchain.dat"
// 	// Create database
// 	testDB := "testDatabase.db"
// 	conn, _ := sql.Open("sqlite3", testDB)
// 	statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
// 	statement.Exec()
// 	conn.Close()
// 	// Add the block
// 	err := AddBlock(expectedBlock, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	actualBlock, err := GetBlockByHash(block.HashBlock(expectedBlock), testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
// 		t.Errorf("Blocks do not match")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	os.Remove(testFile)
// 	os.Remove(testDB)
// }

// func TestPhaseTwoMultiple(t *testing.T) {
// 	// Create a bunch of blocks
// 	block0 := block.Block{
// 		Version:        1,
// 		Height:         0,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashSHA256([]byte{'0'}),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x', 'o', 'x', 'o'})},
// 	}
// 	block0.DataLen = uint16(len(block0.Data))
// 	block1 := block.Block{
// 		Version:        1,
// 		Height:         1,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashBlock(block0),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x', 'y', 'z'})},
// 	}
// 	block1.DataLen = uint16(len(block1.Data))
// 	block2 := block.Block{
// 		Version:        1,
// 		Height:         2,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashBlock(block1),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'a', 'b', 'c'})},
// 	}
// 	block2.DataLen = uint16(len(block2.Data))
// 	// Blockchain datafile
// 	testFile := "testBlockchain.dat"
// 	// Create database
// 	testDB := "testDatabase.db"
// 	conn, _ := sql.Open("sqlite3", testDB)
// 	statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
// 	statement.Exec()
// 	conn.Close()
// 	// Add all the blocks
// 	err := AddBlock(block0, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block0.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	err = AddBlock(block1, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block1.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	err = AddBlock(block2, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to add block2.")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Extract all three blocks
// 	// Block 0 by hash
// 	actualBlock0, err := GetBlockByHash(block.HashBlock(block0), testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 0 by hash).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
// 		t.Errorf("Blocks do not match (block 0 by hash)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Block 0 by height
// 	actualBlock0, err = GetBlockByHeight(0, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 0 by height).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
// 		t.Errorf("Blocks do not match (block 0 by height)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Block 1
// 	actualBlock1, err := GetBlockByHash(block.HashBlock(block1), testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 1 by hash).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
// 		t.Errorf("Blocks do not match (block 1 by hash)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Block 1
// 	actualBlock1, err = GetBlockByHeight(1, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 1 by height).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
// 		t.Errorf("Blocks do not match (block 1 by height)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Block 2
// 	actualBlock2, err := GetBlockByHash(block.HashBlock(block2), testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 2 by hash).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
// 		t.Errorf("Blocks do not match (block 2 by hash)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Block 2
// 	actualBlock2, err = GetBlockByHeight(2, testFile, testDB)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 2 by height).")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}
// 	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
// 		t.Errorf("Blocks do not match (block 2 by height)")
// 		os.Remove(testFile)
// 		os.Remove(testDB)
// 	}

// 	// Remove the files
// 	os.Remove(testFile)
// 	os.Remove(testDB)
// }
