package producer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"

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

func TestRecoverBlockchainMetadata(t *testing.T) {
	var ljr = "blockchain.dat"
	var meta = "metadata.tab"
	var accts = "accounts.tab"

	if file, err := os.Create(ljr); err != nil {
		t.Errorf("Failed to create file.")
	} else {
		file.Close()
	}
	defer func() {
		if err := os.Remove(ljr); err != nil {
			t.Errorf("failed to remove blockchain file")
		}
	}()

	if conn, err := sql.Open("sqlite3", meta); err != nil {
		t.Errorf("failed to create metadata file")
	} else {
		statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
		statement.Exec()
		conn.Close()
	}
	defer func() {
		if err := os.Remove(meta); err != nil {
			t.Errorf("failed to remove metadata file")
		}
	}()

	if conn, err := sql.Open("sqlite3", accts); err != nil {
		t.Errorf("failed to create accounts file")
	} else {
		statement, _ := conn.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
		statement.Exec()
		conn.Close()
	}

	defer func() {
		if err := os.Remove(accts); err != nil {
			t.Errorf("failed to remove accounts file" + err.Error())
		}
	}()

	var pkhashes [][]byte
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := BringOnTheGenesis(pkhashes, 1000)
	if err := Airdrop(ljr, meta, genny); err != nil {
		t.Errorf("airdrop failed")
	}

	type args struct {
		ledgerFilename      string
		metadataFilename    string
		accountBalanceTable string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				ledgerFilename:      ljr,
				metadataFilename:    meta,
				accountBalanceTable: accts,
			},
		},
	}
	var blockchainHeightIdx = 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RecoverBlockchainMetadata(tt.args.ledgerFilename, tt.args.metadataFilename, tt.args.accountBalanceTable); (err != nil) != tt.wantErr {
				t.Errorf("RecoverBlockchainMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
			blockchainGenesisBlockSerialized, err := blockchain.GetBlockByHeight(blockchainHeightIdx, ljr, meta)
			if err != nil {
				t.Errorf("failed to get genesis block")
			}
			blockchainGenesisBlockDeserialized := block.Deserialize(blockchainGenesisBlockSerialized)
			if !reflect.DeepEqual(blockchainGenesisBlockDeserialized, genny) {
				t.Errorf("genesis blocks do not match")
			}
			dbc, _ := sql.Open("sqlite3", accts)
			defer func() { // not sure if this defer will happen before the others, is it stack based?
				if err := dbc.Close(); err != nil {
					t.Errorf("Failed to close database: %s", err)
				}
			}()
			for _, hsh := range pkhashes {
				var pkhash string
				var balance uint64
				var nonce uint64
				foundKey := false
				rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
				if err != nil {
					t.Errorf("Failed to acquire rows from table")
				}
				for rows.Next() {
					err = rows.Scan(&pkhash, &balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan rows: %s", err)
					}
					decodedPkhash, err := hex.DecodeString(pkhash)
					if err != nil {
						t.Errorf("failed to decode public key hash")
					}
					if bytes.Equal(hsh, decodedPkhash) {
						foundKey = true
						if balance != 10 {
							t.Errorf("wrong balance on key: %v", hsh)
						}
						if nonce != 0 {
							t.Errorf("wrong nonce on key: %v", hsh)
						}
					}
				}
				if !foundKey {
					t.Errorf("Key not found in table: %v", hsh)
				}
			}
		})
		blockchainHeightIdx++
	}
}

func TestRecoverBlockchainMetadata_TwoBlocks(t *testing.T) {
	var ljr = "blockchain.dat"
	var meta = "metadata.tab"
	var accts = "accounts.tab"

	// Create ledger file and the two tables
	file, err := os.Create(ljr)
	if err != nil {
		t.Errorf("Failed to create file.")
	} else {
		file.Close()
	}
	metaDB, err := sql.Open("sqlite3", meta)
	if err != nil {
		t.Errorf("failed to create metadata file")
	} else {
		metaDB.Exec("CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
		metaDB.Close()
	}
	acctsDB, err := sql.Open("sqlite3", accts)
	if err != nil {
		t.Errorf("failed to create accounts file")
	} else {
		acctsDB.Exec("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	}

	defer func() {
		if err := os.Remove(ljr); err != nil {
			t.Errorf("failed to remove blockchain file")
		}
		if err := os.Remove(meta); err != nil {
			t.Errorf("failed to remove metadata file")
		}
		if err := os.Remove(accts); err != nil {
			t.Errorf("failed to remove accounts file")
		}
	}()

	var pkhashes [][]byte
	somePVKeys := make([]*ecdsa.PrivateKey, 3) // Grab 3 private keys for creating contracts
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
		if i < 3 {
			somePVKeys[i] = someKey
		}
	}
	genesisBlk, _ := BringOnTheGenesis(pkhashes, 1000)
	if err := Airdrop(ljr, meta, genesisBlk); err != nil {
		t.Errorf("airdrop failed")
	}
	// Insert pkhashes into account table for contract validation
	for i := 0; i < 100; i++ {
		err := accounts.InsertAccountIntoAccountBalanceTable(acctsDB, pkhashes[i], 10)
		if err != nil {
			t.Errorf("Failed to insert pkhash (%v) into account table: %s", pkhashes[i], err.Error())
		}
	}
	acctsDB.Close()

	// Create 3 contracts
	contracts := make([]*accounts.Contract, 3)
	var datum []Data

	recipPKHash := block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[1].PublicKey)))
	contract1, _ := accounts.MakeContract(1, somePVKeys[0], recipPKHash, 5, 1) // pkh1 to pkh2
	contract1.SignContract(somePVKeys[0])
	valid, err := accounts.ValidateContract(contract1, accts, make([][]byte, 0))
	if err != nil {
		t.Error("Failed to validate contract: " + err.Error())
	} else if !valid {
		t.Error("Invalid contract")
	}
	recipPKHash = block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[2].PublicKey)))
	contract2, _ := accounts.MakeContract(1, somePVKeys[1], recipPKHash, 7, 2) // pkh2 to pkh3
	contract2.SignContract(somePVKeys[1])
	valid, err = accounts.ValidateContract(contract2, accts, make([][]byte, 0))
	if err != nil {
		t.Error("Failed to validate contract: " + err.Error())
	} else if !valid {
		t.Error("Invalid contract")
	}
	recipPKHash = block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[1].PublicKey)))
	contract3, _ := accounts.MakeContract(1, somePVKeys[2], recipPKHash, 5, 2) // pkh3 to pkh2
	contract3.SignContract(somePVKeys[2])
	valid, err = accounts.ValidateContract(contract3, accts, make([][]byte, 0))
	if err != nil {
		t.Error("Failed to validate contract: " + err.Error())
	} else if !valid {
		t.Error("Invalid contract")
	}

	contracts[0] = contract1
	contracts[1] = contract2
	contracts[2] = contract3
	ct1Data := Data{
		Hdr: DataHeader{
			Version: 1,
			Type:    0,
		},
		Bdy: contract1,
	}
	ct2Data := Data{
		Hdr: DataHeader{
			Version: 1,
			Type:    0,
		},
		Bdy: contract2,
	}
	ct3Data := Data{
		Hdr: DataHeader{
			Version: 1,
			Type:    0,
		},
		Bdy: contract3,
	}
	datum = append(datum, ct1Data)
	datum = append(datum, ct2Data)
	datum = append(datum, ct3Data)

	firstBlock, err := CreateBlock(1, 1, block.HashBlock(genesisBlk), datum)
	if err != nil {
		t.Errorf("failed to create first block")
	}
	err = blockchain.AddBlock(firstBlock, ljr, meta)
	if err != nil {
		t.Errorf("failed to add first block")
	}

	os.Remove(meta)
	os.Remove(accts)
	type args struct {
		ledgerFilename      string
		metadataFilename    string
		accountBalanceTable string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			args: args{
				ledgerFilename:      ljr,
				metadataFilename:    meta,
				accountBalanceTable: accts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RecoverBlockchainMetadata(tt.args.ledgerFilename, tt.args.metadataFilename, tt.args.accountBalanceTable); err != nil {
				t.Errorf("RecoverBlockchainMetadata() error = %v", err)
			}
			firstBlockSerialized, err := blockchain.GetBlockByHeight(1, ljr, meta)
			if err != nil {
				t.Errorf("failed to get firstBlock block")
			}
			firstBlockDeserialized := block.Deserialize(firstBlockSerialized)
			if !reflect.DeepEqual(firstBlockDeserialized, firstBlock) {
				t.Errorf("first blocks do not match")
			}

			dbc, _ := sql.Open("sqlite3", accts)
			defer func() { // not sure if this defer will happen before the others, is it stack based?
				if err := dbc.Close(); err != nil {
					t.Errorf("Failed to close database: %s", err)
				}
			}()
			for i, key := range somePVKeys {
				someKeyPKhsh := block.HashSHA256(keys.EncodePublicKey(&key.PublicKey))
				var balance uint64
				var nonce uint64
				queryStr := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash=\"%s\"", hex.EncodeToString(someKeyPKhsh))
				row, err := dbc.Query(queryStr)
				if err != nil {
					t.Errorf("Failed to acquire row from table")
				}
				if row.Next() {
					err = row.Scan(&balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan row: %s", err)
					}
					switch i {
					case 0: // first contract
						if balance != 5 { // 10 - 5
							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
						}
						if nonce != 1 {
							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
						}
						break
					case 1: // second contract
						if balance != 13 { // 10 + 5 - 7 + 5
							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
						}
						if nonce != 3 {
							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
						}
						break
					default: // third contract
						if balance != 12 { // 10 + 7 - 5
							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
						}
						if nonce != 2 {
							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
						}
						break
					}
				} else {
					t.Errorf("Key not found in table: %v", someKeyPKhsh)
				}
				row.Close()
			}
		})
	}
}

// func TestRecoverBlockchainMetadataInitial(t *testing.T) {
// 	t.Errorf("ErroRRRR!")
// }

// func TestRecoverBlockchainMetadataInitial(t *testing.T) {
// 	blockchain := "testBlockchain.dat"
// 	table := "table.db"
// 	dbName := "accountBalanceTable.tab"
// 	dbc, _ := sql.Open("sqlite3", dbName)
// 	defer func() {
// 		err := dbc.Close()
// 		if err != nil {
// 			t.Errorf("Failed to remove database: %s", err)
// 		}
// 		err = os.Remove(dbName)
// 		if err != nil {
// 			t.Errorf("Failed to remove database: %s", err)
// 		}
// 	}()
// 	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
// 	statement.Exec()
// 	setUp(blockchain, table)
// 	defer tearDown(blockchain, table)
// 	// Create a bunch of blocks
// 	block0 := block.Block{
// 		Version:        1,
// 		Height:         0,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashSHA256([]byte{'0'}),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x', 'o', 'x', 'o'})},
// 	}
// 	block0.DataLen = uint16(len(block0.Data))
// 	block1 := block.Block{
// 		Version:        1,
// 		Height:         1,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashBlock(block0),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'x', 'y', 'z'})},
// 	}
// 	block1.DataLen = uint16(len(block1.Data))
// 	block2 := block.Block{
// 		Version:        1,
// 		Height:         2,
// 		Timestamp:      time.Now().UnixNano(),
// 		PreviousHash:   block.HashBlock(block1),
// 		MerkleRootHash: block.HashSHA256([]byte{'1'}),
// 		Data:           [][]byte{block.HashSHA256([]byte{'a', 'b', 'c'})},
// 	}
// 	block2.DataLen = uint16(len(block2.Data))
// 	// Add all the blocks
// 	err := AddBlock(block0, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to add block0.")
// 	}
// 	err = AddBlock(block1, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to add block1.")
// 	}
// 	err = AddBlock(block2, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to add block2.")
// 	}
// 	os.Remove(table)
// 	_, err = os.Stat(table)
// 	if err == nil {
// 		t.Errorf("table still exists")
// 	}
// 	err = RecoverBlockchainMetadata(blockchain, table)

// 	// Block 0 by height
// 	actualBlock0, err := GetBlockByHeight(0, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 0 by height).")
// 	}
// 	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
// 		t.Errorf("Blocks do not match (block 0 by height)")
// 	}

// 	// Block 1 by height
// 	actualBlock1, err := GetBlockByHeight(1, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 1 by height).")
// 	}
// 	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
// 		t.Errorf("Blocks do not match (block 1 by height)")
// 	}

// 	// Block 2
// 	actualBlock2, err := GetBlockByHeight(2, blockchain, table)
// 	if err != nil {
// 		t.Errorf("Failed to extract block (block 2 by height).")
// 	}
// 	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
// 		t.Errorf("Blocks do not match (block 2 by height)")
// 	}
// }
