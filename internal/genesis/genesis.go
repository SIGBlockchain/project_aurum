package genesis

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"

	"os"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
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

func BringOnTheGenesis(genesisPublicKeyHashes [][]byte, initialAurumSupply uint64) (block.Block, error) {
	version := uint16(1)
	mintAmt := initialAurumSupply / uint64(len(genesisPublicKeyHashes)) // (initialAurumSupply / n supplied key hashes)
	var datum []contracts.Contract

	for _, pubKeyHash := range genesisPublicKeyHashes {
		// for every public key hashes, make a nil-sender contract with value indicated by mintAmt
		contract, err := contracts.New(version, nil, pubKeyHash, mintAmt, 0)
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

// Open the genesisHashFile
// Read line by line
// use bufio.ReadLine()
func ReadGenesisHashes() ([][]byte, error) {
	//open genesisHashFile
	file, err := os.Open(constants.GenesisAddresses)
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
	genHashfile, _ := os.Create(constants.GenesisAddresses)

	defer genHashfile.Close()

	// create the hashes numHashes times
	for i := 0; i < int(numHashes); i++ {
		// generate private key
		privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		// get public kek and hash it
		hashedPubKey := hashing.New(publickey.Encode(&privateKey.PublicKey))

		// get pub key hash as string to store in txt file
		hashPubKeyStr := hex.EncodeToString(hashedPubKey)

		// write pub key hash into genesisHashFile
		genHashfile.WriteString(hashPubKeyStr + "\n")
	}
}
