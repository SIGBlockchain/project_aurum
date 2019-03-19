package block

import (
	"container/list"
	"crypto/sha256"   // for hashing
	"encoding/binary" // for converting to uints to byte slices
)

type Block struct {
	Version        uint32
	Height         uint64
	Timestamp      int64
	PreviousHash   []byte
	MerkleRootHash []byte
	DataLen        uint16
	Data           [][]byte
}

// Produces a block based on the struct provided
func (b *Block) Serialize() []byte { // Vineet
	//calculate the total length beforehand, to prevent unneccessary appends
	//NOTE: 32 bit ints are used to hold lengths; unsigned 16 bit int is used for the length of Data
	bLen := 86 //size of all fixed size fields
	for _, s := range b.Data {
		bLen += 2 + len(s) //2 bytes for length plus the length of an element in Data
	}
	serializedBlock := make([]byte, bLen)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint32(serializedBlock[0:4], b.Version)
	binary.LittleEndian.PutUint64(serializedBlock[4:12], b.Height)
	binary.LittleEndian.PutUint64(serializedBlock[12:20], uint64(b.Timestamp))
	copy(serializedBlock[20:52], b.PreviousHash)
	copy(serializedBlock[52:84], b.MerkleRootHash)
	binary.LittleEndian.PutUint16(serializedBlock[84:86], b.DataLen)

	i := 86
	for _, s := range b.Data {
		//for every data entry, put the legth, and then the data
		binary.LittleEndian.PutUint16(serializedBlock[i:i+2], uint16(len(s)))
		i += 2
		copy(serializedBlock[i:i+len(s)], s)
		i += len(s)
	}
	return serializedBlock
}

// function hashes data
func HashSHA256(data []byte) []byte {
	result := sha256.Sum256(data)
	return result[:]
}

// Returns the merkle root hash of the list of inputs
func GetMerkleRootHash(input [][]byte) []byte {
	if len(input) == 0 {
		return []byte{} //return an empty slice
	}
	//first add all the slices to a list
	l := list.New()
	for _, s := range input {
		//while pushing elements to the list, double hash them
		l.PushBack(HashSHA256(HashSHA256(s)))
	}
	return getMerkleRoot(l)
}

// recursive helper function
func getMerkleRoot(l *list.List) []byte {
	if l.Len() == 1 {
		return l.Front().Value.([]byte)
	}
	if l.Len()%2 != 0 { //list is of odd length
		l.PushBack(l.Back().Value.([]byte))
	}
	listLen := l.Len()
	buff := make([]byte, 64) //each hash is 32 bytes
	for i := 0; i < listLen/2; i++ {
		//"pop" off 2 vales
		v1 := l.Remove(l.Front()).([]byte)
		v2 := l.Remove(l.Front()).([]byte)
		copy(buff[0:32], v1)
		copy(buff[32:64], v2)
		l.PushBack(HashSHA256(HashSHA256(buff)))
	}
	return getMerkleRoot(l)
}

// Concatenate all the fields of the block **header**
// Return the SHA-256 hash of that concatenation
func HashBlock(b Block) []byte {
	const blength = 84 // calculate the total length of the slice
	concatenated := make([]byte, blength)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint32(concatenated[0:4], b.Version)
	binary.LittleEndian.PutUint64(concatenated[4:12], b.Height)
	binary.LittleEndian.PutUint64(concatenated[12:20], uint64(b.Timestamp))
	copy(concatenated[20:52], b.PreviousHash)
	copy(concatenated[52:84], b.MerkleRootHash)
	return HashSHA256(concatenated)
}

// Given a block in byte string form, return a block in struct form
func Deserialize(block []byte) Block {
	dataLen := binary.LittleEndian.Uint16(block[84:86])
	data := make([][]byte, dataLen)
	index := 86

	for i := 0; i < int(dataLen); i++ { // deserialize each individual element in Data
		elementLen := int(block[index])
		index += 2
		data[i] = make([]byte, elementLen)
		copy(data[i], block[index:index+elementLen])
		index += elementLen
	}

	previousHash := make([]byte, 32)
	merkleRootHash := make([]byte, 32)
	copy(previousHash, block[20:52])
	copy(merkleRootHash, block[52:84])
	// initialize the deserialized block
	deserializeBlock := Block{
		Version:        binary.LittleEndian.Uint32(block[0:4]),
		Height:         binary.LittleEndian.Uint64(block[4:12]),
		Timestamp:      int64(binary.LittleEndian.Uint64(block[12:20])),
		PreviousHash:   previousHash,
		MerkleRootHash: merkleRootHash,
		DataLen:        dataLen,
		Data:           data,
	}
	return deserializeBlock
}
