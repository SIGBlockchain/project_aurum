package blockchain

import (
	"database/sql"
	"errors"
	"os"

	block "github.com/SIGBlockchain/project_aurum/internal/block"
)

type BlockchainReader interface {
	GetBlockByHeight(height uint) (block.Block, error)
}

//DbFileReader is a BlockchainReader that uses a database connection and the blockchain file
type DbFileReader struct {
	file     *os.File
	database *sql.DB
}

func (r DbFileReader) GetBlockByHeight(height uint) (block.Block, error) {
	blockBytes, err := GetBlockByHeight(int(height), r.file, r.database)
	if err != nil {
		return block.Block{}, errors.New("Failed to Get block by height: " + err.Error())
	}

	block := block.Deserialize(blockBytes)
	return block, nil
}
