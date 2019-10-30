package ifaces

import "github.com/SIGBlockchain/project_aurum/internal/block"

type IBlockFetcher interface {
	FetchBlockByHeight(uint64) ([]byte, error)
}

// IBlockchainStreamer returns the number of blocks that are relevant from the block slice
type IBlockchainStreamer interface {
	Stream([]block.Block) (int, error)
}

type ILedgerManager interface {
	AddBlock(b block.Block) error
	GetBlockByHeight(height int) ([]byte, error)
	GetBlockByPosition(position int) ([]byte, error)
	GetBlockByHash(hash []byte) ([]byte, error)
	GetYoungestBlock() (block.Block, error)
	GetYoungestBlockHeader() (block.BlockHeader, error)
	Airdrop(genesisBlock block.Block) error
	Lock()
	Unlock()
}
