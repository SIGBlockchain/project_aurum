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
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

// Adds a block to a given file, also adds metadata file about that block into a database
//
// This metadata include height, position, size and hash
func AddBlock(b block.Block, file *os.File, database *sql.DB) error {
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

	statement, err := database.Prepare(sqlstatements.INSERT_BLANK_VALUES_INTO_METADATA)
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

// Given a height number and extracts the block of that height
func GetBlockByHeight(height int, file *os.File, db *sql.DB) ([]byte, error) {
	file.Seek(0, io.SeekStart) // reset seek pointer

	var blockPos int
	var blockSize int
	// only need the height, position and size of the block
	rows, err := db.Query(sqlstatements.GET_HEIGHT_POSITION_SIZE_FROM_METADATA)
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

	return bl, nil
}

// Given a file position and extracts the block at that position
func GetBlockByPosition(position int, file *os.File, db *sql.DB) ([]byte, error) {
	file.Seek(0, io.SeekStart) // reset seek pointer

	var wantedSize int
	var wantedPos int
	// will only need the position and size of the block
	rows, err := db.Query(sqlstatements.GET_POSITION_SIZE_FROM_METADATA)
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

	return bl, nil
}

// Given a block hash and extracts the block that matches that block's hash
func GetBlockByHash(hash []byte, file *os.File, db *sql.DB) ([]byte, error) {
	file.Seek(0, io.SeekStart) // reset seek pointer

	var blockPos int
	var blockSize int
	// need the position, size and hash of the block from databse
	rows, err := db.Query(sqlstatements.GET_POSITION_SIZE_HASH_FROM_METADATA)
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

	return bl, nil
}

/*
Retrieves Block with the largest height in deserialized form
*/
func GetYoungestBlock(file *os.File, db *sql.DB) (block.Block, error) {
	// create rows to find blocks' height from metadata
	rows, err := db.Query(sqlstatements.GET_HEIGHT_FROM_METADATA)
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
	youngestBlock, err := GetBlockByHeight(maxBlockHeight, file, db)
	if err != nil {
		return block.Block{}, err
	}
	return block.Deserialize(youngestBlock), nil
}

/*
Calls GetYoungestBlock and returns a Header version of the result
*/
func GetYoungestBlockHeader(file *os.File, metadata *sql.DB) (block.BlockHeader, error) {
	latestBlock, err := GetYoungestBlock(file, metadata)
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
