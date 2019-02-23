package block

//data a block contains
type Block struct {
	version        uint32 // 4 bytes
	height         uint64 // 8 bytes
	previousHash   []byte // 32bytes
	merkleRootHash []byte
	timeStamp      uint64
	data           [][]byte // byte size tbd
}
