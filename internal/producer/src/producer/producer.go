// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/pkg/keys"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"
)

// This stores connection information for the producer
type BlockProducer struct {
	Server        net.Listener
	NewConnection chan net.Conn
	Logger        *log.Logger
	// Add ledger name, metadata name, and contract table name
	// Slice of Contracts representing contract pool
}

// This checks if the producer is connected to the internet
func CheckConnectivity() error {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		return errors.New("connectivity check failed")
	}
	conn.Close()
	return nil
}

// This will accept any incoming connections
func (bp *BlockProducer) AcceptConnections() {
	for {
		conn, err := bp.Server.Accept()
		if err != nil {
			return
		}
		bp.NewConnection <- conn
		bp.Logger.Printf("%s connected\n", conn.RemoteAddr())
	}
}

// Should contain version, payload type, payload size
type Header struct{}

// Messages have headers and payloads
// Payloads should correspond to message type
type Message struct{}

// Handles incoming connections, accepting _ of at most 1024 bytes
func (bp *BlockProducer) Handle(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}
	bp.Logger.Printf("%s sent: %s", conn.RemoteAddr(), buf)
	/*
		Check if keybytes are present.
		If they aren't, close the connection, add IP to greylist

		Check type of message (Balance or Contract)

		If Balance, query the database with the public key
		If no public key exists, send message with `invalid public key`
		Otherwise send balance back

		If Contract, validate contract first
		If validation fails, send message with `invalid contract`
		Otherwise, add contract to contract pool and send verification message
	*/
	conn.Write(buf)
	conn.Close()
}

// The main work loop which handles communication, block production, and ledger maintenance
func (bp *BlockProducer) WorkLoop() {
	// Creates signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	// TODO: Extract last block blockhash, timestamp
	// Initialize contract pool
	for {
		select {
		case conn := <-bp.NewConnection:
			go bp.Handle(conn) // TODO: remove `go`
		// If an interrupt signal is encountered, exit
		case <-signalCh:
			// If loop is exited properly, interrupt signal had been recieved
			fmt.Print("\r")
			bp.Logger.Println("Interrupt signal encountered, program terminating.")
			return
		default:
			/*
				TODO:
				Check to see if it's time to make a block
				Block interval = (timeNow - timeSinceLastBlock)
				If it is time, make the block from the contract pool
				(Merkle Root, add block)
				Reset last block metadata
			*/
		}
	}
}

type DataHeader struct {
	Version uint16 // Version denotes how the Data piece is structured
	Type    uint16 // Identifies what the type of the Data Body is
}

type DataElem interface {
	Serialize() ([]byte, error) // Call serialize function of DataElem
	Deserialize([]byte) error
}

type Data struct {
	Hdr DataHeader
	Bdy DataElem
}

func (d *Data) Serialize() ([]byte, error) {
	serializedData := make([]byte, 4) // 2 + 2 bytes for Dataheader version and type
	binary.LittleEndian.PutUint16(serializedData[:2], d.Hdr.Version)
	binary.LittleEndian.PutUint16(serializedData[2:], d.Hdr.Type)

	dataBdy, err := d.Bdy.Serialize() // serialize data body
	if err != nil {
		return nil, errors.New("Failed to serialize data body")
	}
	serializedData = append(serializedData, dataBdy...)
	return serializedData, nil
}

func (d *Data) Deserialize(serializedData []byte) error {
	d.Hdr.Version = binary.LittleEndian.Uint16(serializedData[:2]) // data version
	d.Hdr.Type = binary.LittleEndian.Uint16(serializedData[2:4])   // data type

	d.Bdy = &accounts.Contract{}
	err := d.Bdy.Deserialize(serializedData[4:]) // data body
	if err != nil {
		return errors.New("Failed to deserialize data: " + err.Error())
	}

	return nil
}

func CreateBlock(version uint16, height uint64, previousHash []byte, data []Data) (block.Block, error) {
	var serializedDatum [][]byte // A series of serialized data for Merkle root hash

	for i := range data {
		serializedData, err := data[i].Serialize()
		if err != nil {
			return block.Block{}, errors.New("Failed to serialize data")
		}

		serializedDatum = append(serializedDatum, serializedData)
	}

	// create the block
	block := block.Block{
		Version:        version,
		Height:         height,
		Timestamp:      time.Now().UnixNano(),
		PreviousHash:   previousHash,
		MerkleRootHash: block.GetMerkleRootHash(serializedDatum),
		DataLen:        uint16(len(data)),
		Data:           serializedDatum,
	}

	return block, nil
}

func BringOnTheGenesis(genesisPublicKeyHashes [][]byte, initialAurumSupply uint64) (block.Block, error) {
	version := uint16(1)
	mintAmt := initialAurumSupply / uint64(len(genesisPublicKeyHashes)) // (initialAurumSupply / n supplied key hashes)
	var datum []Data

	for _, pubKeyHash := range genesisPublicKeyHashes {
		// for every public key hashes, make a nil-sender contract with value indicated by mintAmt
		contract, err := accounts.MakeContract(version, nil, pubKeyHash, mintAmt, 0)
		if err != nil {
			return block.Block{}, errors.New("Failed to make contracts")
		}

		// data that contains data version and type, and the contract
		data := Data{
			Hdr: DataHeader{
				Version: version,
				Type:    0,
			},
			Bdy: contract,
		}
		datum = append(datum, data)
	}

	// create genesis block with null previous hash
	genesisBlock, err := CreateBlock(version, 0, make([]byte, 32), datum)
	if err != nil {
		return block.Block{}, errors.New("Failed to create genesis block")
	}

	return genesisBlock, nil
}

func Airdrop(blockchainz string, metadata string, genesisBlock block.Block) error {
	// create blockchain file
	file, err := os.Create(blockchainz)
	if err != nil {
		return errors.New("Failed to create blockchain file")
	}
	file.Close()

	// create metadata file
	file, err = os.Create(metadata)
	if err != nil {
		return errors.New("Failed to create metadata table")
	}
	file.Close()

	// open metadata file and create the table
	db, err := sql.Open("sqlite3", metadata)
	if err != nil {
		return errors.New("Failed to open table")
	}

	_, err = db.Exec("CREATE table METADATA (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	if err != nil {
		return errors.New("Failed to create table")
	}
	db.Close()

	// add genesis block into blockchain
	err = blockchain.AddBlock(genesisBlock, blockchainz, metadata)
	if err != nil {
		return errors.New("Failed to add genesis block into blockchain")
	}
	return nil
}

// This is a security feature for the ledger. If the metadata table gets lost somehow, this function will restore it completely.
//
// Another situation is when a producer in a decentralized system joins the network and wants the full ledger.
func RecoverBlockchainMetadata(ledgerFilename string, metadataFilename string, accountBalanceTable string) error {
	//check if metadata file exists
	err := emptyFile(metadataFilename)
	if err != nil {
		log.Printf("Failed to create an empty metadata file: %s", err.Error())
		return errors.New("Failed to create an empty metadata file")
	}
	//check if account balance file exists
	err = emptyFile(accountBalanceTable)
	if err != nil {
		log.Printf("Failed to create an empty account file: %s", err.Error())
		return errors.New("Failed to create an empty account file")
	}

	//set up the two database tables
	metaDb, err := sql.Open("sqlite3", metadataFilename)
	if err != nil {
		return errors.New("Failed to open newly created metadata db")
	}
	defer metaDb.Close()
	_, err = metaDb.Exec("CREATE TABLE metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
	if err != nil {
		return errors.New("Failed to create metadata table")
	}

	accDb, err := sql.Open("sqlite3", accountBalanceTable)
	if err != nil {
		return errors.New("Failed to open newly created accounts db")
	}
	defer accDb.Close()
	_, err = accDb.Exec("CREATE TABLE account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	if err != nil {
		return errors.New("Failed to create acount_balances table")
	}

	// open ledger file
	ledgerFile, err := os.OpenFile(ledgerFilename, os.O_RDONLY, 0644)
	if err != nil {
		return errors.New("Failed to open ledger file")
	}
	defer ledgerFile.Close()

	// loop that adds blocks' metadata into database
	bPosition := int64(0)
	for {
		bOldPos := bPosition
		deserializedBlock, bLen, err := extractBlock(ledgerFile, &bPosition)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Failed to extract block from ledger: %s", err.Error())
			return errors.New("Failed to extract block from ledger")
		}

		//update the metadata table
		err = insertMetadata(metaDb, deserializedBlock, bLen, bOldPos)
		if err != nil {
			return err
		}

		//update the account table
		err = updateAccountTable(accDb, deserializedBlock)
		if err != nil {
			return err
		}

	}

	return err
}

//creates an empty file if the file doesn't exist, or clears if the contents of the file if it exists
func emptyFile(fileName string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		f, err := os.Create(fileName)
		if err != nil {
			return errors.New("Failed to create " + fileName)
		}
		f.Close()
	} else { //file exits, so clear the file
		err = os.Truncate(fileName, 0)
		if err != nil {
			return errors.New("Failed to truncate " + fileName)
		}
	}
	return nil
}

//extract a block from the file, also update file position
func extractBlock(ledgerFile *os.File, pos *int64) (*block.Block, uint32, error) {
	length := make([]byte, 4)

	// read 4 bytes for blocks' length
	_, err := ledgerFile.Read(length)
	if err == io.EOF {
		return nil, 0, err
	} else if err != nil {
		return nil, 0, errors.New("Failed to read ledger file")
	}

	bLen := binary.LittleEndian.Uint32(length)

	// set offset for next read to get to the position of the block
	ledgerFile.Seek(*pos+int64(len(length)), 0)
	serialized := make([]byte, bLen)

	//update file position
	*pos += int64(len(length) + len(serialized))

	//extract block
	_, err = io.ReadAtLeast(ledgerFile, serialized, int(bLen))
	if err != nil {
		return nil, 0, errors.New("Failed to retrieve serialized block")
	}

	deserializedBlock := block.Deserialize(serialized)
	return &deserializedBlock, bLen, nil
}

/*inserts the block metadata into the metadata table
  NOTE: the db connection passed in should be open
*/
func insertMetadata(db *sql.DB, b *block.Block, bLen uint32, pos int64) error {
	bHeight := b.Height
	bHash := block.HashBlock(*b)

	sqlQuery := "INSERT INTO metadata (height, position, size, hash) VALUES ($1, $2, $3, $4)"
	_, err := db.Exec(sqlQuery, bHeight, pos, bLen, bHash)
	if err != nil {
		log.Printf("Failed to execute statement: %s", err.Error())
		return errors.New("Failed to execute statement")
	}

	return nil
}

/*calculates and inserts accounts' balance and nonce into the account balance table
  NOTE: the db connection passed in should be open
*/
func updateAccountTable(db *sql.DB, b *block.Block) error {
	//first deserialize the data
	deserializedDatum := make([]Data, len(b.Data))
	for i, serializedData := range b.Data {
		dataStruct := &Data{}
		dataStruct.Deserialize(serializedData)
		deserializedDatum[i] = *dataStruct
	}

	//retrieve contracts
	contracts := make([]*accounts.Contract, len(deserializedDatum))
	for i, data := range deserializedDatum {
		contracts[i] = data.Bdy.(*accounts.Contract)
	}

	//struct to keep track of everyone's account info
	type accountInfo struct {
		accountPKH []byte
		balance    int64
		nonce      uint64
	}

	totalBalances := make([]accountInfo, 0)
	minting := false
	for _, contract := range contracts {
		addRecip := true
		addSender := true

		if contract.SenderPubKey == nil { // minting contracts
			minting = true
			err := accounts.InsertAccountIntoAccountBalanceTable(db, contract.RecipPubKeyHash, contract.Value)
			if err != nil {
				return err
			}
			continue
		}

		for i := 0; i < len(totalBalances); i++ {
			if bytes.Compare(totalBalances[i].accountPKH, block.HashSHA256(keys.EncodePublicKey(contract.SenderPubKey))) == 0 {
				//subtract the value of the contract from the sender's account
				addSender = false
				totalBalances[i].balance -= int64(contract.Value)
				totalBalances[i].nonce++
			} else if bytes.Compare(totalBalances[i].accountPKH, contract.RecipPubKeyHash) == 0 {
				//add the value of the contract to the recipient's account
				addRecip = false
				totalBalances[i].balance += int64(contract.Value)
				totalBalances[i].nonce++
			}
		}

		//add the sender's account info into totalBalances
		if addSender {
			totalBalances = append(totalBalances,
				accountInfo{accountPKH: block.HashSHA256(keys.EncodePublicKey(contract.SenderPubKey)), balance: -1 * int64(contract.Value), nonce: 1})
		}

		//add the recipient's account info into totalBalances
		if addRecip {
			totalBalances = append(totalBalances,
				accountInfo{accountPKH: contract.RecipPubKeyHash, balance: int64(contract.Value), nonce: 1})
		}
	}

	//insert the accounts in totalBalances into account balance table
	if !minting {
		for _, acc := range totalBalances {
			var balance int
			var nonce int

			sqlQuery := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash= \"%s\"", hex.EncodeToString(acc.accountPKH))
			row, _ := db.Query(sqlQuery)
			if row.Next() {
				row.Scan(&balance, &nonce) // retrieve balance and nonce from account_balances
				row.Close()

				// update balance and nonce
				sqlUpdate := fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"",
					acc.balance+int64(balance), acc.nonce+uint64(nonce), hex.EncodeToString(acc.accountPKH))
				_, err := db.Exec(sqlUpdate)
				if err != nil {
					return errors.New("Failed to execute query to update balance and nonce: " + err.Error())
				}
			} else {
				row.Close()
				return errors.New("Failed to find row to update balance and nonce")
			}

		}
	}
	return nil
}
