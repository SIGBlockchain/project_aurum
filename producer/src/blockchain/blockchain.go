package blockchain

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	bPosition := fileInfo.Size()

	serialized := b.Serialize()
	bLen := len(serialized)
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, uint32(bLen))
	payload = append(payload, serialized...)

	if _, err := file.Write(payload); err != nil {
		fmt.Println(err)
		return err
	}

	if err := file.Close(); err != nil {
		log.Fatalln(err)
	}

	database, err := sql.Open("sqlite3", databaseName)
	// Checks if the opening was successful
	if err != nil {
		return err
	}

	statement, err := database.Prepare("INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = statement.Exec(b.Height, bPosition, bLen, block.HashBlock(b))
	if err != nil {
		return err
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
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	// open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}

	var blockPos int
	var blockSize int
	// only need the height, position and size of the block
	rows, err := db.Query("SELECT height, position, size FROM metadata")
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
	_, _ = file.Seek(int64(blockPos)+4, 0)

	// store the bytes from the file
	bl := make([]byte, blockSize)
	_, _ = io.ReadAtLeast(file, bl, blockSize)

	return bl, nil
}

// Phase 2:
// Given a file position, opens the file filename and extracts the block at that position
// Use a database query to find block's position and size in the file
// If this fails, return an error
// Make sure to close file and database connection before returning
func GetBlockByPosition(position int, filename string, database string) ([]byte, error) { // Additional parameter is DB connection
	// open file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	//open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}

	var wantedSize int
	var wantedPos int
	// will only need the position and size of the block
	rows, err := db.Query("SELECT position, size FROM metadata")
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
	_, _ = file.Seek(int64(wantedPos)+4, 0)

	// store the bytes from the file reading from the seeked position to the size of the block
	bl := make([]byte, wantedSize)
	_, _ = io.ReadAtLeast(file, bl, wantedSize)

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
