package block

import ("encoding/binary" // for converting to uints to byte slices
	//"crypto/sha256" //hashing library:has not been implemented yet
	"fmt"//debugging purposes
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

func appendString(A string, B string)string{ //accepts two strings, returns single string
	A += B
	return A
}
func recursiveAdd(myArray[] string)string{
	if(len(myArray)==0){//if empty slice is passed
		return ""
	}else if (len(myArray) == 1){//end of function
		return myArray[0]
	}
	tempArray := []string{} //will be used to append to

	if(len(myArray)%2 == 1){ //if array is of odd length, makes it even
		myArray = append(myArray,myArray[len(myArray)-1] ) //append last element to self to make it even
	}
	
	for i:=0;i<len(myArray);i=i+2{
		tempArray = append(tempArray, appendString(myArray[i],myArray[i+1] ))//calls append, which returns a single string
	}
	if(false){//debugging code, used to print out element
		for i:=0;i<len(tempArray);i++{
			fmt.Println(tempArray[i])
		}
	}
	return recursiveAdd(tempArray) //recursion, returns single string
}
