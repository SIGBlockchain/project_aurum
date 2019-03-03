package blockchain

import block "../block"

// Phase 1:
// Serializes the block
// Gets size of the block string
// Prepends the size of the block to the serialized block
// Takes the resulting concatenation and writes it to file named filename
// If the write fails, return an error

// Phase 2:
// Store the height, file position, size, and hash of the block into a database row
// If this fails, return an error
func AddBlock(b block.Block, filename string) error { // Additional parameter is DB connection
	// TODO
	return nil
}

// Phase 2:
// Given a height number, opens the file filename and extracts the block of that height
// Use a database query to find block's position and size in the file
// If this fails, return an error
func GetBlockByHeight(height int, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}

// Phase 2:
// Given a file position, opens the file filename and extracts the block at that position
// Use a database query to find block's position and size in the file
// If this fails, return an error
func GetBlockByPosition(position int, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}

// Phase 2:
// Given a block hash, opens the file filename and extracts the block that matches that block's hash
// Use a database query to find block's position and size in the file
// If this fails, return an error
func GetBlockByHash(hash []byte, filename string) ([]byte, error) { // Additional parameter is DB connection
	// TODO
	return []byte{}, nil
}
