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

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/genesis"
	"github.com/SIGBlockchain/project_aurum/internal/accountinfo"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

var removeFiles = true

func TestRunServer(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPublicKeyHash := hashing.New(publickey.Encode(&recipientPrivateKey.PublicKey))
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
	defer func() {
		if removeFiles {
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
	recipientPublicKeyHash := hashing.New(publickey.Encode(&recipientPrivateKey.PublicKey))
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

	youngestBlock, err := blockchain.GetYoungestBlock(ledger, metadataTable)
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

func TestResponseToAccountInfoRequest(t *testing.T) {
	if err := client.SetupWallet(); err != nil {
		t.Errorf("failed to setup wallet:\n%s", err.Error())
	}
	defer func() {
		if err := os.Remove("aurum_wallet.json"); err != nil {
			t.Errorf("failed to remove aurum_wallet.json:\n%s", err.Error())
		}
	}()
	dbName := constants.AccountsTable
	dbc, _ := sql.Open("sqlite3", dbName)
	defer func() {
		err := dbc.Close()
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
		err = os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	statement, _ := dbc.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	statement.Exec()
	walletAddress, err := client.GetWalletAddress()
	// t.Logf("Wallet address: %v", walletAddress)
	if err != nil {
		t.Errorf("failed to retrieve wallet address:\n%s", err.Error())
	}
	ln, err := net.Listen("tcp", "localhost:10500")
	if err != nil {
		t.Errorf("failed to start server:\n%s", err.Error())
	}
	byteChan := make(chan []byte)
	debug := false

	go RunServer(ln, byteChan, debug)

	// Request
	var requestInfoMessage []byte
	requestInfoMessage = append(requestInfoMessage, SecretBytes...)
	requestInfoMessage = append(requestInfoMessage, 2)
	requestInfoMessage = append(requestInfoMessage, walletAddress...)
	conn, err := net.Dial("tcp", "localhost:10500")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	// t.Logf("Sending message: %v", requestInfoMessage)
	if _, err := conn.Write(requestInfoMessage); err != nil {
		t.Errorf("failed to send request info message:\n%s", err.Error())
	}
	buf := make([]byte, 1024)
	nRead, err := conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if !bytes.Equal(buf[:nRead], []byte("Thank you.\n")) {
		t.Errorf("expected different response: %v != %v", string(buf[:nRead]), string([]byte("Thank you\n")))
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if buf[8] != 1 {
		t.Errorf("failed to get errored response from producer")
	}
	conn.Close()

	// Check for successful insertion
	if err := accountstable.InsertAccountIntoAccountBalanceTable(dbc, walletAddress, 1000); err != nil {
		t.Errorf("failed to insert sender account")
	}
	_, err = accountstable.GetAccountInfo(walletAddress)
	if err != nil {
		t.Errorf("failed to retrieve account info:\n%s", err.Error())
	}
	// t.Logf("account info: %v", accInfo)
	dbc.Close()

	// New request
	conn, err = net.Dial("tcp", "localhost:10500")
	if err != nil {
		t.Errorf("failed to connect to server:\n%s", err.Error())
	}
	// t.Logf("Sending message: %v", requestInfoMessage)
	if _, err := conn.Write(requestInfoMessage); err != nil {
		t.Errorf("failed to send request info message:\n%s", err.Error())
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}
	if !bytes.Equal(buf[:nRead], []byte("Thank you.\n")) {
		t.Errorf("expected different response: %v != %v", string(buf[:nRead]), string([]byte("Thank you\n")))
	}
	buf = make([]byte, 1024)
	nRead, err = conn.Read(buf)
	if err != nil {
		t.Errorf("failed to read from socket:\n%s", err.Error())
	}

	if buf[8] != 0 {
		t.Errorf("failed to get success response from producer: %d != %d", buf[8], 0)
	}
	var accInfo accountinfo.AccountInfo
	if err := accInfo.Deserialize(buf[9:nRead]); err != nil {
		t.Errorf("failed to deserialize account info:\n%s", err.Error())
	}
}

func TestData_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := hashing.New(publickey.Encode(&senderPrivateKey.PublicKey))
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
	spkh := hashing.New(publickey.Encode(&senderPrivateKey.PublicKey))
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

// func TestRecoverBlockchainMetadata(t *testing.T) {
// 	var ljr = "blockchain.dat"
// 	var meta = constants.MetadataTable
// 	var accts = constants.AccountsTable

// 	if file, err := os.Create(ljr); err != nil {
// 		t.Errorf("Failed to create file.")
// 	} else {
// 		file.Close()
// 	}
// 	defer func() {
// 		if err := os.Remove(ljr); err != nil {
// 			t.Errorf("failed to remove blockchain file")
// 		}
// 	}()

// 	if conn, err := sql.Open("sqlite3", meta); err != nil {
// 		t.Errorf("failed to create metadata file")
// 	} else {
// 		statement, _ := conn.Prepare(sqlstatements.CREATE_METADATA_TABLE)
// 		statement.Exec()
// 		conn.Close()
// 	}
// 	defer func() {
// 		if err := os.Remove(meta); err != nil {
// 			t.Errorf("failed to remove metadata file")
// 		}
// 	}()

// 	if conn, err := sql.Open("sqlite3", accts); err != nil {
// 		t.Errorf("failed to create accounts file")
// 	} else {
// 		statement, _ := conn.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
// 		statement.Exec()
// 		conn.Close()
// 	}

// 	defer func() {
// 		if err := os.Remove(accts); err != nil {
// 			t.Errorf("failed to remove accounts file" + err.Error())
// 		}
// 	}()

// 	var pkhashes [][]byte
// 	for i := 0; i < 100; i++ {
// 		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
// 		pkhashes = append(pkhashes, someKeyPKHash)
// 	}
// 	genny, _ := genesis.BringOnTheGenesis(pkhashes, 1000)
// 	if err := genesis.Airdrop(ljr, meta, constants.AccountsTable, genny); err != nil {
// 		t.Errorf("airdrop failed")
// 	}

// 	type args struct {
// 		ledgerFilename      string
// 		metadataFilename    string
// 		accountBalanceTable string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			args: args{
// 				ledgerFilename:      ljr,
// 				metadataFilename:    meta,
// 				accountBalanceTable: accts,
// 			},
// 		},
// 	}
// 	var blockchainHeightIdx = 0
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := RecoverBlockchainMetadata(tt.args.ledgerFilename, tt.args.metadataFilename, tt.args.accountBalanceTable); (err != nil) != tt.wantErr {
// 				t.Errorf("RecoverBlockchainMetadata() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			blockchainGenesisBlockSerialized, err := blockchain.GetBlockByHeight(blockchainHeightIdx, ljr, meta)
// 			if err != nil {
// 				t.Errorf("failed to get genesis block")
// 			}
// 			blockchainGenesisBlockDeserialized := block.Deserialize(blockchainGenesisBlockSerialized)
// 			if !reflect.DeepEqual(blockchainGenesisBlockDeserialized, genny) {
// 				t.Errorf("genesis blocks do not match")
// 			}
// 			dbc, _ := sql.Open("sqlite3", accts)
// 			defer func() { // not sure if this defer will happen before the others, is it stack based?
// 				if err := dbc.Close(); err != nil {
// 					t.Errorf("Failed to close database: %s", err)
// 				}
// 			}()
// 			for _, hsh := range pkhashes {
// 				var pkhash string
// 				var balance uint64
// 				var nonce uint64
// 				foundKey := false
// 				rows, err := dbc.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
// 				if err != nil {
// 					t.Errorf("Failed to acquire rows from table")
// 				}
// 				for rows.Next() {
// 					err = rows.Scan(&pkhash, &balance, &nonce)
// 					if err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, err := hex.DecodeString(pkhash)
// 					if err != nil {
// 						t.Errorf("failed to decode public key hash")
// 					}
// 					if bytes.Equal(hsh, decodedPkhash) {
// 						foundKey = true
// 						if balance != 10 {
// 							t.Errorf("wrong balance on key: %v", hsh)
// 						}
// 						if nonce != 0 {
// 							t.Errorf("wrong nonce on key: %v", hsh)
// 						}
// 					}
// 				}
// 				if !foundKey {
// 					t.Errorf("Key not found in table: %v", hsh)
// 				}
// 			}
// 		})
// 		blockchainHeightIdx++
// 	}
// }

// func TestRecoverBlockchainMetadata_TwoBlocks(t *testing.T) {
// 	var ljr = "blockchain.dat"
// 	var meta = constants.MetadataTable
// 	var accts = constants.AccountsTable

// 	// Create ledger file and the two tables
// 	file, err := os.Create(ljr)
// 	if err != nil {
// 		t.Errorf("Failed to create file.")
// 	} else {
// 		file.Close()
// 	}
// 	metaDB, err := sql.Open("sqlite3", meta)
// 	if err != nil {
// 		t.Errorf("failed to create metadata file")
// 	} else {
// 		metaDB.Exec(sqlstatements.CREATE_METADATA_TABLE)
// 		metaDB.Close()
// 	}
// 	acctsDB, err := sql.Open("sqlite3", accts)
// 	if err != nil {
// 		t.Errorf("failed to create accounts file")
// 	} else {
// 		acctsDB.Exec(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
// 	}

// 	defer func() {
// 		if err := os.Remove(ljr); err != nil {
// 			t.Errorf("failed to remove blockchain file")
// 		}
// 		if err := os.Remove(meta); err != nil {
// 			t.Errorf("failed to remove metadata file")
// 		}
// 		if err := os.Remove(accts); err != nil {
// 			t.Errorf("failed to remove accounts file")
// 		}
// 	}()

// 	var pkhashes [][]byte
// 	somePVKeys := make([]*ecdsa.PrivateKey, 3) // Grab 3 private keys for creating contracts
// 	for i := 0; i < 100; i++ {
// 		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 		someKeyPKHash := hashing.New(publickey.Encode(&someKey.PublicKey))
// 		pkhashes = append(pkhashes, someKeyPKHash)
// 		if i < 3 {
// 			somePVKeys[i] = someKey
// 		}
// 	}
// 	genesisBlk, _ := genesis.BringOnTheGenesis(pkhashes, 1000)
// 	if err := genesis.Airdrop(ljr, meta, accts, genesisBlk); err != nil {
// 		t.Errorf("airdrop failed")
// 	}
// 	// Insert pkhashes into account table for contract validation
// 	for i := 0; i < 100; i++ {
// 		err := accountstable.InsertAccountIntoAccountBalanceTable(acctsDB, pkhashes[i], 10)
// 		if err != nil {
// 			t.Errorf("Failed to insert pkhash (%v) into account table: %s", pkhashes[i], err.Error())
// 		}
// 	}

// 	// Create 3 contracts
// 	contrcts := make([]contracts.Contract, 3)

// 	// Contract 1
// 	recipPKHash := hashing.New(publickey.Encode(&(somePVKeys[1].PublicKey)))
// 	contract1, _ := contracts.New(1, somePVKeys[0], recipPKHash, 5, 1) // pkh1 to pkh2
// 	contract1.Sign(somePVKeys[0])
// 	err = validation.ValidateContract(contract1)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	senderPKHash := hashing.New(publickey.Encode(&(somePVKeys[0].PublicKey)))
// 	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table for further contracts

// 	// Contract 2
// 	recipPKHash = hashing.New(publickey.Encode(&(somePVKeys[2].PublicKey)))
// 	contract2, _ := contracts.New(1, somePVKeys[1], recipPKHash, 7, 2) // pkh2 to pkh3
// 	contract2.Sign(somePVKeys[1])
// 	err = validation.ValidateContract(contract2)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	senderPKHash = hashing.New(publickey.Encode(&somePVKeys[1].PublicKey))
// 	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 7) // update accts table for further contracts

// 	// Contract 3
// 	recipPKHash = hashing.New(publickey.Encode(&(somePVKeys[1].PublicKey)))
// 	contract3, _ := contracts.New(1, somePVKeys[2], recipPKHash, 5, 2) // pkh3 to pkh2
// 	contract3.Sign(somePVKeys[2])
// 	err = validation.ValidateContract(contract3)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	senderPKHash = hashing.New(publickey.Encode(&somePVKeys[2].PublicKey))
// 	accountstable.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table
// 	acctsDB.Close()

// 	contrcts[0] = *contract1
// 	contrcts[1] = *contract2
// 	contrcts[2] = *contract3

// 	firstBlock, err := block.New(1, 1, block.HashBlock(genesisBlk), contrcts)
// 	if err != nil {
// 		t.Errorf("failed to create first block")
// 	}
// 	err = blockchain.AddBlock(firstBlock, ljr, meta)
// 	if err != nil {
// 		t.Errorf("failed to add first block")
// 	}

// 	os.Remove(meta)
// 	os.Remove(accts)
// 	type args struct {
// 		ledgerFilename      string
// 		metadataFilename    string
// 		accountBalanceTable string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			args: args{
// 				ledgerFilename:      ljr,
// 				metadataFilename:    meta,
// 				accountBalanceTable: accts,
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := RecoverBlockchainMetadata(tt.args.ledgerFilename, tt.args.metadataFilename, tt.args.accountBalanceTable); err != nil {
// 				t.Errorf("RecoverBlockchainMetadata() error = %v", err)
// 			}
// 			firstBlockSerialized, err := blockchain.GetBlockByHeight(1, ljr, meta)
// 			if err != nil {
// 				t.Errorf("failed to get firstBlock block")
// 			}
// 			firstBlockDeserialized := block.Deserialize(firstBlockSerialized)
// 			if !reflect.DeepEqual(firstBlockDeserialized, firstBlock) {
// 				t.Errorf("first blocks do not match")
// 			}

// 			dbc, _ := sql.Open("sqlite3", accts)
// 			defer func() { // not sure if this defer will happen before the others, is it stack based?
// 				if err := dbc.Close(); err != nil {
// 					t.Errorf("Failed to close database: %s", err)
// 				}
// 			}()

// 			for i, key := range somePVKeys {
// 				someKeyPKhsh := hashing.New(publickey.Encode(&key.PublicKey))
// 				var balance uint64
// 				var nonce uint64
// 				queryStr := fmt.Sprintf(sqlstatements.GET_BALANCE_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, hex.EncodeToString(someKeyPKhsh))
// 				row, err := dbc.Query(queryStr)
// 				if err != nil {
// 					t.Errorf("Failed to acquire row from table")
// 				}
// 				if row.Next() {
// 					err = row.Scan(&balance, &nonce)
// 					if err != nil {
// 						t.Errorf("failed to scan row: %s", err)
// 					}
// 					switch i {
// 					case 0: // first contract
// 						if balance != 5 { // 10 - 5
// 							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
// 						}
// 						if nonce != 1 {
// 							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
// 						}
// 						break
// 					case 1: // second contract
// 						if balance != 13 { // 10 + 5 - 7 + 5
// 							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
// 						}
// 						if nonce != 3 {
// 							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
// 						}
// 						break
// 					default: // third contract
// 						if balance != 12 { // 10 + 7 - 5
// 							t.Errorf("wrong balance (%v) on key: %v", balance, someKeyPKhsh)
// 						}
// 						if nonce != 2 {
// 							t.Errorf("wrong nonce (%v) on key: %v", nonce, someKeyPKhsh)
// 						}
// 						break
// 					}
// 				} else {
// 					t.Errorf("Key not found in table: %v", someKeyPKhsh)
// 				}
// 				row.Close()
// 			}
// 		})
// 	}
// }
