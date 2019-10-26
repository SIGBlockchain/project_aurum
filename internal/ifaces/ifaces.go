package ifaces

import "github.com/SIGBlockchain/project_aurum/internal/block"

type IBlockFetcher interface {
	FetchBlockByHeight(uint64) ([]byte, error)
}

// IBlockchainStreamer returns the number of blocks that are relevant from the block slice
type IBlockchainStreamer interface {
	Stream([]block.Block) (int, error)
}
