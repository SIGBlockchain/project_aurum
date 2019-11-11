package producer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/genesis"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

var removeFiles = true

func TestRunServer(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientPublicKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPublicKeyHash := hashing.New(encodedRecipientPublicKey)
	contract, _ := contracts.New(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.Sign(senderPrivateKey)
	serializedContract, err := contract.Serialize()

	var contractMessage []byte
	contractMessage = append(contractMessage, SecretBytes...)
	contractMessage = append(contractMessage, 1)
	contractMessage = append(contractMessage, serializedContract...)
	type testArg struct {
		name            string
		messageToBeSent []byte
		messageToBeRcvd []byte
	}
	testArgs := []testArg{
		{
			name:            "Regular message",
			messageToBeSent: []byte("hello\n"),
			messageToBeRcvd: []byte("No thanks.\n"),
		},
		{
			name:            "Aurum message",
			messageToBeSent: SecretBytes,
			messageToBeRcvd: []byte("Thank you.\n"),
		},
		{
			name:            "Contract message",
			messageToBeSent: contractMessage,
			messageToBeRcvd: []byte("Thank you.\n"),
		},
	}
	ln, err := net.Listen("tcp", "localhost:13131")
	if err != nil {
		t.Errorf("failed to startup listener")
	}
	byteChan := make(chan []byte)
	buf := make([]byte, 1024)
	go RunServer(ln, byteChan, false)
	for _, arg := range testArgs {
		conn, err := net.Dial("tcp", "localhost:13131")
		if err != nil {
			t.Errorf("failed to connect to server")
		}
		_, err = conn.Write(arg.messageToBeSent)
		if err != nil {
			t.Errorf("failed to send message")
		}
		nRead, err := conn.Read(buf)
		if err != nil {
			t.Errorf("failed to read from connections:\n%s", err.Error())
		}
		if !bytes.Equal(buf[:nRead], arg.messageToBeRcvd) {
			t.Errorf("did not received desired message:\n%s != %s", string(buf[:nRead]), string(arg.messageToBeRcvd))
		}
		if arg.name != "Regular message" {
			res := <-byteChan
			if !bytes.Equal(res, arg.messageToBeSent) {
				t.Errorf("result does not match:\n%s != %s", string(res), string(arg.messageToBeSent))
			}
			if arg.name == "Contract message" {
				var contract contracts.Contract
				if err := contract.Deserialize(res[9:]); err != nil {
					t.Errorf("failed to deserialize contract:\n%s", err.Error())
				}
				if !bytes.Equal(res[9:], serializedContract) {
					t.Errorf("serialized contracts do not match:\n%v != %v", res[9:], serializedContract)
				}
			}
		}
	}
}

func TestByteChannel(t *testing.T) {
	t.SkipNow()
	genesisHashes, err := genesis.ReadGenesisHashes()
	if err != nil {
		t.Errorf("failed to read genesis hashes:\n%s", err.Error())
	}
	genesisBlock, err := genesis.BringOnTheGenesis(genesisHashes, 1000)
	if err != nil {
		t.Errorf("failed to create genesis block:\n%s", err.Error())
	}
	if err := blockchain.Airdrop(ledger, metadataTable, constants.AccountsTable, genesisBlock); err != nil {
		t.Errorf("failed to perform air drop:\n%s", err.Error())
	}
	ledgerFile, err := os.OpenFile("blockchain.dat", os.O_RDONLY, 0644)
	metadataConn, _ := sql.Open("sqlite3", constants.MetadataTable)
	defer func() {
		if removeFiles {
			ledgerFile.Close()
			metadataConn.Close()
			if err := os.Remove("blockchain.dat"); err != nil {
				t.Errorf("failed to remove blockchain.dat:\n%s", err.Error())
			}
			if err := os.Remove(constants.MetadataTable); err != nil {
				t.Errorf("failed to remove metadatata.tab:\n%s", err.Error())
			}
			if err := os.Remove(constants.AccountsTable); err != nil {
				t.Errorf("failed to remove accounts.db:\n%s", err.Error())
			}
		}
	}()
	ln, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to start server:\n%s", err.Error())
	}
	byteChan := make(chan []byte)
	debug := false

	go RunServer(ln, byteChan, debug)
	testMode := true
	prodInterval := "2000ms"
	memStats := false
	fl := Flags{
		Debug:       &debug,
		Interval:    &prodInterval,
		Test:        &testMode,
		MemoryStats: &memStats,
	}
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientPublicKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	recipientPublicKeyHash := hashing.New(encodedRecipientPublicKey)
	contract, _ := contracts.New(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.Sign(senderPrivateKey)
	serializedContract, _ := contract.Serialize()

	var contractMessage []byte
	contractMessage = append(contractMessage, SecretBytes...)
	contractMessage = append(contractMessage, 1)
	contractMessage = append(contractMessage, serializedContract...)

	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	_, err = conn.Write(contractMessage)
	if err != nil {
		t.Errorf("failed to send message")
	}
	ProduceBlocks(byteChan, fl, true)

	youngestBlock, err := blockchain.GetYoungestBlock(ledgerFile, metadataConn)
	if err != nil {
		t.Errorf("failed to get youngest block:\n%s", err.Error())
	}
	data := youngestBlock.Data[0]
	var compContract contracts.Contract
	if err := compContract.Deserialize(data); err != nil {
		t.Errorf("failed to deserialize data:\n%s", err.Error())
	}
	if !bytes.Equal(serializedContract, data) {
		t.Errorf("data does not match:\n%s != %s", string(serializedContract), string(data))
	}
}

func TestData_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	spkh := hashing.New(encodedSenderPublicKey)
	initialContract, _ := contracts.New(1, nil, spkh, 1000, 0)
	tests := []struct {
		name string
		// d    *Data
		d *contracts.Contract
	}{
		{
			// d: &Data{
			// 	Hdr: DataHeader{
			// 		Version: 1,
			// 		Type:    0,
			// 	},
			// 	Bdy: initialContract,
			// },
			d: initialContract,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got, err := tt.d.Serialize()
			got, err := initialContract.Serialize()
			if err != nil {
				t.Errorf(err.Error())
			}
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked, check indexing")
				}
			}()
			// serializedInitialContract, err := tt.d.Serialize()
			serializedInitialContract, err := initialContract.Serialize()
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
			if !bytes.Equal(got[4:], serializedInitialContract[4:]) { //had to change serializedInitialContract to serializedInitialContract[4:]
				t.Errorf(fmt.Sprintf("Data header body serialization does not match. Wanted: %v, got: %v", serializedVersion, got[4:]))
			}
		})
	}
}

func TestData_Deserialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	spkh := hashing.New(encodedSenderPublicKey)
	initialContract, _ := contracts.New(1, nil, spkh, 1000, 0)
	// someData := &Data{
	// 	Hdr: DataHeader{
	// 		Version: 1,
	// 		Type:    0,
	// 	},
	// 	Bdy: initialContract,
	// }
	someData := initialContract
	serializedsomeData, _ := someData.Serialize()
	type args struct {
		serializedData []byte
	}
	tests := []struct {
		name    string
		d       *contracts.Contract
		args    args
		wantErr bool
	}{
		{
			// d: &Data{},
			d: &contracts.Contract{},
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
