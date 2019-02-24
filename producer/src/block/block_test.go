package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"
)

func TestSerialize(t *testing.T) {
	// get time stamp
	ti := time.Now()
	nowTime := ti.UnixNano()
	//fmt.Printf("time %d \n", nowTime)

	// create the block header
	b := Block{
		Version:        3,
		Height:         300,
		PreviousHash:   []byte{'g', 'u', 'a', 'v', 'a'},
		MerkleRootHash: []byte{'g', 'r', 'a', 'p', 'e'},
		TimeStamp:      nowTime,
		Data:           [][]byte{{'1', '2'}, {'3', '4', '5'}, {'9'}},
	}
	// now use the serialize function
	serial := b.Serialize()
	fmt.Printf("serial... %x\n", serial)
	fmt.Printf("b.Version %x\n", b.Version)
	fmt.Printf("spot 0-4 bytes %d\n", binary.LittleEndian.Uint32(serial[0:4]))

	fmt.Printf("b.Height %d\n", b.Height)
	fmt.Printf("spot 4-12 bytes %d\n", binary.LittleEndian.Uint32(serial[4:12]))

	fmt.Printf("b.previousHash %x\n", b.PreviousHash)
	fmt.Printf("spot 12-44 bytes %x\n", serial[12:44])

	fmt.Printf("b.merkleRootHash %x\n", b.MerkleRootHash)
	fmt.Printf("spot 44-76 bytes %x\n", serial[44:76])

	fmt.Printf("b.TimeStamp %x\n", b.TimeStamp)
	fmt.Printf("spot 76-84 bytes %x\n", binary.LittleEndian.Uint64(serial[76:84]))

	fmt.Printf("b.data %x\n", b.Data)
	fmt.Printf("spot 84: bytes %x\n", serial[84:])

	tb := Block{}
	r := bytes.NewBuffer(serial)
	err := binary.Read(r, binary.LittleEndian, &tb.Version)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Println(tb.Version)

	err = binary.Read(r, binary.LittleEndian, &tb.Height)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Println(tb.Height)

	// reset r for the timestamp slice
	r = bytes.NewBuffer(serial[76:84])
	err = binary.Read(r, binary.LittleEndian, &tb.TimeStamp)
	fmt.Println(tb.TimeStamp)
	fmt.Println(b.TimeStamp)

	// the above method does not work with []bytes/[][]bytes for some odd reason

	/*
		/*if bytes.Equal(b.PreviousHash, serial[12:44]) != true {
			t.Errorf("Blocks' previouHash not the same as in the Serialize output\n")
		}
	*/
}
