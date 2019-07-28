package genesis

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

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

var removeFiles = true

func TestBringOnTheGenesis(t *testing.T) {
	var pkhashes [][]byte
	var datum []contracts.Contract
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
		someAirdropContract, _ := contracts.New(1, nil, someKeyPKHash, 10, 0)
		datum = append(datum, *someAirdropContract)
	}
	genny, _ := block.New(1, 0, make([]byte, 32), datum)
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
				deserializedData := contracts.Contract{}
				err := deserializedData.Deserialize(got.Data[i])
				if err != nil {
					t.Errorf("failed to deserialize data")
				}
				deserializedContract := &contracts.Contract{}
				serializedDataBdy, _ := deserializedData.Serialize()
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
		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
	}
	genny, _ := BringOnTheGenesis(pkhashes, 1000)
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

func TestReadGenesisHashes(t *testing.T) {
	GenerateGenesisHashFile(50)

	tests := []struct {
		name    string
		want    [][]byte
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadGenesisHashes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadGenesisHashes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 50 {
				t.Errorf("wrong count on number of hashes")
			}
		})
	}
}

func TestGenesisReadsAppropriately(t *testing.T) {
	var testGenesisHash = "8db5d191bf333f96179c5f2ec7acd20a8c01378a1af120e2f2ded3672896931a"
	genHashfile, _ := os.Create(constants.GenesisHashFile)
	genHashfile.WriteString(testGenesisHash + "\n") // from GenerateGenesisHashFile
	genHashfile.Close()
	defer func() {
		if err := os.Remove(constants.GenesisAddresses); err != nil {
			t.Errorf("failed to remove genesis hash file: %v", err)
		}
	}()
	hashSlice, err := ReadGenesisHashes()
	if err != nil {
		t.Errorf("failed to read from hash file: %v", hashSlice)
	}
	if hex.EncodeToString(hashSlice[0]) != testGenesisHash {
		t.Errorf("hash from genesis_hashes.txt don't match: %s != %s", hex.EncodeToString(hashSlice[0]), testGenesisHash)
	}
	genesisBlock, err := BringOnTheGenesis(hashSlice, 1000)
	if err != nil {
		t.Errorf("failed to create genesis block: %v", err)
	}
	serializedCtc := genesisBlock.Data[0]
	var ctc contracts.Contract
	err = ctc.Deserialize(serializedCtc)
	if err != nil {
		t.Errorf("failed to deserialize block: %v", err)
	}
	recipient := hex.EncodeToString(ctc.RecipPubKeyHash)
	if recipient != testGenesisHash {
		t.Errorf("hashes don't match: %s != %s", recipient, testGenesisHash)
	}

	//airdrop
	err = Airdrop("blockchain.dat", constants.MetadataTable, constants.AccountsTable, genesisBlock)
	if err != nil {
		t.Errorf("failed to airdrop genesis block: %s", err.Error())
	}
	defer func() {
		if err := os.Remove("blockchain.dat"); err != nil {
			t.Errorf("failed to remove blockchain.dat:\n%s", err.Error())
		}
		if err := os.Remove(constants.MetadataTable); err != nil {
			t.Errorf("failed to remove metadatata.tab:\n%s", err.Error())
		}
		if err := os.Remove(constants.AccountsTable); err != nil {
			t.Errorf("failed to remove accounts.db:\n%s", err.Error())
		}
	}()

	db, err := sql.Open("sqlite3", constants.AccountsTable)
	if err != nil {
		t.Errorf("failed to open accounts table")
	}
	defer db.Close()

	rows, err := db.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
	if err != nil {
		t.Errorf("failed to create rows for queries")
	}
	defer rows.Close()

	var pkhash string
	var balance int
	var nonce int
	for rows.Next() {
		rows.Scan(&pkhash, &balance, &nonce)
		if pkhash != testGenesisHash {
			t.Errorf("hash in accounts table doesn't match: %s != %s\n", pkhash, testGenesisHash)
		}
		if balance != 1000 {
			t.Errorf("balance in accounts table doesn't match: %v != %v\n", balance, 1000)
		}
		if nonce != 0 {
			t.Errorf("nonce in accounts table doesn't match: %v != %v\n", nonce, 0)
		}
	}
}
