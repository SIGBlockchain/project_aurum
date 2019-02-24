package block

type Block struct {
	Version        uint32
	Height         uint64
	PreviousHash   []byte
	MerkleRootHash []byte
	TimeStamp      int64
	Data           [][]byte
}
