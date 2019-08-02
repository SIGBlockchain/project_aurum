package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	block "github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/genesis"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
	"github.com/SIGBlockchain/project_aurum/internal/validation"
)

func setUp(filename string, database string) *sql.DB {
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		panic("Failed to open database")
	}
	statement, _ := conn.Prepare(sqlstatements.CREATE_METADATA_TABLE)
	statement.Exec()

	file, err := os.Create(filename)
	if err != nil {
		panic("Failed to create file.")
	}
	file.Close()

	return conn
}

func tearDown(metadata *sql.DB, filename string, database string) {
	metadata.Close()
	os.Remove(filename)
	os.Remove(database)
}

func addBlockHelper(b block.Block, filename string, metadata *sql.DB) error {
	f, _ := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	defer f.Close()
	return AddBlock(b, f, metadata)
}

func TestPhaseOneAddBlock(t *testing.T) {

	// Create a block
	b := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	b.DataLen = uint16(len(b.Data))

	// Setup
	metadata := setUp("testFile.txt", "testDatabase.db")
	defer tearDown(metadata, "testFile.txt", "testDatabase.db")

	err := addBlockHelper(b, "testFile.txt", metadata)
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestPhaseTwoGetBlockByHeight(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))

	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block. " + err.Error())
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByHeight(0, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block." + err.Error())
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}

func TestPhaseTwoGetBlockPosition(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByPosition(0, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}

func TestPhaseTwoGetBlockByHash(t *testing.T) {
	// Create a block
	expectedBlock := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x'})},
	}
	expectedBlock.DataLen = uint16(len(expectedBlock.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add the block
	err := addBlockHelper(expectedBlock, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	actualBlock, err := GetBlockByHash(block.HashBlock(expectedBlock), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block.")
	}
	if bytes.Equal(expectedBlock.Serialize(), actualBlock) == false {
		t.Errorf("Blocks do not match")
	}
}

func TestPhaseTwoMultiple(t *testing.T) {
	// Create a bunch of blocks
	block0 := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x', 'o', 'x', 'o'})},
	}
	block0.DataLen = uint16(len(block0.Data))
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block0),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'x', 'y', 'z'})},
	}
	block1.DataLen = uint16(len(block1.Data))
	block2 := block.Block{
		Version:        1,
		Height:         2,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   block.HashBlock(block1),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte{'a', 'b', 'c'})},
	}
	block2.DataLen = uint16(len(block2.Data))
	// Setup
	metadata := setUp("testBlockchain.dat", "testDatabase.db")
	defer tearDown(metadata, "testBlockchain.dat", "testDatabase.db")

	// Add all the blocks
	err := addBlockHelper(block0, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block0.")
	}
	err = addBlockHelper(block1, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block1.")
	}
	err = addBlockHelper(block2, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block2.")
	}

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	defer file.Close()
	// Extract all three blocks
	// Block 0 by hash
	actualBlock0, err := GetBlockByHash(block.HashBlock(block0), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by hash).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by hash)")
	}

	// Block 0 by height
	actualBlock0, err = GetBlockByHeight(0, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 0 by height).")
	}
	if bytes.Equal(block0.Serialize(), actualBlock0) == false {
		t.Errorf("Blocks do not match (block 0 by height)")
	}

	// Block 1 by hash
	actualBlock1, err := GetBlockByHash(block.HashBlock(block1), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by hash).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by hash)")
	}

	// Block 1 by height
	actualBlock1, err = GetBlockByHeight(1, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 1 by height).")
	}
	if bytes.Equal(block1.Serialize(), actualBlock1) == false {
		t.Errorf("Blocks do not match (block 1 by height)")
	}

	// Block 2
	actualBlock2, err := GetBlockByHash(block.HashBlock(block2), file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 2 by hash).")
	}
	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
		t.Errorf("Blocks do not match (block 2 by hash)")
	}

	// Block 2
	actualBlock2, err = GetBlockByHeight(2, file, metadata)
	if err != nil {
		t.Errorf("Failed to extract block (block 2 by height).")
	}
	if bytes.Equal(block2.Serialize(), actualBlock2) == false {
		t.Errorf("Blocks do not match (block 2 by height)")
	}
}

func TestGetYoungestBlockAndBlockHeader(t *testing.T) {
	blockchain := "testBlockchain.dat"
	table := "testTable.dat"
	metadata := setUp(blockchain, table)
	defer tearDown(metadata, blockchain, table)

	file, err := os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	_, err = GetYoungestBlock(file, metadata)
	if err == nil {
		t.Errorf("Should return error if blockchain is empty")
	}
	file.Close()
	block0 := block.Block{
		Version:        1,
		Height:         0,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte("xoxo"))},
	}
	block0.DataLen = uint16(len(block0.Data))
	err = addBlockHelper(block0, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	file, err = os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	actualBlock0, err := GetYoungestBlock(file, metadata)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock0, block0) {
		t.Errorf("Blocks do not match")
	}
	file.Close()
	block1 := block.Block{
		Version:        1,
		Height:         1,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
		Data:           [][]byte{hashing.New([]byte("xoxo"))},
	}
	block1.DataLen = uint16(len(block1.Data))
	block1Header := block.BlockHeader{
		Version:        1,
		Height:         1,
		Timestamp:      block1.Timestamp,
		PreviousHash:   hashing.New([]byte{'0'}),
		MerkleRootHash: hashing.New([]byte{'1'}),
	}
	err = addBlockHelper(block1, "testBlockchain.dat", metadata)
	if err != nil {
		t.Errorf("Failed to add block")
	}
	file, err = os.OpenFile("testBlockchain.dat", os.O_RDONLY, 0644)
	actualBlock1Header, err := GetYoungestBlockHeader(file, metadata)
	if err != nil {
		t.Errorf("Error extracting youngest block")
	}
	if !cmp.Equal(actualBlock1Header, block1Header) {
		t.Errorf("Blocks Headers do not match")
	}
	file.Close()
}

func TestRecoverBlockchainMetadata(t *testing.T) {
	var ljr = "blockchain.dat"
	var meta = constants.MetadataTable
	var accts = constants.AccountsTable

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

	metadataConn, err := sql.Open("sqlite3", meta)
	if err != nil {
		t.Errorf("failed to create metadata file")
	} else {
		statement, _ := metadataConn.Prepare(sqlstatements.CREATE_METADATA_TABLE)
		statement.Exec()
	}
	defer func() {
		metadataConn.Close()
		if err := os.Remove(meta); err != nil {
			t.Errorf("failed to remove metadata file")
		}
	}()

	if conn, err := sql.Open("sqlite3", accts); err != nil {
		t.Errorf("failed to create accounts file")
	} else {
		statement, _ := conn.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
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
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := genesis.BringOnTheGenesis(pkhashes, 1000)
	if err := Airdrop(ljr, meta, constants.AccountsTable, genny); err != nil {
		t.Errorf("airdrop failed")
	}

	ledgerFile, _ := os.OpenFile(ljr, os.O_RDONLY, 0644)
	defer ledgerFile.Close()

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
			blockchainGenesisBlockSerialized, err := GetBlockByHeight(blockchainHeightIdx, ledgerFile, metadataConn)
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
				rows, err := dbc.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
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
	var meta = constants.MetadataTable
	var accts = constants.AccountsTable

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
		metaDB.Exec(sqlstatements.CREATE_METADATA_TABLE)
	}
	acctsDB, err := sql.Open("sqlite3", accts)
	if err != nil {
		t.Errorf("failed to create accounts file")
	} else {
		acctsDB.Exec(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
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
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
		if i < 3 {
			somePVKeys[i] = someKey
		}
	}
	genesisBlk, _ := genesis.BringOnTheGenesis(pkhashes, 1000)
	if err := Airdrop(ljr, meta, accts, genesisBlk); err != nil {
		t.Errorf("airdrop failed")
	}
	// Insert pkhashes into account table for contract validation
	for i := 0; i < 100; i++ {
		err := accountstable.InsertAccountIntoAccountBalanceTable(acctsDB, pkhashes[i], 10)
		if err != nil {
			t.Errorf("Failed to insert pkhash (%v) into account table: %s", pkhashes[i], err.Error())
		}
	}

	// Create 3 contracts
	contrcts := make([]contracts.Contract, 3)

	// Contract 1
	recipPKHash := hashing.New(publickey.Encode(&(somePVKeys[1].PublicKey)))
	contract1, _ := contracts.New(1, somePVKeys[0], recipPKHash, 5, 1) // pkh1 to pkh2
	contract1.Sign(somePVKeys[0])
	err = validation.ValidateContract(acctsDB, contract1)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash := hashing.New(publickey.Encode(&(somePVKeys[0].PublicKey)))
	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table for further contracts

	// Contract 2
	recipPKHash = hashing.New(publickey.Encode(&(somePVKeys[2].PublicKey)))
	contract2, _ := contracts.New(1, somePVKeys[1], recipPKHash, 7, 2) // pkh2 to pkh3
	contract2.Sign(somePVKeys[1])
	err = validation.ValidateContract(acctsDB, contract2)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash = hashing.New(publickey.Encode(&somePVKeys[1].PublicKey))
	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 7) // update accts table for further contracts

	// Contract 3
	recipPKHash = hashing.New(publickey.Encode(&(somePVKeys[1].PublicKey)))
	contract3, _ := contracts.New(1, somePVKeys[2], recipPKHash, 5, 2) // pkh3 to pkh2
	contract3.Sign(somePVKeys[2])
	err = validation.ValidateContract(acctsDB, contract3)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash = hashing.New(publickey.Encode(&somePVKeys[2].PublicKey))
	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table
	acctsDB.Close()

	contrcts[0] = *contract1
	contrcts[1] = *contract2
	contrcts[2] = *contract3

	firstBlock, err := block.New(1, 1, block.HashBlock(genesisBlk), contrcts)
	if err != nil {
		t.Errorf("failed to create first block")
	}
	ledgerFile, _ := os.OpenFile(ljr, os.O_APPEND|os.O_WRONLY, 0644)
	err = AddBlock(firstBlock, ledgerFile, metaDB)
	if err != nil {
		t.Errorf("failed to add first block")
	}
	ledgerFile.Close()
	metaDB.Close()

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
			ledgerFile, _ := os.OpenFile(tt.args.ledgerFilename, os.O_RDONLY, 0644)
			metadataConn, _ := sql.Open("sqlite3", tt.args.metadataFilename)
			firstBlockSerialized, err := GetBlockByHeight(1, ledgerFile, metadataConn)
			if err != nil {
				t.Errorf("failed to get firstBlock block")
			}
			ledgerFile.Close()
			metadataConn.Close()
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
				someKeyPKhsh := hashing.New(publickey.Encode(&key.PublicKey))
				var balance uint64
				var nonce uint64
				row, err := dbc.Query(sqlstatements.GET_BALANCE_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, hex.EncodeToString(someKeyPKhsh))
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

func TestAirdrop(t *testing.T) {
	var pkhashes [][]byte
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := genesis.BringOnTheGenesis(pkhashes, 1000)
	type args struct {
		blockchain   string
		metadata     string
		accounts     string
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
				metadata:     constants.MetadataTable,
				accounts:     constants.AccountsTable,
				genesisBlock: genny,
			},
		},
	}
	for _, tt := range tests {
		defer func() {
			os.Remove(tt.args.metadata)
			os.Remove(tt.args.blockchain)
			os.Remove(tt.args.accounts)
		}()
		t.Run(tt.name, func(t *testing.T) {
			if err := Airdrop(tt.args.blockchain, tt.args.metadata, tt.args.accounts, tt.args.genesisBlock); (err != nil) != tt.wantErr {
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

			db, err := sql.Open("sqlite3", tt.args.accounts)
			if err != nil {
				t.Errorf("Failed to open accounts table: " + err.Error())
			}
			defer db.Close()

			rows, err := db.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
			if err != nil {
				t.Errorf("failed to create rows for queries")
			}
			defer rows.Close()

			var pkhCount int
			var pkhStr string
			var balance int
			var nonce int
			for rows.Next() {
				rows.Scan(&pkhStr, &balance, &nonce)
				pkhash, _ := hex.DecodeString(pkhStr)
				if !bytes.Equal(pkhash, pkhashes[pkhCount]) {
					t.Errorf("hashes don't match: %v != %v\n", pkhash, pkhashes[pkhCount])
				}
				if balance != 10 {
					t.Errorf("balance does not match: %v != %v\n", balance, 10)
				}
				if nonce != 0 {
					t.Errorf("nonce does not match: %v != %v\n", nonce, 0)
				}
				pkhCount++
			}
		})
	}
}
