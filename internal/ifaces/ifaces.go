package ifaces

import (
	"github.com/SIGBlockchain/project_aurum/internal/accountinfo"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
)

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
	Lock()
	Unlock()
}

// IAccountsTableConnection is used for mocking sql.DB functions
type IAccountsTableConnection interface {
	InsertAccountIntoAccountBalanceTable(pkhash []byte, value uint64) error
	ExchangeAndUpdateAccounts(c *contracts.Contract) error
	MintAurumUpdateAccountBalanceTable(pkhash []byte, value uint64) error
	GetBalance(pkhash []byte) (uint64, error)
	GetStateNonce(pkhash []byte) (uint64, error)
	GetAccountInfo(pkhash []byte) (*accountinfo.AccountInfo, error)
	UpdateAccountTable(b *block.Block) error
	Lock()
	Unlock()
}
