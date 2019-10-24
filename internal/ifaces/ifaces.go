package ifaces

type IBlockFetcher interface {
	FetchBlockByHeight(uint64) ([]byte, error)
}
