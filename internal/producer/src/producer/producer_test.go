package producer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

// Test will fail in airplane mode, or just remove wireless connection.
func TestCheckConnectivity(t *testing.T) {
	err := CheckConnectivity()
	if err != nil {
		t.Errorf("Internet connection check failed.")
	}
}

// Tests a single connection
func TestAcceptConnections(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:10000")
	var buffer bytes.Buffer
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        log.New(&buffer, "LOG:", log.Ldate),
	}
	go bp.AcceptConnections()
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Errorf("Failed to connect to server")
	}
	contentsOfChannel := <-bp.NewConnection
	actual := contentsOfChannel.RemoteAddr().String()
	expected := conn.LocalAddr().String()
	if actual != expected {
		t.Errorf("Failed to store connection")
	}
	conn.Close()
	ln.Close()
}

// Sends a message to the listener and checks to see if it echoes back
func TestHandler(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:10000")
	var buffer bytes.Buffer
	bp := BlockProducer{
		Server:        ln,
		NewConnection: make(chan net.Conn, 128),
		Logger:        log.New(&buffer, "LOG:", log.Ldate),
	}
	go bp.AcceptConnections()
	go bp.WorkLoop()
	conn, err := net.Dial("tcp", ":10000")
	if err != nil {
		t.Errorf("Failed to connect to server")
	}
	expected := []byte("This is a test.")
	conn.Write(expected)
	actual := make([]byte, len(expected))
	_, readErr := conn.Read(actual)
	if readErr != nil {
		t.Errorf("Failed to read from socket.")
	}
	if bytes.Equal(expected, actual) == false {
		t.Errorf("Message mismatch")
	}
	conn.Close()
	ln.Close()
}

func TestData_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))
	initialContract, _ := accounts.MakeContract(1, nil, spkh, 1000, 0)
	tests := []struct {
		name string
		d    *Data
	}{
		{
			d: &Data{
				Hdr: DataHeader{
					Version: 1,
					Type:    0,
				},
				Bdy: initialContract,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Serialize()
			if err != nil {
				t.Errorf(err.Error())
			}
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked, check indexing")
				}
			}()
			serializedInitialContract, err := tt.d.Bdy.Serialize()
			if err != nil {
				t.Errorf(err.Error())
			}
			serializedVersion := make([]byte, 2)
			binary.LittleEndian.PutUint16(serializedVersion, 1)
			serializedType := make([]byte, 2)
			binary.LittleEndian.PutUint16(serializedType, 0)
			if !bytes.Equal(got[:2], serializedVersion) {
				t.Errorf(fmt.Sprintf("Data header version serialization does not match. Wanted: %v, got: %v", serializedVersion, got[:2]))
			}
			if !bytes.Equal(got[2:4], serializedType) {
				t.Errorf(fmt.Sprintf("Data header type serialization does not match. Wanted: %v, got: %v", serializedVersion, got[2:4]))
			}
			if !bytes.Equal(got[4:], serializedInitialContract) {
				t.Errorf(fmt.Sprintf("Data header body serialization does not match. Wanted: %v, got: %v", serializedVersion, got[4:]))
			}
		})
	}
}

func TestData_Deserialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))
	initialContract, _ := accounts.MakeContract(1, nil, spkh, 1000, 0)
	someData := &Data{
		Hdr: DataHeader{
			Version: 1,
			Type:    0,
		},
		Bdy: initialContract,
	}
	serializedsomeData, _ := someData.Serialize()
	type args struct {
		serializedData []byte
	}
	tests := []struct {
		name    string
		d       *Data
		args    args
		wantErr bool
	}{
		{
			d: &Data{},
			args: args{
				serializedData: serializedsomeData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Deserialize(tt.args.serializedData); (err != nil) != tt.wantErr {
				t.Errorf("Data.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.d, someData) {
				t.Errorf("Deserialized Data struct failed to match")
			}
		})
	}
}

func TestCreateBlock(t *testing.T) {
	var datum []Data
	for i := 0; i < 50; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		someAirdropContract, _ := accounts.MakeContract(1, nil, someKeyPKHash, 1000, 0)
		someDataHdr := DataHeader{
			Version: 1,
			Type:    0,
		}
		someData := Data{
			Hdr: someDataHdr,
			Bdy: someAirdropContract,
		}
		datum = append(datum, someData)
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
		data         []Data
	}
	tests := []struct {
		name    string
		args    args
		want    block.Block
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
			want: block.Block{
				Version:        1,
				Height:         0,
				Timestamp:      time.Now().UnixNano(),
				PreviousHash:   make([]byte, 32),
				MerkleRootHash: block.GetMerkleRootHash(serializedDatum),
				Data:           serializedDatum,
				DataLen:        50,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateBlock(tt.args.version, tt.args.height, tt.args.previousHash, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Version, tt.want.Version) ||
				!reflect.DeepEqual(got.Height, tt.want.Height) ||
				!reflect.DeepEqual(got.PreviousHash, tt.want.PreviousHash) ||
				!reflect.DeepEqual(got.MerkleRootHash, tt.want.MerkleRootHash) ||
				!reflect.DeepEqual(got.DataLen, tt.want.DataLen) ||
				!reflect.DeepEqual(got.Data, tt.want.Data) {
				t.Errorf("CreateBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBringOnTheGenesis(t *testing.T) {
	var pkhashes [][]byte
	var datum []Data
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
		someAirdropContract, _ := accounts.MakeContract(1, nil, someKeyPKHash, 10, 0)
		someDataHdr := DataHeader{
			Version: 1,
			Type:    0,
		}
		someData := Data{
			Hdr: someDataHdr,
			Bdy: someAirdropContract,
		}
		datum = append(datum, someData)
	}
	genny, _ := CreateBlock(1, 0, make([]byte, 32), datum)
	type args struct {
		genesisPublicKeyHashes [][]byte
		initialAurumSupply     uint64
	}
	tests := []struct {
		name    string
		args    args
		want    block.Block
		wantErr bool
	}{
		{
			args: args{
				genesisPublicKeyHashes: pkhashes,
				initialAurumSupply:     1000,
			},
			want:    genny,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BringOnTheGenesis(tt.args.genesisPublicKeyHashes, tt.args.initialAurumSupply)
			if (err != nil) != tt.wantErr {
				t.Errorf("BringOnTheGenesis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Version, tt.want.Version) ||
				!reflect.DeepEqual(got.Height, tt.want.Height) ||
				!reflect.DeepEqual(got.PreviousHash, tt.want.PreviousHash) ||
				!reflect.DeepEqual(got.MerkleRootHash, tt.want.MerkleRootHash) ||
				!reflect.DeepEqual(got.DataLen, tt.want.DataLen) ||
				!reflect.DeepEqual(got.Data, tt.want.Data) {
				t.Errorf("BringOnTheGenesis() = %v, want %v", got, tt.want)
			}
			for i := range got.Data {
				deserializedData := Data{}
				err := deserializedData.Deserialize(got.Data[i])
				if err != nil {
					t.Errorf("failed to deserialize data")
				}
				deserializedContract := &accounts.Contract{}
				serializedDataBdy, _ := deserializedData.Bdy.Serialize()
				deserializedContract.Deserialize(serializedDataBdy)
				if deserializedContract.Value != 10 {
					t.Errorf("failed to distribute aurum properly")
				}
			}
		})
	}
}

func TestAirdrop(t *testing.T) {
	var pkhashes [][]byte
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := BringOnTheGenesis(pkhashes, 1000)
	type args struct {
		blockchain   string
		metadata     string
		genesisBlock block.Block
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				blockchain:   "blockchain.dat",
				metadata:     "metadata.tab",
				genesisBlock: genny,
			},
		},
	}
	for _, tt := range tests {
		defer func() {
			os.Remove(tt.args.metadata)
			os.Remove(tt.args.blockchain)
		}()
		t.Run(tt.name, func(t *testing.T) {
			if err := Airdrop(tt.args.blockchain, tt.args.metadata, tt.args.genesisBlock); (err != nil) != tt.wantErr {
				t.Errorf("Airdrop() error = %v, wantErr %v", err, tt.wantErr)
			}
			fileGenny, err := ioutil.ReadFile(tt.args.blockchain)
			if err != nil {
				t.Errorf("Failed to open file" + err.Error())
			}
			serializedGenny := genny.Serialize()
			if !bytes.Equal(fileGenny[4:], serializedGenny) {
				t.Errorf("Genesis block does not match file block")
			}
		})
	}
}
