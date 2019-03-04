package blockchain

import (
	"bytes"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"

	block "../block"
)

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
	// Get files
	testFile := "testBlockchain.dat"
	testDB := "testDatabase.db"
	// Add the block
	err := AddBlock(b, testFile, testDB)
	if err != nil {
		t.Errorf("Failed to add block")
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
	// Blockchain datafile
	testFile := "testBlockchain.dat"
	// Create database
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	conn.Close()
	// Add the block
	err := AddBlock(expectedBlock, testFile, testDB)
	if err != nil {
		t.Errorf("Failed to add block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	actualBlock, err := GetBlockByHeight(0, testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	os.Remove(testFile)
	os.Remove(testDB)
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
	// Blockchain datafile
	testFile := "testBlockchain.dat"
	// Create database
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	conn.Close()
	// Add the block
	err := AddBlock(expectedBlock, testFile, testDB)
	if err != nil {
		t.Errorf("Failed to add block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	actualBlock, err := GetBlockByPosition(0, testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	os.Remove(testFile)
	os.Remove(testDB)
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
	// Blockchain datafile
	testFile := "testBlockchain.dat"
	// Create database
	testDB := "testDatabase.db"
	conn, _ := sql.Open("sqlite3", testDB)
	conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	conn.Close()
	// Add the block
	err := AddBlock(expectedBlock, testFile, testDB)
	if err != nil {
		t.Errorf("Failed to add block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	actualBlock, err := GetBlockByHash(block.HashSHA256(expectedBlock.Serialize()), testFile)
	if err != nil {
		t.Errorf("Failed to extract block.")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
		os.Remove(testFile)
		os.Remove(testDB)
	}
	os.Remove(testFile)
	os.Remove(testDB)
}
