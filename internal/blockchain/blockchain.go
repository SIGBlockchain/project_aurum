// Package contains all necessary tools to interact with  and store the block chain
package blockchain

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	block "github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

type LedgerManager struct {
	file     *os.File
	database *sql.DB
	mutex    sync.RWMutex
}

func (m *LedgerManager) Lock() {
	m.mutex.Lock()
}

func (m *LedgerManager) Unlock() {
	m.mutex.Unlock()
}

func (m *LedgerManager) AddBlock(b block.Block) error {
	m.Lock()
	err := AddBlock(b, m.file, m.database)
	m.Unlock()
	return err
}

func (m *LedgerManager) GetBlockByHeight(height int) ([]byte, error) {
	return GetBlockByHeight(height, m.file, m.database)
}

func (m *LedgerManager) GetBlockByPosition(position int) ([]byte, error) {
	return GetBlockByPosition(position, m.file, m.database)
}

func (m *LedgerManager) GetBlockByHash(hash []byte) ([]byte, error) {
	return GetBlockByHash(hash, m.file, m.database)
}

func (m *LedgerManager) GetYoungestBlock() (block.Block, error) {
	return GetYoungestBlock(m.file, m.database)
}

func (m *LedgerManager) GetYoungestBlockHeader() (block.BlockHeader, error) {
	return GetYoungestBlockHeader(m.file, m.database)
}

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

	statement, err := database.Prepare(sqlstatements.INSERT_VALUES_INTO_METADATA)
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
		return nil, errors.New("Unable to read from blocks position to it's end: " + err.Error())
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
		return block.BlockHeader{}, errors.New("Failed to retreive youngest block: " + err.Error())
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

// This is a security feature for the ledger. If the metadata table gets lost somehow, this function will restore it completely.
//
// Another situation is when a producer in a decentralized system joins the network and wants the full ledger.
func RecoverBlockchainMetadata(ledgerFilename string, metadataFilename string, accountBalanceTable string) error {
	//check if metadata file exists
	err := emptyFile(metadataFilename)
	if err != nil {
		log.Printf("Failed to create an empty metadata file: %s", err.Error())
		return errors.New("Failed to create an empty metadata file")
	}
	//check if account balance file exists
	err = emptyFile(accountBalanceTable)
	if err != nil {
		log.Printf("Failed to create an empty account file: %s", err.Error())
		return errors.New("Failed to create an empty account file")
	}

	//set up the two database tables
	metaDb, err := sql.Open("sqlite3", metadataFilename)
	if err != nil {
		return errors.New("Failed to open newly created metadata db")
	}
	defer metaDb.Close()
	_, err = metaDb.Exec(sqlstatements.CREATE_METADATA_TABLE)
	if err != nil {
		return errors.New("Failed to create metadata table")
	}

	accDb, err := sql.Open("sqlite3", accountBalanceTable)
	if err != nil {
		return errors.New("Failed to open newly created accounts db")
	}
	defer accDb.Close()
	_, err = accDb.Exec(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	if err != nil {
		return errors.New("Failed to create acount_balances table")
	}

	// open ledger file
	ledgerFile, err := os.OpenFile(ledgerFilename, os.O_RDONLY, 0644)
	if err != nil {
		return errors.New("Failed to open ledger file")
	}
	defer ledgerFile.Close()

	// loop that adds blocks' metadata into database
	bPosition := int64(0)
	for {
		bOldPos := bPosition
		deserializedBlock, bLen, err := extractBlock(ledgerFile, &bPosition)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Failed to extract block from ledger: %s", err.Error())
			return errors.New("Failed to extract block from ledger")
		}

		//update the metadata table
		err = insertMetadata(metaDb, deserializedBlock, bLen, bOldPos)
		if err != nil {
			return err
		}

		//update the account table
		err = accountstable.UpdateAccountTable(accDb, deserializedBlock)
		if err != nil {
			return err
		}

	}

	return err
}

//creates an empty file if the file doesn't exist, or clears if the contents of the file if it exists
func emptyFile(fileName string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		f, err := os.Create(fileName)
		if err != nil {
			return errors.New("Failed to create " + fileName)
		}
		f.Close()
	} else { //file exits, so clear the file
		err = os.Truncate(fileName, 0)
		if err != nil {
			return errors.New("Failed to truncate " + fileName)
		}
	}
	return nil
}

//extract a block from the file, also update file position
func extractBlock(ledgerFile *os.File, pos *int64) (*block.Block, uint32, error) {
	length := make([]byte, 4)

	// read 4 bytes for blocks' length
	_, err := ledgerFile.Read(length)
	if err == io.EOF {
		return nil, 0, err
	} else if err != nil {
		return nil, 0, errors.New("Failed to read ledger file")
	}

	bLen := binary.LittleEndian.Uint32(length)

	// set offset for next read to get to the position of the block
	ledgerFile.Seek(*pos+int64(len(length)), 0)
	serialized := make([]byte, bLen)

	//update file position
	*pos += int64(len(length) + len(serialized))

	//extract block
	_, err = io.ReadAtLeast(ledgerFile, serialized, int(bLen))
	if err != nil {
		return nil, 0, errors.New("Failed to retrieve serialized block")
	}

	deserializedBlock := block.Deserialize(serialized)
	return &deserializedBlock, bLen, nil
}

/*inserts the block metadata into the metadata table
  NOTE: the db connection passed in should be open
*/
func insertMetadata(db *sql.DB, b *block.Block, bLen uint32, pos int64) error {
	bHeight := b.Height
	bHash := block.HashBlock(*b)

	sqlQuery := sqlstatements.INSERT_VALUES_INTO_METADATA
	_, err := db.Exec(sqlQuery, bHeight, pos, bLen, bHash)
	if err != nil {
		log.Printf("Failed to execute statement: %s", err.Error())
		return errors.New("Failed to execute statement")
	}

	return nil
}

func Airdrop(blockchain string, metadata string, accountBalanceTable string, genesisBlock block.Block) error {
	// create blockchain file
	file, err := os.Create(blockchain)
	if err != nil {
		return errors.New("Failed to create blockchain file: " + err.Error())
	}
	file.Close()

	// create metadata file
	file, err = os.Create(metadata)
	if err != nil {
		return errors.New("Failed to create metadata table")
	}
	file.Close()

	// open metadata file and create the table
	db, err := sql.Open("sqlite3", metadata)
	if err != nil {
		return errors.New("Failed to open table")
	}
	defer db.Close()

	_, err = db.Exec(sqlstatements.CREATE_METADATA_TABLE)
	if err != nil {
		return errors.New("Failed to create table")
	}

	// open ledger file
	ledgerFile, err := os.OpenFile(blockchain, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("Failed to open ledger file")
	}

	// add genesis block into blockchain
	err = AddBlock(genesisBlock, ledgerFile, db)
	if err != nil {
		return errors.New("Failed to add genesis block into blockchain")
	}
	ledgerFile.Close()

	// create accounts file
	file, err = os.Create(accountBalanceTable)
	if err != nil {
		return errors.New("Failed to create accounts table")
	}
	file.Close()

	accDb, err := sql.Open("sqlite3", accountBalanceTable)
	if err != nil {
		return errors.New("Failed to open newly created accounts db")
	}
	defer accDb.Close()

	_, err = accDb.Exec(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	if err != nil {
		return errors.New("Failed to create acount_balances table")
	}

	stmt, err := accDb.Prepare(sqlstatements.INSERT_VALUES_INTO_ACCOUNT_BALANCES)
	if err != nil {
		return errors.New("Failed to create statement for inserting into account table")
	}

	for _, contrcts := range genesisBlock.Data {
		var contract contracts.Contract
		contract.Deserialize(contrcts)
		_, err := stmt.Exec(hex.EncodeToString(contract.RecipPubKeyHash), contract.Value, 0)
		if err != nil {
			return errors.New("Failed to execute statement for inserting into account table")
		}
	}
	return nil
}
