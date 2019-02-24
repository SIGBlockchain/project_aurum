package block

import (
	"encoding/binary" // for converting to uints to byte slices
)

type Block struct {
	Version        uint32 //32/8 = 4 bytes
	Height         uint64 // 64/8 = 8 bytes
	PreviousHash   []byte // 32bytes 255 bits
	MerkleRootHash []byte // 32bytes
	TimeStamp      int64
	Data           [][]byte
}

func (b *Block) Serialize() []byte {
	// slice that will only contain the Version and Height (size = 12)
	slice1 := make([]byte, 12)
	// Version and Height to bytes in little endian and inside newSlice
	binary.LittleEndian.PutUint32(slice1[0:4], b.Version)
	binary.LittleEndian.PutUint64(slice1[4:12], b.Height)

	// slice that will contain both hashes that are already in byte slices with remaining zeros of previous hash then merkleRoot
	//hashSlice := make([]byte, 64)
	hashSlice1 := make([]byte, 32-len(b.PreviousHash))
	hashSlice1 = append(b.PreviousHash, hashSlice1...)
	//fmt.Printf("hashSlice1 %x\n", hashSlice1)

	hashSlice2 := make([]byte, 32-len(b.MerkleRootHash))
	hashSlice2 = append(b.MerkleRootHash, hashSlice2...)
	//	fmt.Printf("hashSlice2 %x\n", hashSlice2)

	// both hashes together
	//hashSlice = append(b.PreviousHash, b.MerkleRootHash...)
	hashSlices := append(hashSlice1, hashSlice2...)
	//	fmt.Print(b.PreviousHash)
	//	fmt.Println(b.MerkleRootHash)
	//	fmt.Printf("hashSlice %d\n", hashSlices)

	// concat what we currently have which will be of size 76 bytes
	// Version, Height, PreviousHash and MerkleRootHash
	//slice1Nhash := make([]byte, 76)
	slice1Nhashes := append(slice1, hashSlices...)

	actualSliceNhashes := make([]byte, 76-len(slice1Nhashes))
	actualSliceNhashes = append(slice1Nhashes, actualSliceNhashes...)

	//TimeStamp to bytes in little endian and inside another slice
	timeSlice := make([]byte, 8)
	//binary.LittleEndian.PutUint64(timeSlice, uint64(b.TimeStamp))
	binary.LittleEndian.PutUint64(timeSlice[0:8], uint64(b.TimeStamp))

	// now get to the data
	var dataSlice []byte // will store data from Data
	// for each sub slice in Data
	for i := 0; i < len(b.Data); i++ {
		//fmt.Println(b.Data[i])
		//fmt.Printf("len of b.Data[i] %d val of start %d\n", len(b.Data[i]), start)
		//fmt.Printf("lower lim %d upper lim %d\n", start, start+len(b.Data[i]))
		for x := 0; x < len(b.Data[i]); x++ {
			dataSlice = append(dataSlice, []byte{b.Data[i][x]}...)
		}
	}
	//fmt.Println(dataSlice)

	//slice1HashTimeSlice := make([]byte, 84)
	//slice1HashTimeSlice = append(slice1Nhashes, timeSlice...)
	slice1NhashesTimeSlice := append(slice1Nhashes, timeSlice...)
	actualSliceHashTime := make([]byte, 84-len(slice1NhashesTimeSlice))
	actualSliceHashTime = append(slice1NhashesTimeSlice, actualSliceHashTime...)

	// now concatenate all
	var allData []byte
	//allData = append(slice1HashTimeSlice, dataSlice...)
	allData = append(actualSliceHashTime, dataSlice...)
	return allData
}
