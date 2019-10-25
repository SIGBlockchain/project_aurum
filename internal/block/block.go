// Package block contains the block struct and functions to transform a block
package block

import (
	"bytes"           // for hashing
	"encoding/binary" // for converting to uints to byte slices
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
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

// Allows for easy Marshaling into a JSON string
type JSONBlock struct {
	Version        uint16
	Height         uint64
	Timestamp      int64
	PreviousHash   string
	MerkleRootHash string
	DataLen        uint16
	Data           []string
}

func (b *Block) GetHeader() BlockHeader {
	return BlockHeader{b.Version, b.Height, b.Timestamp, b.PreviousHash, b.MerkleRootHash}
}

func New(version uint16, height uint64, previousHash []byte, data []contracts.Contract) (Block, error) {
	var serializedDatum [][]byte // A series of serialized data for Merkle root hash

	for i := range data {
		serializedData, err := data[i].Serialize()
		if err != nil {
			return Block{}, errors.New("Failed to serialize data")
		}

		serializedDatum = append(serializedDatum, serializedData)
	}

	// create the block
	block := Block{
		Version:        version,
		Height:         height,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   previousHash,
		MerkleRootHash: hashing.GetMerkleRootHash(serializedDatum),
		DataLen:        uint16(len(data)),
		Data:           serializedDatum,
	}

	return block, nil
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
func (block1 *Block) Equals(block2 Block) bool {

	blk1value := reflect.ValueOf(*block1) // get an instance of the Value struct with block1 values
	blk2value := reflect.ValueOf(block2)  // get an instance of the Value struct with block2 values

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
func (b Block) ToString() string {
	prevHash := "Previous Hash: " + hex.EncodeToString(b.PreviousHash) + "\n"
	merkleHash := "Merkle Root Hash: " + hex.EncodeToString(b.MerkleRootHash) + "\n"
	data := "Data:\n"
	for _, d := range b.Data {
		data += hex.EncodeToString(d) + "\n"
	}

	blockStr := fmt.Sprintf("Version: %v\nHeight: %v\nTimestamp: %v\n", b.Version, b.Height, b.Timestamp)
	blockStr += prevHash + merkleHash + fmt.Sprintf("DataLen: %v\n", b.DataLen) + data
	return blockStr
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
	return hashing.New(concatenated)
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
	return hashing.New(concatenated)
}

func (b *Block) Marshal() (JSONBlock, error) {
	jsonBlock := JSONBlock{
		Version:        b.Version,
		Height:         b.Height,
		Timestamp:      b.Timestamp,
		PreviousHash:   hex.EncodeToString(b.PreviousHash),
		MerkleRootHash: hex.EncodeToString(b.MerkleRootHash),
		DataLen:        b.DataLen,
	}
	jsonBlock.Data = make([]string, len(b.Data))
	for i, d := range b.Data {
		jsonBlock.Data[i] = hex.EncodeToString(d)
	}

	return jsonBlock, nil
}

// Unmarshal converts a JSONBlock to a Block
func (jB *JSONBlock) Unmarshal() Block {
	blockData := make([][]byte, jB.DataLen)
	for i, d := range jB.Data {
		blockData[i] = []byte(d)
	}

	return Block{
		Version:        jB.Version,
		Height:         jB.Height,
		Timestamp:      jB.Timestamp,
		PreviousHash:   []byte(jB.PreviousHash),
		MerkleRootHash: []byte(jB.MerkleRootHash),
		DataLen:        jB.DataLen,
		Data:           blockData,
	}
}
