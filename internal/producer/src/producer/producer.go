// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SIGBlockchain/project_aurum/pkg/keys"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/blockchain"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
)

type Flags struct {
	Help        *bool
	Debug       *bool
	Version     *bool
	Height      *bool
	Genesis     *bool
	Test        *bool
	Globalhost  *bool
	MemoryStats *bool
	Logs        *string
	Port        *string
	Interval    *string
	InitSupply  *uint64
	NumBlocks   *uint64
}

var version = uint16(1)
var ledger = "blockchain.dat"
var metadataTable = constants.MetadataTable

// TODO: Will need to change the name to support get functions
var accountsTable = constants.AccountsTable

var SecretBytes = hashing.New([]byte("aurum"))[8:16]

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

func RunServer(ln net.Listener, bChan chan []byte, debug bool) {
	// Set logger
	var lgr = log.New(ioutil.Discard, "SRVR_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if debug {
		lgr.SetOutput(os.Stdout)
	}
	for {
		lgr.Println("Waiting for connection...")

		// Block for connection
		conn, err := ln.Accept()
		var nRcvd int
		buf := make([]byte, 1024)
		if err != nil {
			lgr.Println("connection failed")
			goto End
		}
		lgr.Printf("%s connected\n", conn.RemoteAddr())
		defer conn.Close()

		// Block for receiving message
		nRcvd, err = conn.Read(buf)
		if err != nil {
			goto End
		}

		lgr.Println("Received message from", conn.RemoteAddr())

		// Determine the type of message
		if nRcvd < 8 || (!bytes.Equal(buf[:8], SecretBytes)) {
			conn.Write([]byte("No thanks.\n"))
			goto End
		} else {
			conn.Write([]byte("Thank you.\n"))
			if buf[8] == 2 {
				lgr.Println("Received account info request")
				// TODO: Will require a sync.Mutex lock here eventually
				accInfo, err := accounts.GetAccountInfo(buf[9:nRcvd])
				var responseMessage []byte
				responseMessage = append(responseMessage, SecretBytes...)
				if err != nil {
					lgr.Printf("Failed to get account info for %s: %s", hex.EncodeToString(buf[9:nRcvd]), err.Error())
					responseMessage = append(responseMessage, 1)
				} else {
					responseMessage = append(responseMessage, 0)
					if serializedAccInfo, err := accInfo.Serialize(); err == nil {
						responseMessage = append(responseMessage, serializedAccInfo...)
					}
				}

				time.Sleep(3 * time.Second)
				conn.Write(responseMessage)
				goto End
			}
		}

		lgr.Println("Sending to channel")
		// Send to channel if aurum-related message
		bChan <- buf[:nRcvd]

		lgr.Println("Message successfully sent to main")
		goto End
	End:
		lgr.Println("Closing connection.")
		conn.Close()
	}
}

func ProduceBlocks(byteChan chan []byte, fl Flags, limit bool) {
	// Set logger
	var lgr = log.New(ioutil.Discard, "PROD_LOG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	if *fl.Debug {
		lgr.SetOutput(os.Stdout)
	}

	// Retrieve youngest block header
	youngestBlockHeader, err := blockchain.GetYoungestBlockHeader(ledger, metadataTable)
	if err != nil {
		lgr.Fatalf("failed to retrieve youngest block: %s\n", err)
	}

	// Set up SIGINT channel
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Initialize other variables
	var numBlocksGenerated uint64
	var dataPool []accounts.Contract
	var ms runtime.MemStats

	// Determine production interval and start trigger goroutine
	productionInterval, err := time.ParseDuration(*fl.Interval)
	if err != nil {
		lgr.Fatalln("failed to parse block production interval")
	} else {
		lgr.Println("block production interval: " + *fl.Interval)
	}
	var intervalChannel = make(chan bool)
	go triggerInterval(intervalChannel, productionInterval)

	// Main loop
	for {
		var chainHeight = youngestBlockHeader.Height
		select {
		case message := <-byteChan:

			// If it's a contract, add it to the contract pool
			switch message[8] {
			case 1:
				lgr.Println("Received contract")
				var newContract accounts.Contract
				if err := newContract.Deserialize(message[9:]); err == nil {
					// TODO: Validate the contract prior to adding
					if err := accounts.ValidateContract(&newContract); err != nil {
						lgr.Println("Invalid contract because: " + err.Error())
					} else {
						dataPool = append(dataPool, newContract)
						lgr.Println("Valid contract")
					}
				}
				break
			}
		case <-intervalChannel:
			// Triggered if it's time to produce a block
			lgr.Printf("block ready for production: #%d\n", chainHeight+1)
			// lgr.Printf("Production block dataPool: %v", dataPool)
			if newBlock, err := block.New(version, chainHeight+1, block.HashBlockHeader(youngestBlockHeader), dataPool); err != nil {
				lgr.Fatalf("failed to add block %s", err.Error())
				os.Exit(1)
			} else {

				// Add the block
				if err := blockchain.AddBlock(newBlock, ledger, metadataTable); err != nil {
					lgr.Fatalf("failed to add block: %s", err.Error())
					os.Exit(1)
				} else {
					lgr.Printf("block produced: #%d\n", chainHeight+1)
					numBlocksGenerated++
					youngestBlockHeader = newBlock.GetHeader()
					go triggerInterval(intervalChannel, productionInterval)

					// TODO: for each contract in the dataPool, update the accounts table
					// TODO: will require a sync.Mutex for the accounts table
					dbConn, err := sql.Open("sqlite3", constants.AccountsTable)
					if err != nil {
						lgr.Fatalf("Failed to connect to accounts database: %v", err)
					}
					for _, contract := range dataPool {
						senderPKH := hashing.New(keys.EncodePublicKey(contract.SenderPubKey))
						err := accounts.ExchangeBetweenAccountsUpdateAccountBalanceTable(dbConn, senderPKH, contract.RecipPubKeyHash, contract.Value)
						if err != nil {
							lgr.Printf("Failed to add contract to accounts database: %v", err)
						}
					}
					dbConn.Close()
					dataPool = nil

					// Memory stats
					if *fl.MemoryStats {
						runtime.ReadMemStats(&ms)
						printMemstats(ms)
					}
					// If in test mode, break the loop
					if *fl.Test {
						lgr.Printf("Test mode: breaking loop")
						return
					}

					// If reached limit of blocks desired to be generated, break the loop
					if limit && (numBlocksGenerated >= *fl.NumBlocks) {
						lgr.Printf("Limit reached: # blocks generated: %d, blocks desired: %d\n", numBlocksGenerated, *fl.NumBlocks)
						return
					}
				}

			}

		case <-signalCh:

			// If you receive a SIGINT, exit the loop
			fmt.Print("\r")
			lgr.Println("Interrupt signal encountered, program terminating.")
			return
		}

	}
}

func printMemstats(ms runtime.MemStats) {
	// useful commands: go run -gcflags='-m -m' main.go <main flags>
	fmt.Printf("Bytes of allocated heap objects: %d", ms.Alloc)
	fmt.Printf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
	fmt.Printf("Count of heap objects allocated: %d", ms.Mallocs)
	fmt.Printf("Count of heap objects freed: %d", ms.Frees)
}

func triggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	// Triggers block production case
	time.Sleep(productionInterval)
	intervalChannel <- true
}

func calculateInterval(youngestBlockHeader block.BlockHeader, productionInterval time.Duration, intervalChannel chan bool) {
	var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
	timeSince := time.Since(lastTimestamp)
	if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
		go triggerInterval(intervalChannel, time.Duration(0))
	} else {
		diff := productionInterval.Nanoseconds() - timeSince.Nanoseconds()
		go triggerInterval(intervalChannel, time.Duration(diff))
	}
}

func BringOnTheGenesis(genesisPublicKeyHashes [][]byte, initialAurumSupply uint64) (block.Block, error) {
	version := uint16(1)
	mintAmt := initialAurumSupply / uint64(len(genesisPublicKeyHashes)) // (initialAurumSupply / n supplied key hashes)
	var datum []accounts.Contract

	for _, pubKeyHash := range genesisPublicKeyHashes {
		// for every public key hashes, make a nil-sender contract with value indicated by mintAmt
		contract, err := accounts.MakeContract(version, nil, pubKeyHash, mintAmt, 0)
		if err != nil {
			return block.Block{}, errors.New("Failed to make contracts")
		}
		datum = append(datum, *contract) // switched second parameter from data to contract
	}

	// create genesis block with null previous hash
	genesisBlock, err := block.New(version, 0, make([]byte, 32), datum)
	if err != nil {
		return block.Block{}, errors.New("Failed to create genesis block")
	}

	return genesisBlock, nil
}

func Airdrop(blockchainz string, metadata string, accountBalanceTable string, genesisBlock block.Block) error {
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

	// create accounts file
	file, err = os.Create(accountBalanceTable)
	if err != nil {
		return errors.New("Failed to create accounts table")
	}
	file.Close()

	accDb, err := sql.Open("sqlite3", accountBalanceTable)
	if err != nil {
		return errors.New("Failed to open newly created accounts db")
	}
	defer accDb.Close()

	_, err = accDb.Exec("CREATE TABLE account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	if err != nil {
		return errors.New("Failed to create acount_balances table")
	}

	stmt, err := accDb.Prepare("INSERT INTO account_balances VALUES (?, ?, ?)")
	if err != nil {
		return errors.New("Failed to create statement for inserting into account table")
	}

	for _, contracts := range genesisBlock.Data {
		var contract accounts.Contract
		contract.Deserialize(contracts)
		_, err := stmt.Exec(hex.EncodeToString(contract.RecipPubKeyHash), contract.Value, 0)
		if err != nil {
			return errors.New("Failed to execute statement for inserting into account table")
		}
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

	//retrieve contracts
	contracts := make([]*accounts.Contract, len(b.Data))
	for i, data := range b.Data {
		contracts[i] = &accounts.Contract{}
		err := contracts[i].Deserialize(data)
		if err != nil {
			return errors.New("Failed to deserialize contracts: " + err.Error())
		}
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
			if bytes.Compare(totalBalances[i].accountPKH, hashing.New(keys.EncodePublicKey(contract.SenderPubKey))) == 0 {
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
				accountInfo{accountPKH: hashing.New(keys.EncodePublicKey(contract.SenderPubKey)), balance: -1 * int64(contract.Value), nonce: 1})
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

var genesisHashFile = "genesis_hashes.txt"

// Open the genesisHashFile
// Read line by line
// use bufio.ReadLine()
func ReadGenesisHashes() ([][]byte, error) {
	//open genesisHashFile
	file, err := os.Open(genesisHashFile)
	if err != nil {
		return [][]byte{}, errors.New("Unable to open genesis_hashs.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	var hashesInBytes [][]byte

	// while loop to loop till EOF
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		decodedHash, _ := hex.DecodeString(string(line))
		// append to the byte slice that is going to be returned
		hashesInBytes = append(hashesInBytes, decodedHash)
	}

	return hashesInBytes, err
}

// Create the genesisHashFile
// Generate numHashes number of public key hashes
// Store them AS STRINGS (not bytes) in the file, line by line
func GenerateGenesisHashFile(numHashes uint16) {

	// creating the new file
	genHashfile, _ := os.Create(genesisHashFile)

	defer genHashfile.Close()

	// create the hashes numHashes times
	for i := 0; i < int(numHashes); i++ {
		// generate private key
		privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		// get public kek and hash it
		hashedPubKey := hashing.New(keys.EncodePublicKey(&privateKey.PublicKey))

		// get pub key hash as string to store in txt file
		hashPubKeyStr := hex.EncodeToString(hashedPubKey)

		// write pub key hash into genesisHashFile
		genHashfile.WriteString(hashPubKeyStr + "\n")
	}
}

// Generates n random keys and writes them to the file at filename
func GenerateNRandomKeys(filename string, n uint32) error {
	// Opens the file, if it does not exist the O_CREATE flag tells it to create the file otherwise overwrite file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	if n < 1 {
		return errors.New("Must generate at least one private key")
	}
	// Checks if the opening was successful
	if err != nil {
		return err
	}
	// jsonStruct, will contain the information inside the json file
	type jsonStruct struct {
		Privates []ecdsa.PrivateKey
	}

	var keys []string // This will hold all pem encoded private key strings
	var i uint32 = 0  // Iterator, is uint32 to be able to compare with n

	// Create n private keys
	for ; i < n; i++ {
		p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}
		// Encodes the private key
		x509Encoded, _ := x509.MarshalECPrivateKey(p)
		pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
		// Converts the encoded byte string into a string so it can be used in the json struct
		encodedStr := hex.EncodeToString(pemEncoded)
		// Add the encoded byte string to the string slice
		keys = append(keys, encodedStr)
	}

	// Write strings to file
	jbyte, err := json.Marshal(keys)

	// Checks if marshalling was successful
	if err != nil {
		return err
	}
	// Write into the file that was given
	_, err = file.Write(jbyte)

	// Checks if the writing was successful
	if err != nil {
		return err
	}

	return nil
}
