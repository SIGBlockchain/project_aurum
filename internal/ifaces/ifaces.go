package ifaces

import (
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

type IBlockFetcher interface {
	FetchBlockByHeight(uint64) ([]byte, error)
}

// BlockchainScanner allows for implementation of a function that returns
// collection of block objects
type BlockchainScanner interface {
	ScanBlockChain(senderWalletAddress publickey.AurumPublicKey) ([]block.Block, error)
}
