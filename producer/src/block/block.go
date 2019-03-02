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
func HashSHA256(data []byte) []byte {
	result := sha256.Sum256(data)
	return result[:]
}

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
	if l.Len()%2 == 1 { //list is of odd length
		l.PushBack(l.Back().Value.([]byte))
	}
	return getMerkleRoot(l)
}

//recursive helper fucntion
func getMerkleRoot(l *list.List) []byte {
	if l.Len() == 1 {
		return l.Front().Value.([]byte)
	}

	listLen := l.Len()
	for i := 0; i < listLen / 2; i++ {
		//"pop" off 2 vales
		v1 := l.Remove(l.Front()).([]byte)
		v2 := l.Remove(l.Front()).([]byte)
		concat := append(v1, v2...)
		l.PushBack(HashSHA256(HashSHA256(concat)))
	}
	return getMerkleRoot(l)
}
