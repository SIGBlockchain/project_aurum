package blockchain

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"

	block "../block"
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
	// open file for appending
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("File containing block informations failed to open")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return errors.New("Could not get file stats")
	}
	bPosition := fileInfo.Size()

	serialized := b.Serialize()
	bLen := len(serialized)
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, uint32(bLen))
	payload = append(payload, serialized...)

	if _, err := file.Write(payload); err != nil {
		fmt.Println(err)
		return errors.New("Unable to write serialized block with it's size prepended onto file")
	}

	if err := file.Close(); err != nil {
		return errors.New("Failed to close file")
	}

	database, err := sql.Open("sqlite3", databaseName)
	// Checks if the opening was successful
	if err != nil {
		return errors.New("Unable to open sqlite3 database")
	}

	defer database.Close()

	statement, err := database.Prepare("INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return errors.New("Failed to prepare a statement for further queries")
	}
	_, err = statement.Exec(b.Height, bPosition, bLen, block.HashBlock(b))
	if err != nil {
		return errors.New("Failed to execute query")
	}

	return nil
}

// Phase 2:
// Given a height number, opens the file filename and extracts the block of that height
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByHeight(height int, filename string, database string) ([]byte, error) { // Additional parameter is DB connection
	//open the file
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, errors.New("Unable to open file used to extract block from by height")
	}

	defer file.Close()

	// open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.New("Failed to open sqlite3 database")
	}

	defer db.Close()

	var blockPos int
	var blockSize int
	// only need the height, position and size of the block
	rows, err := db.Query("SELECT height, position, size FROM metadata")
	if err != nil {
		return nil, errors.New("Failed to create rows to iterate database to find height, position, and size of block")
	}
	var ht int
	var pos int
	var size int
	for rows.Next() {
		rows.Scan(&ht, &pos, &size)
		if ht == height {
			// save the wanted blocks size and position
			blockSize = size
			blockPos = pos
		}
	}

	// goes to the positition of the block
	_, err = file.Seek(int64(blockPos)+4, 0)
	if err != nil {
		return nil, errors.New("Failed to seek up to given block position in file")
	}

	// store the bytes from the file
	bl := make([]byte, blockSize)
	_, err = io.ReadAtLeast(file, bl, blockSize)
	if err != nil {
		return nil, errors.New("Unable to read from blocks position to it's end")
	}

	if err := file.Close(); err != nil {
		return nil, errors.New("Unable to close file properly")
	}

	return bl, nil
}

// Phase 2:
// Given a file position, opens the file filename and extracts the block at that position
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByPosition(position int, filename string, database string) ([]byte, error) { // Additional parameter is DB connection
	// open file
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, errors.New("Unable to open file used to extract block from by position")
	}

	defer file.Close()

	//open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.New("Unable to open sqlite3 database")
	}

	defer db.Close()

	var wantedSize int
	var wantedPos int
	// will only need the position and size of the block
	rows, err := db.Query("SELECT position, size FROM metadata")
	if err != nil {
		return nil, errors.New("Failed to create rows to iterate to find position and size of wanted block")
	}
	var pos int
	var size int
	for rows.Next() {
		rows.Scan(&pos, &size)
		if pos == position {
			// save the wanted block size and position
			wantedSize = size
			wantedPos = pos
		}
	}

	// goes to the positition of the block given through param
	_, err = file.Seek(int64(wantedPos)+4, 0)
	if err != nil {
		return nil, errors.New("Failed to seek up to given blocks position in file")
	}

	// store the bytes from the file reading from the seeked position to the size of the block
	bl := make([]byte, wantedSize)
	_, err = io.ReadAtLeast(file, bl, wantedSize)
	if err != nil {
		return nil, errors.New("Unable to read file data from the blocks start to it's end")
	}

	if err := file.Close(); err != nil {
		return nil, errors.New("Failed to close file after use")
	}

	return bl, nil
}

// Phase 2:
// Given a block hash, opens the file filename and extracts the block that matches that block's hash
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByHash(hash []byte, filename string, database string) ([]byte, error) { // Additional parameter is DB connection
	//TODO
	return []byte{}, nil
}
