// Package block contains the block struct and functions to transform a block
package block

import (
	"bytes"
	"container/list"
	"crypto/sha256"   // for hashing
	"encoding/binary" // for converting to uints to byte slices
	"encoding/hex"
	"fmt"
	"reflect"
)

type BlockHeader struct {
	Version        uint16
	Height         uint64
	Timestamp      int64
	PreviousHash   []byte
	MerkleRootHash []byte
}

// Block is a struct that represents a block in a blockchain.
type Block struct {
	Version        uint16   // Version is the version of the software this block was created with
	Height         uint64   // Height is the distance from the bottom of the tree, with the genesis block starting with height 0
	Timestamp      int64    // Timestamp is the time of creation for this block
	PreviousHash   []byte   // PreviousHash is the hash of the previous block in the blockchain,
	MerkleRootHash []byte   // MerkleRootHash is the hash of the MerkleRoot of all inputs
	DataLen        uint16   // DataLen is the number of objects in the following Data variable
	Data           [][]byte // Data is an abritrary variable, holding the actual contents of this block
}

func (b *Block) GetHeader() BlockHeader {
	return BlockHeader{b.Version, b.Height, b.Timestamp, b.PreviousHash, b.MerkleRootHash}
}

// Produces a byte string based on the block struct provided
//
// Block Header Structure:
//
//      Bytes 0-2   : Version
//      Bytes 2-10  : Height
//      Bytes 10-18 : Timestamp
//      Bytes 18-50 : Previous Hash
//      Bytes 50-82 : Merkle Root Hash
//      Bytes 82-84 : Data Length
func (b *Block) Serialize() []byte { // Vineet
	//calculate the total length beforehand, to prevent unneccessary appends
	//NOTE: 32 bit ints are used to hold lengths; unsigned 16 bit int is used for the length of Data
	bLen := 84 //size of all fixed size fields
	for _, s := range b.Data {
		bLen += 2 + len(s) //2 bytes for length plus the length of an element in Data
	}
	serializedBlock := make([]byte, bLen)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint16(serializedBlock[0:2], b.Version)
	binary.LittleEndian.PutUint64(serializedBlock[2:10], b.Height)
	binary.LittleEndian.PutUint64(serializedBlock[10:18], uint64(b.Timestamp))
	copy(serializedBlock[18:50], b.PreviousHash)
	copy(serializedBlock[50:82], b.MerkleRootHash)
	binary.LittleEndian.PutUint16(serializedBlock[82:84], b.DataLen)

	i := 84
	for _, s := range b.Data {
		//for every data entry, put the legth, and then the data
		binary.LittleEndian.PutUint16(serializedBlock[i:i+2], uint16(len(s)))
		i += 2
		copy(serializedBlock[i:i+len(s)], s)
		i += len(s)
	}
	return serializedBlock
}

// Hashes the given byte slice using SHA256 and returns it
func HashSHA256(data []byte) []byte {
	result := sha256.Sum256(data)
	return result[:]
}

// Returns the merkle root hash of the list of inputs
//
// If there are no inputs a empty slice is returned, otherwise the merkle root is generated recursively
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

// Recursive Helper function for GetMerkleRootHash()
//
// This will combine every two adjacent values, hash them, and add to the list
// This is done until the list is half of its original length.
// If the list originally had an odd length, the last element is duplicated.
// This will recursively repeat until the list has a length of one
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

// Concatenate all the fields of the block header and return its SHA256 hash
func HashBlock(b Block) []byte {
	const blength = 82 // calculate the total length of the slice
	concatenated := make([]byte, blength)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint16(concatenated[0:2], b.Version)
	binary.LittleEndian.PutUint64(concatenated[2:10], b.Height)
	binary.LittleEndian.PutUint64(concatenated[10:18], uint64(b.Timestamp))
	copy(concatenated[18:50], b.PreviousHash)
	copy(concatenated[50:82], b.MerkleRootHash)
	return HashSHA256(concatenated)
}

// Concatenate all the fields of the block header and return its SHA256 hash
func HashBlockHeader(b BlockHeader) []byte {
	const blength = 82 // calculate the total length of the slice
	concatenated := make([]byte, blength)

	// convert the known variables to byte slices in little endian and add to slice
	binary.LittleEndian.PutUint16(concatenated[0:2], b.Version)
	binary.LittleEndian.PutUint64(concatenated[2:10], b.Height)
	binary.LittleEndian.PutUint64(concatenated[10:18], uint64(b.Timestamp))
	copy(concatenated[18:50], b.PreviousHash)
	copy(concatenated[50:82], b.MerkleRootHash)
	return HashSHA256(concatenated)
}

// Converts a block in byte form into a block struct, returns the struct
func Deserialize(block []byte) Block {
	dataLen := binary.LittleEndian.Uint16(block[82:84])
	data := make([][]byte, dataLen)
	index := 84

	for i := 0; i < int(dataLen); i++ { // deserialize each individual element in Data
		elementLen := int(binary.LittleEndian.Uint16(block[index : index+2]))
		index += 2
		data[i] = make([]byte, elementLen)
		copy(data[i], block[index:index+elementLen])
		index += elementLen
	}

	previousHash := make([]byte, 32)
	merkleRootHash := make([]byte, 32)
	copy(previousHash, block[18:50])
	copy(merkleRootHash, block[50:82])
	// initialize the deserialized block
	deserializeBlock := Block{
		Version:        binary.LittleEndian.Uint16(block[0:2]),
		Height:         binary.LittleEndian.Uint64(block[2:10]),
		Timestamp:      int64(binary.LittleEndian.Uint64(block[10:18])),
		PreviousHash:   previousHash,
		MerkleRootHash: merkleRootHash,
		DataLen:        dataLen,
		Data:           data,
	}
	return deserializeBlock
}

// Compares two block structs and returns true if all the fields in both blocks are equal, false otherwise
func Equals(block1 Block, block2 Block) bool {

	blk1value := reflect.ValueOf(block1) // get an instance of the Value struct with block1 values
	blk2value := reflect.ValueOf(block2) // get an instance of the Value struct with block2 values

	for i := 0; i < blk1value.NumField(); i++ { // loop through the fields of both blocks
		finterface1 := blk1value.Field(i).Interface() // get the value of the current field from block1 as an interface{}
		finterface2 := blk2value.Field(i).Interface() // get the value of the current field from block2 as an interface{}

		switch finterface1.(type) { // type switch
		case uint16, uint64, int64:
			if finterface1 != finterface2 {
				return false
			}
		case []byte:
			if !bytes.Equal(finterface1.([]byte), finterface2.([]byte)) {
				return false
			}
		case [][]byte:
			for i := 0; i < len(finterface1.([][]byte)); i++ {
				if !bytes.Equal(finterface1.([][]byte)[i], finterface2.([][]byte)[i]) {
					return false
				}
			}
		}
	}
	return true
}

// Returns a string of the block
func (b Block) toString() string {
	prevHash := hex.EncodeToString(b.PreviousHash) + "\n"
	merkleHash := hex.EncodeToString(b.MerkleRootHash) + "\n"
	var data string
	for i, d := range b.Data {
		if i == len(b.Data)-1 {
			data += hex.EncodeToString(d)
		} else {
			data += hex.EncodeToString(d) + "\n"
		}
	}

	blockStr := fmt.Sprintf("%v\n%v\n%v\n", b.Version, b.Height, b.Timestamp)
	blockStr += prevHash + merkleHash + fmt.Sprintf("%v\n", b.DataLen) + data
	fmt.Println(blockStr)
	return blockStr
}
