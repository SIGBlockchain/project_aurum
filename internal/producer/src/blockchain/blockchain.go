// Package contains all necessary tools to interact with  and store the block chain
package blockchain

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"

	block "github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
)

// Adds a block to a given file, also adds metadata file about that block into a database
//
// This metadata include height, position, size and hash
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

// Given a height number, opens the file filename and extracts the block of that height
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

// Given a file position, opens the file filename and extracts the block at that position
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

// Given a block hash, opens the file filename and extracts the block that matches that block's hash
func GetBlockByHash(hash []byte, filename string, database string) ([]byte, error) { // Additional parameter is DB connection
	// open file
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, errors.New("Unable to open file used to extract block from by position")
	}

	//open database
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.New("Unable to open sqlite3 database")
	}

	defer file.Close()
	defer db.Close()

	var blockPos int
	var blockSize int
	// need the position, size and hash of the block from databse
	rows, err := db.Query("SELECT position, size, hash FROM metadata")
	if err != nil {
		return nil, errors.New("Failed to create rows to iterate to find position and size of wanted block")
	}
	var pos int
	var size int
	var bHash string
	for rows.Next() {
		rows.Scan(&pos, &size, &bHash)
		if bHash == string(hash) {
			// save the wanted block size and position
			blockPos = pos
			blockSize = size
		}
	}

	// goes to the positition of the block given through param
	_, err = file.Seek(int64(blockPos)+4, 0)
	if err != nil {
		return nil, errors.New("Failed to seek up to given blocks position in file")
	}

	// store the bytes from the file reading from the seeked position to the size of the block
	bl := make([]byte, blockSize)
	_, err = io.ReadAtLeast(file, bl, blockSize)
	if err != nil {
		return nil, errors.New("Unable to read file data from the blocks start to it's end")
	}

	if err := file.Close(); err != nil {
		return nil, errors.New("Failed to close file after use")
	}

	return bl, nil
}

// This is a security feature for the ledger. If the block table gets lost somehow, this function will restore it completely. 
//
// Another situation is when a producer in a decentralized system joins the network and wants the full ledger.
func RecoverBlockchainMetadata(ledgerFilename string, metadataFilename string) error {
	_, err := os.Stat(metadataFilename)
	if err != nil {
		// create database
		f, err := os.Create(metadataFilename)
		f.Close()

		// open database
		db, err := sql.Open("sqlite3", metadataFilename)
		if err != nil {
			return errors.New("Failed to open newly created database")
		}
		defer db.Close()

		// create metadata table in database
		statement, _ := db.Prepare("CREATE TABLE metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
		statement.Exec()
		statement.Close()

		// create a prepared statement to insert into metadata
		statement, err = db.Prepare("INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)")
		if err != nil {
			return errors.New("Failed to prepare a statement for further executes")
		}
		defer statement.Close()

		// open ledger file
		file, err := os.OpenFile(ledgerFilename, os.O_RDONLY, 0644)
		if err != nil {
			return errors.New("Failed to open ledger file")
		}
		defer file.Close()

		// loop that adds blocks' metadata into database
		bPosition := int64(0)
		for {
			length := make([]byte, 4)

			// read 4 bytes for blocks' length
			_, err = file.Read(length)
			if err == io.EOF {
				// if reader reaches EOF, exit loop
				break
			} else if err != nil {
				return errors.New("Failed to read ledger file")
			}
			bLen := binary.LittleEndian.Uint32(length)

			// set offset for next read to get to the position of the block
			file.Seek(bPosition+int64(len(length)), 0)
			serialized := make([]byte, bLen)
			_, err = io.ReadAtLeast(file, serialized, int(bLen))
			if err != nil {
				return errors.New("Failed to retrieve serialized block")
			}

			// need to deserialize block for block's height and hash
			deserializedBlock := block.Deserialize(serialized)
			bHeight := deserializedBlock.Height
			bHash := block.HashBlock(deserializedBlock)

			// execute statement
			_, err = statement.Exec(bHeight, bPosition, bLen, bHash)
			if err != nil {
				return errors.New("Failed to execute statement")
			}

			// position of next block
			bPosition += int64(len(length) + len(serialized))
		}
	}
	return err
}

/*
Retrieves Block with the largest height in deserialized form
*/
func GetYoungestBlock(blockchain string, table string) (block.Block, error) {
	// open table
	db, err := sql.Open("sqlite3", table)
	if err != nil {
		return block.Block{}, errors.New("Failed to open table")
	}
	defer db.Close()

	// create rows to find blocks' height from metadata
	rows, err := db.Query("SELECT height FROM metadata")
	if err != nil {
		return block.Block{}, errors.New("Failed to create rows to find height from metadata")
	}
	defer rows.Close()

	if !rows.Next() {
		// if there are no rows in the table, return error
		return block.Block{}, errors.New("Empty blockchain")
	}

	// find the largest height in the table
	var maxBlockHeight int
	rows.Scan(&maxBlockHeight)
	var blockHeight int
	for rows.Next() {
		rows.Scan(&blockHeight)
		if blockHeight > maxBlockHeight {
			maxBlockHeight = blockHeight
		}
	}

	// get the block with the largest height
	youngestBlock, err := GetBlockByHeight(maxBlockHeight, blockchain, table)
	if err != nil {
		return block.Block{}, err
	}
	return block.Deserialize(youngestBlock), nil
}

/*
Calls GetYoungestBlock and returns a Header version of the result
*/
func GetYoungestBlockHeader(blockchain string, table string) (block.BlockHeader, error) {
	latestBlock, err := GetYoungestBlock(blockchain, table)
	if err != nil {
		return block.BlockHeader{}, errors.New("Failed to retreive youngest block")
	}

	latestBlockHeader := block.BlockHeader{
		Version:        latestBlock.Version,
		Height:         latestBlock.Height,
		Timestamp:      latestBlock.Timestamp,
		PreviousHash:   latestBlock.PreviousHash,
		MerkleRootHash: latestBlock.MerkleRootHash,
	}
	return latestBlockHeader, nil
}
