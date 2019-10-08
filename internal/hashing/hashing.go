package hashing

import (
	"bytes"
	"container/list"
	"crypto/sha256"
)

type SHA256Hash struct {
	SecureHash []byte
}

// Hashes the given byte slice using SHA256 and returns it
func New(data []byte) []byte {
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
		l.PushBack(New(New(s)))
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
		l.PushBack(New(New(buff)))
	}
	return getMerkleRoot(l)
}

// Equals returns true if the SHA256 hash is the hash of the given
// byte slice, false otherwise
func (hash SHA256Hash) Equals(bSlice []byte) bool {
	return bytes.Equal(hash.SecureHash, New(bSlice))
}

// MerkleRootHashCompare determines if the merkle-root hash is the merkle root of the array of hashes
func MerkleRootHashOf(merkRHash []byte, sha256Hashes [][]byte) bool {
	return bytes.Equal(GetMerkleRootHash(sha256Hashes), merkRHash)
}
