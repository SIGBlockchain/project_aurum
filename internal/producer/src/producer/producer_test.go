package producer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/client/src/client"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

var removeFiles = true

// Test will fail in airplane mode, or just remove wireless connection.
func TestCheckConnectivity(t *testing.T) {
	err := CheckConnectivity()
	if err != nil {
		t.Errorf("Internet connection check failed.")
	}
}

// Tests a single connection
// func TestAcceptConnections(t *testing.T) {
// 	ln, _ := net.Listen("tcp", "localhost:10000")
// 	var buffer bytes.Buffer
// 	bp := BlockProducer{
// 		Server:        ln,
// 		NewConnection: make(chan net.Conn, 128),
// 		Logger:        log.New(&buffer, "LOG:", log.Ldate),
// 	}
// 	go bp.AcceptConnections()
// 	conn, err := net.Dial("tcp", ":10000")
// 	if err != nil {
// 		t.Errorf("Failed to connect to server")
// 	}
// 	contentsOfChannel := <-bp.NewConnection
// 	actual := contentsOfChannel.RemoteAddr().String()
// 	expected := conn.LocalAddr().String()
// 	if actual != expected {
// 		t.Errorf("Failed to store connection")
// 	}
// 	conn.Close()
// 	ln.Close()
// }

// Sends a message to the listener and checks to see if it echoes back
// func TestHandler(t *testing.T) {
// 	ln, _ := net.Listen("tcp", "localhost:10000")
// 	var buffer bytes.Buffer
// 	bp := BlockProducer{
// 		Server:        ln,
// 		NewConnection: make(chan net.Conn, 128),
// 		Logger:        log.New(&buffer, "LOG:", log.Ldate),
// 	}
// 	go bp.AcceptConnections()
// 	go bp.WorkLoop()
// 	conn, err := net.Dial("tcp", ":10000")
// 	if err != nil {
// 		t.Errorf("Failed to connect to server")
// 	}
// 	expected := []byte("This is a test.")
// 	conn.Write(expected)
// 	actual := make([]byte, len(expected))
// 	_, readErr := conn.Read(actual)
// 	if readErr != nil {
// 		t.Errorf("Failed to read from socket.")
// 	}
// 	if bytes.Equal(expected, actual) == false {
// 		t.Errorf("Message mismatch")
// 	}
// 	conn.Close()
// 	ln.Close()
// }

func TestRunServer(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
	contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.SignContract(senderPrivateKey)
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
				var contract accounts.Contract
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
	genesisHashes, err := ReadGenesisHashes()
	if err != nil {
		t.Errorf("failed to read genesis hashes:\n%s", err.Error())
	}
	genesisBlock, err := BringOnTheGenesis(genesisHashes, 1000)
	if err != nil {
		t.Errorf("failed to create genesis block:\n%s", err.Error())
	}
	if err := Airdrop(ledger, metadataTable, constants.AccountsTable, genesisBlock); err != nil {
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
	recipientPublicKeyHash := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
	contract, _ := accounts.MakeContract(1, senderPrivateKey, recipientPublicKeyHash, 1000, 1)
	contract.SignContract(senderPrivateKey)
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
	var compContract accounts.Contract
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
	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
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
	if err := accounts.InsertAccountIntoAccountBalanceTable(dbc, walletAddress, 1000); err != nil {
		t.Errorf("failed to insert sender account")
	}
	_, err = accounts.GetAccountInfo(walletAddress)
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
	var accInfo accounts.AccountInfo
	if err := accInfo.Deserialize(buf[9:nRead]); err != nil {
		t.Errorf("failed to deserialize account info:\n%s", err.Error())
	}
}

func TestData_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))
	initialContract, _ := accounts.MakeContract(1, nil, spkh, 1000, 0)
	tests := []struct {
		name string
		// d    *Data
		d *accounts.Contract
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
	spkh := block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))
	initialContract, _ := accounts.MakeContract(1, nil, spkh, 1000, 0)
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
		d       *accounts.Contract
		args    args
		wantErr bool
	}{
		{
			// d: &Data{},
			d: &accounts.Contract{},
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
	var datum []accounts.Contract
	for i := 0; i < 12; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		someAirdropContract, _ := accounts.MakeContract(1, nil, someKeyPKHash, 1000, 0)
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
		data         []accounts.Contract
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
				DataLen:        12,
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
	var datum []accounts.Contract
	for i := 0; i < 100; i++ {
		someKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
		pkhashes = append(pkhashes, someKeyPKHash)
		someAirdropContract, _ := accounts.MakeContract(1, nil, someKeyPKHash, 10, 0)
		datum = append(datum, *someAirdropContract)
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
				deserializedData := accounts.Contract{}
				err := deserializedData.Deserialize(got.Data[i])
				if err != nil {
					t.Errorf("failed to deserialize data")
				}
				deserializedContract := &accounts.Contract{}
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
		someKeyPKHash := block.HashSHA256(keys.EncodePublicKey(&someKey.PublicKey))
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

			rows, err := db.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
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
	if err := Airdrop(ljr, meta, constants.AccountsTable, genny); err != nil {
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
	if err := Airdrop(ljr, meta, accts, genesisBlk); err != nil {
		t.Errorf("airdrop failed")
	}
	// Insert pkhashes into account table for contract validation
	for i := 0; i < 100; i++ {
		err := accounts.InsertAccountIntoAccountBalanceTable(acctsDB, pkhashes[i], 10)
		if err != nil {
			t.Errorf("Failed to insert pkhash (%v) into account table: %s", pkhashes[i], err.Error())
		}
	}

	// Create 3 contracts
	contracts := make([]accounts.Contract, 3)

	// Contract 1
	recipPKHash := block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[1].PublicKey)))
	contract1, _ := accounts.MakeContract(1, somePVKeys[0], recipPKHash, 5, 1) // pkh1 to pkh2
	contract1.SignContract(somePVKeys[0])
	err = accounts.ValidateContract(contract1)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash := block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[0].PublicKey)))
	accounts.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table for further contracts

	// Contract 2
	recipPKHash = block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[2].PublicKey)))
	contract2, _ := accounts.MakeContract(1, somePVKeys[1], recipPKHash, 7, 2) // pkh2 to pkh3
	contract2.SignContract(somePVKeys[1])
	err = accounts.ValidateContract(contract2)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash = block.HashSHA256(keys.EncodePublicKey(&somePVKeys[1].PublicKey))
	accounts.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 7) // update accts table for further contracts

	// Contract 3
	recipPKHash = block.HashSHA256(keys.EncodePublicKey(&(somePVKeys[1].PublicKey)))
	contract3, _ := accounts.MakeContract(1, somePVKeys[2], recipPKHash, 5, 2) // pkh3 to pkh2
	contract3.SignContract(somePVKeys[2])
	err = accounts.ValidateContract(contract3)
	if err != nil {
		t.Error(err.Error())
	}
	senderPKHash = block.HashSHA256(keys.EncodePublicKey(&somePVKeys[2].PublicKey))
	accounts.ExchangeBetweenAccountsUpdateAccountBalanceTable(acctsDB, senderPKHash, recipPKHash, 5) // update accts table
	acctsDB.Close()

	contracts[0] = *contract1
	contracts[1] = *contract2
	contracts[2] = *contract3

	firstBlock, err := CreateBlock(1, 1, block.HashBlock(genesisBlk), contracts)
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

func TestGenerateNRandomKeys(t *testing.T) {
	type args struct {
		filename string
		n        uint32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test with N = 0",
			args: args{
				filename: "testfile.json",
				n:        0,
			},
			wantErr: true,
			// Error should say "must generate at least 1 key"
		},
		{
			name: "Test with N = 100",
			args: args{
				filename: "testfile.json",
				n:        100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateNRandomKeys(tt.args.filename, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNRandomKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, err := os.Stat(tt.args.filename); os.IsNotExist(err) {
				t.Errorf("Test file for keys not detected: %s", err)
			}
			if tt.args.n > 0 {
				jsonFile, err := os.Open(tt.args.filename)
				if err != nil {
					t.Errorf("Failed to open json file: %s", err)
				}
				defer jsonFile.Close()
				var keys []string
				byteKeys, err := ioutil.ReadAll(jsonFile)
				if err != nil {
					t.Errorf("Failed to read in keys from json file: %s", err)
				}
				err = json.Unmarshal(byteKeys, &keys)
				if err != nil {
					t.Errorf("Failed to unmarshall keys because: %s", err)
				}
				if uint32(len(keys)) != tt.args.n {
					t.Errorf("Number of private keys do not match n: %d", len(keys))
				}

			}
		})
	}
	if err := os.Remove("testfile.json"); err != nil {
		t.Errorf("Failed to remove file: %s because: %s", "testfile.json", err)
	}
}

func TestGenesisReadsAppropriately(t *testing.T) {
	var testGenesisHash = "8db5d191bf333f96179c5f2ec7acd20a8c01378a1af120e2f2ded3672896931a"
	genHashfile, _ := os.Create(genesisHashFile)
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
	var ctc accounts.Contract
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

	rows, err := db.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
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
