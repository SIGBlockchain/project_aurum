package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"
	"github.com/google/go-cmp/cmp"
)

func TestSerialize(t *testing.T) {
	// get time stamp
	ti := time.Now()
	nowTime := ti.UnixNano()

	// create the block
	b := Block{
		Version:        3,
		Height:         300,
		PreviousHash:   []byte("guavapineapplemango1234567890abc"),
		MerkleRootHash: []byte("grapewatermeloncoconut1emonsabcd"),
		Timestamp:      nowTime,
		Data:           [][]byte{{12, 3}, {132, 90, 23}, {23}},
	}
	// set data length
	b.DataLen = uint16(len(b.Data))
	// now use the serialize function
	serial := b.Serialize()
	// indicies are fixed since we know what the max sizes are going to be

	// check Version
	blockVersion := binary.LittleEndian.Uint16(serial[0:2])
	if blockVersion != b.Version {
		t.Errorf("Version does not match")
	}

	// check Height
	blockHeight := binary.LittleEndian.Uint64(serial[2:10])
	if blockHeight != b.Height {
		t.Errorf("Height does not match")
	}

	// check Timestamp
	blockTimestamp := binary.LittleEndian.Uint64(serial[10:18])
	if int64(blockTimestamp) != b.Timestamp {
		t.Errorf("Timestamps do not match")
	}

	// check PreviousHash
	blockPrevHash := serial[18:50]
	if bytes.Equal(blockPrevHash, b.PreviousHash) != true {
		t.Errorf("PreviousHashes do not match")
	}

	// check MerkleRootHash
	blockMerkleHash := serial[50:82]
	if bytes.Equal(blockMerkleHash, b.MerkleRootHash) != true {
		t.Errorf("MerkleRootHashes do not match")
	}

	// check DataLen
	blockDataLen := binary.LittleEndian.Uint16(serial[82:84])
	if blockDataLen != b.DataLen {
		t.Errorf("DataLen does not match")
	}

	// check Data
	testslice := [][]byte{{12, 3}, {132, 90, 23}, {23}}
	dataLen := int(blockDataLen)
	blockData := make([][]byte, dataLen)
	index := 84

	for i := 0; i < dataLen; i++ {
		elementLen := int(serial[index])
		index += 2
		blockData[i] = serial[index : index+elementLen]
		index += elementLen
	}

	for i := 0; i < dataLen; i++ {
		if bytes.Compare(testslice[i], blockData[i]) != 0 {
			t.Errorf("Data does not match")
		}
	}
}

func TestDeserialize(t *testing.T) {
	expected := Block{
		Version:        1,
		Height:         0,
		PreviousHash:   hashing.New([]byte{'x'}),
		MerkleRootHash: hashing.New([]byte{'q'}),
		Timestamp:      time.Now().UnixNano(),
		Data:           [][]byte{hashing.New([]byte{'r'})},
	}
	expected.DataLen = uint16(len(expected.Data))
	intermed := expected.Serialize()
	actual := Deserialize(intermed)
	if !cmp.Equal(expected, actual) {
		t.Errorf("Blocks do not match")
	}

	//change itermed to see if that changes the deserialized block
	intermed[18] = uint8(21)
	intermed[54] = uint8(21)
	if !cmp.Equal(expected, actual) {
		t.Errorf("Blocks do not match")
	}

}

func TestNew(t *testing.T) {
	var datum []contracts.Contract
	for i := 0; i < 12; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		someAirdropContract, _ := contracts.New(1, nil, someKeyPKHash, 1000, 0)
		datum = append(datum, *someAirdropContract)
	}
	var serializedDatum [][]byte
	for i := range datum {
		serialData, _ := datum[i].Serialize()
		serializedDatum = append(serializedDatum, serialData)
	}
	type args struct {
		version      uint16
		height       uint64
		previousHash []byte
		data         []contracts.Contract
	}
	tests := []struct {
		name    string
		args    args
		want    Block
		wantErr bool
	}{
		{
			args: args{
				version:      1,
				height:       0,
				previousHash: make([]byte, 32),
				data:         datum,
			},
			wantErr: false,
			want: Block{
				Version:        1,
				Height:         0,
				Timestamp:      time.Now().UnixNano(),
				PreviousHash:   make([]byte, 32),
				MerkleRootHash: hashing.GetMerkleRootHash(serializedDatum),
				Data:           serializedDatum,
				DataLen:        12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.version, tt.args.height, tt.args.previousHash, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Version, tt.want.Version) ||
				!reflect.DeepEqual(got.Height, tt.want.Height) ||
				!reflect.DeepEqual(got.PreviousHash, tt.want.PreviousHash) ||
				!reflect.DeepEqual(got.MerkleRootHash, tt.want.MerkleRootHash) ||
				!reflect.DeepEqual(got.DataLen, tt.want.DataLen) ||
				!reflect.DeepEqual(got.Data, tt.want.Data) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_GetHeader(t *testing.T) {
	expected := Block{
		Version:        1,
		Height:         0,
		PreviousHash:   hashing.New([]byte{'x'}),
		MerkleRootHash: hashing.New([]byte{'q'}),
		Timestamp:      time.Now().UnixNano(),
		Data:           [][]byte{hashing.New([]byte{'r'})},
	}
	expected.DataLen = uint16(len(expected.Data))
	tests := []struct {
		name string
		b    *Block
		want BlockHeader
	}{
		{
			b: &expected,
			want: BlockHeader{
				Version:        1,
				Height:         0,
				PreviousHash:   hashing.New([]byte{'x'}),
				MerkleRootHash: hashing.New([]byte{'q'}),
				Timestamp:      expected.Timestamp,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.b.GetHeader()
			v := reflect.ValueOf(got)
			values := make([]interface{}, v.NumField())
			for i := 0; i < v.NumField(); i++ {
				values[i] = v.Field(i).Interface()
			}

			v = reflect.ValueOf(BlockHeader{
				Version:        1,
				Height:         0,
				PreviousHash:   hashing.New([]byte{'x'}),
				MerkleRootHash: hashing.New([]byte{'q'}),
				Timestamp:      expected.Timestamp,
			})
			for i := 0; i < v.NumField(); i++ {
				if !reflect.DeepEqual(values[i], v.Field(i).Interface()) {
					t.Error("fields do not match")
				}
			}

		})
	}
}

func TestEquals(t *testing.T) {
	block1 := Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'a'}),
		MerkleRootHash: hashing.New([]byte{'b'}),
		DataLen:        1,
		Data:           [][]byte{hashing.New([]byte{'c'}), hashing.New([]byte{'g'})},
	}

	blocks := make([]Block, 7)
	for i := 0; i < 7; i++ {
		blocks[i] = block1
	}
	blocks[0].Version = 5
	blocks[1].Height = 10
	blocks[2].Timestamp = time.Now().UnixNano() + 100
	blocks[3].PreviousHash = hashing.New([]byte{'d'})
	blocks[4].MerkleRootHash = hashing.New([]byte{'e'})
	blocks[5].DataLen = 15
	blocks[6].Data = [][]byte{hashing.New([]byte{'f'}), hashing.New([]byte{'o'})}

	tests := []struct {
		name string
		b1   Block
		b2   Block
		want bool
	}{
		{
			name: "equal blocks",
			b1:   block1,
			b2:   block1,
			want: true,
		},
		{
			name: "different block version",
			b1:   block1,
			b2:   blocks[0],
			want: false,
		},
		{
			name: "different block height",
			b1:   block1,
			b2:   blocks[1],
			want: false,
		},
		{
			name: "different block timestamp",
			b1:   block1,
			b2:   blocks[2],
			want: false,
		},
		{
			name: "different block previousHash",
			b1:   block1,
			b2:   blocks[3],
			want: false,
		},
		{
			name: "different block merklerootHash",
			b1:   block1,
			b2:   blocks[4],
			want: false,
		},
		{
			name: "different block dataLen",
			b1:   block1,
			b2:   blocks[5],
			want: false,
		},
		{
			name: "different block data",
			b1:   block1,
			b2:   blocks[6],
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := Equals(tt.b1, tt.b2); result != tt.want {
				t.Errorf("Error: Equals() returned %v for %s\n Wanted: %v", result, tt.name, tt.want)
			}
		})
	}
}

func TestBlockToString(t *testing.T) {
	testBlock := Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'a'}),
		MerkleRootHash: hashing.New([]byte{'b'}),
		DataLen:        1,
		Data:           [][]byte{hashing.New([]byte{'c'}), hashing.New([]byte{'g'})},
	}
	nilblock := Block{}

	tests := []struct {
		blk Block
	}{
		{
			blk: testBlock,
		},
		{
			blk: nilblock,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := tt.blk.ToString()

			expected := fmt.Sprintf("Version: %v\nHeight: %v\nTimestamp: %v\nPrevious Hash: %v\nMerkle Root Hash: %v\nDataLen: %v\n",
				tt.blk.Version, tt.blk.Height, tt.blk.Timestamp, hex.EncodeToString(tt.blk.PreviousHash),
				hex.EncodeToString(tt.blk.MerkleRootHash), tt.blk.DataLen)
			data := "Data:\n"
			for _, d := range tt.blk.Data {
				data += hex.EncodeToString(d) + "\n"
			}
			expected += data
			if actual != expected {
				t.Errorf("The strings are not equal\nExpected:\n%+v\nActual:\n%+v", expected, actual)
			}
		})
	}
}

func TestHashBlockHeader(t *testing.T) {
	expected := BlockHeader{
		Version:        1,
		Height:         0,
		PreviousHash:   hashing.New([]byte{'x'}),
		MerkleRootHash: hashing.New([]byte{'q'}),
		Timestamp:      time.Now().UnixNano(),
	}
	type args struct {
		b BlockHeader
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			args: args{
				b: expected,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := HashBlockHeader(tt.args.b); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("HashBlockHeader() = %v, want %v", got, tt.want)
			// }
		})
	}
}
