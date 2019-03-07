package blockchain

import (
	"errors"

	block "project_aurum/producer/src/block"
)

// Phase 1:
// Open file for appending
// If open fails return and error
// Serializes the block
// Gets size of the block string
// Prepends the size of the block to the serialized block
// Takes the resulting concatenation and append it to file named filename
// If the write fails, return an error
// Close the file

// Phase 2:
// Open the database
// Store the height, file position, size, and hash of the block header into a database row
// If this fails, return an error
// Close the database connection
func AddBlock(b block.Block, filename string, databaseName string) error { // Additional parameter is DB connection
	// TODO
	return errors.New("") // change this to nil once the function is completed
}

// Phase 2:
// Given a height number, opens the file filename and extracts the block of that height
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByHeight(height int, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}

// Phase 2:
// Given a file position, opens the file filename and extracts the block at that position
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByPosition(position int, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}

// Phase 2:
// Given a block hash, opens the file filename and extracts the block that matches that block's hash
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByHash(hash []byte, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}
