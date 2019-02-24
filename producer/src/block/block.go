package block

import "encoding/binary" // for converting to uints to byte slices

type Block struct {
	Version        uint32
	Height         uint64
	PreviousHash   []byte
	MerkleRootHash []byte
	TimeStamp      int64
	Data           [][]byte
}

// function only serializes the block header for now
// need to add in the Data
func (b *Block) Serialize() []byte {
	// slice that will only contain the Version and Height (size = 12)
	slice1 := make([]byte, 12)
	// Version and Height to bytes in little endian and inside slice1
	binary.LittleEndian.PutUint32(slice1[0:4], b.Version)
	binary.LittleEndian.PutUint64(slice1[4:12], b.Height)

	// both hashes together
	hashSlices := append(b.PreviousHash, b.MerkleRootHash...)

	// concat what is available so far Version, Height, PreviousHash and MerkleRootHash
	slice1Nhashes := append(slice1, hashSlices...)

	//TimeStamp to bytes in little endian and inside ist own slice
	timeSlice := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeSlice[0:8], uint64(b.TimeStamp))

	// now concatenate all of the block header pieces
	slice1NhashesTimeSlice := append(slice1Nhashes, timeSlice...)

	return slice1NhashesTimeSlice
}
