package block

import (
	"encoding/binary" // for converting to uints to byte slices
	"crypto/sha256" // for hashing
)

type Block struct {
	Version        uint32
	Height         uint64
	Timestamp      int64
	PreviousHash   []byte
	MerkleRootHash []byte
	Data           [][]byte
}

// function only serializes the block header for now
// need to add in the Data
func (b *Block) Serialize() []byte {
	// allocates space for the known variables
	serializedBlock := make([]byte, 20)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint32(serializedBlock[0:4], b.Version)
	binary.LittleEndian.PutUint64(serializedBlock[4:12], b.Height)
	binary.LittleEndian.PutUint64(serializedBlock[12:20], uint64(b.Timestamp))

	// now append the remaining information and return the complete block header byte slice
	serializedBlock = append(serializedBlock, b.PreviousHash...)
	return append(serializedBlock, b.MerkleRootHash...)
}

// function hashes data
func HashSHA256 (data []byte) ([32]byte) {
	return sha256.Sum256(data)
}