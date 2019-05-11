// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// This is a security feature for the ledger. If the block table gets lost somehow, this function will restore it completely.
//
// Another situation is when a producer in a decentralized system joins the network and wants the full ledger.
func RecoverBlockchainMetadata(ledgerFilename string, metadataFilename string, accountBalanceTable string) error {
	_, err := os.Stat(metadataFilename)
	if err != nil {
		// create database
		f, err := os.Create(metadataFilename)
		f.Close()

		// open database
		db, err := sql.Open("sqlite3", metadataFilename)
		if err != nil {
			return errors.New("Failed to open newly created database")
		}
		defer db.Close()

		// create metadata table in database
		statement, _ := db.Prepare("CREATE TABLE metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)")
		statement.Exec()
		statement.Close()

		// create a prepared statement to insert into metadata
		statement, err = db.Prepare("INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)")
		if err != nil {
			return errors.New("Failed to prepare a statement for further executes")
		}
		defer statement.Close()

		// open ledger file
		file, err := os.OpenFile(ledgerFilename, os.O_RDONLY, 0644)
		if err != nil {
			return errors.New("Failed to open ledger file")
		}
		defer file.Close()

		// loop that adds blocks' metadata into database
		bPosition := int64(0)
		for {
			length := make([]byte, 4)

			// read 4 bytes for blocks' length
			_, err = file.Read(length)
			if err == io.EOF {
				// if reader reaches EOF, exit loop
				break
			} else if err != nil {
				return errors.New("Failed to read ledger file")
			}
			bLen := binary.LittleEndian.Uint32(length)

			// set offset for next read to get to the position of the block
			file.Seek(bPosition+int64(len(length)), 0)
			serialized := make([]byte, bLen)
			_, err = io.ReadAtLeast(file, serialized, int(bLen))
			if err != nil {
				return errors.New("Failed to retrieve serialized block")
			}

			// need to deserialize block for block's height and hash
			deserializedBlock := block.Deserialize(serialized)
			bHeight := deserializedBlock.Height
			bHash := block.HashBlock(deserializedBlock)

			// execute statement
			_, err = statement.Exec(bHeight, bPosition, bLen, bHash)
			if err != nil {
				return errors.New("Failed to execute statement")
			}

			// position of next block
			bPosition += int64(len(length) + len(serialized))
		}
	}
	return err
}
