package validation

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"encoding/asn1"
	"errors"
	"math/big"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

func ValidateContract(dbConnection *sql.DB, c *contracts.Contract) error {
	// check for zero value transaction
	if c.Value == 0 {
		return errors.New("Invalid contract: zero value transaction")
	}

	// check for nil sender public key and recip == sha-256 hash of senderPK
	encodedCSenderPublicKey, err := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return err
	}
	if c.SenderPubKey == nil || bytes.Equal(c.RecipPubKeyHash, hashing.New(encodedCSenderPublicKey)) {
		return errors.New("Invalid contract: sender cannot be nil nor same as recipient")
	}

	// verify the signature in the contract
	// Serialize the Contract
	copyOfSigLen := c.SigLen
	c.SigLen = 0
	serializedContract, err := c.Serialize()
	if err != nil {
		return errors.New(err.Error())
	}
	hashedContract := hashing.New(serializedContract)

	// stores r and s values needed for ecdsa.Verify
	var esig struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(c.Signature, &esig); err != nil {
		return errors.New("Failed to unmarshal signature")
	}

	// if ecdsa.Verify returns false, the signature is invalid
	if !ecdsa.Verify(c.SenderPubKey, hashedContract, esig.R, esig.S) {
		return errors.New("Invalid contract: signature is invalid")
	}

	// retrieve sender's balance from account balance table
	senderPubKeyHash := hashing.New(encodedCSenderPublicKey)
	senderAccountInfo, errAccount := accountstable.GetAccountInfo(dbConnection, senderPubKeyHash)

	if errAccount == nil {
		// check insufficient funds
		if senderAccountInfo.Balance < c.Value {
			// invalid contract because the sender's balance is less than the contract amount
			return errors.New("Invalid contract: sender's balance is less than the contract amount")
		}

		if senderAccountInfo.StateNonce+1 != c.StateNonce {
			// invalid contract because contract state nonce is not the expected number
			return errors.New("Invalid contract: contract state nonce is not the expected number")
		}

		/* valid contract */
		c.SigLen = copyOfSigLen
		return nil
	}

	return errors.New("Failed to validate contract")
}

// ValidatePending validates a contract with the given pending balance and pending state nonce
func ValidatePending(c *contracts.Contract, pBalance *uint64, pNonce *uint64) error {
	// check for zero value transaction
	if c.Value == 0 {
		return errors.New("Invalid contract: zero value transaction")
	}

	// check for nil sender public key and recip == sha-256 hash of senderPK
	recipPKhash := hashing.SHA256Hash{SecureHash: c.RecipPubKeyHash}

	encodedCSenderPublicKey, err := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return err
	}
	if c.SenderPubKey == nil || recipPKhash.Equals(encodedCSenderPublicKey) {
		return errors.New("Invalid contract: sender cannot be nil nor same as recipient")
	}

	// verify the signature in the contract
	// Serialize the Contract
	copyOfSigLen := c.SigLen
	c.SigLen = 0
	serializedContract, err := c.Serialize()
	if err != nil {
		return errors.New(err.Error())
	}
	hashedContract := hashing.New(serializedContract)

	// stores r and s values needed for ecdsa.Verify
	var esig struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(c.Signature, &esig); err != nil {
		return errors.New("Failed to unmarshal signature")
	}

	// if ecdsa.Verify returns false, the signature is invalid
	if !ecdsa.Verify(c.SenderPubKey, hashedContract, esig.R, esig.S) {
		return errors.New("Invalid contract: signature is invalid")
	}

	// if sender's pending balance is less than the contract amount, invalid contract
	if *pBalance < c.Value {
		return errors.New("Invalid contract: sender's pending balance is less than the contract amount")
	}

	// if contract state nonce is not one greater than the sender's pending state nonce, invalid contract
	if (*pNonce)+1 != c.StateNonce {
		return errors.New("Invalid contract: contract state nonce is not the expected number")
	}

	/* valid contract, return updated pending balance and state nonce */
	c.SigLen = copyOfSigLen
	*pBalance -= c.Value
	(*pNonce)++
	return nil
}

// ValidateBlock takes in expected version, height, previousHash, and timeStamp
// and compares them with the block's
func ValidateBlock(b block.Block, version uint16, prevHeight uint64, previousHash []byte, prevTimeStamp int64) bool {
	// Check Version
	if b.Version != version {
		return false
	}
	// Check Height
	if b.Height != prevHeight+1 {
		return false
	}
	// Check Previous Hash
	if !(bytes.Equal(b.PreviousHash, previousHash)) {
		return false
	}
	// Check timestamp
	if b.Timestamp <= prevTimeStamp || b.Timestamp > time.Now().UnixNano() {
		return false
	}
	// Check MerkleRoot
	if !hashing.MerkleRootHashOf(b.MerkleRootHash, b.Data) {
		return false
	}

	return true
}

// ValidateProducerTimestamp checks the parameter timestamp p to see if
// it is greater than the sum of the interval itv and
// the table timestamp t (corresponding to the walletAddr).
// False if p < t + itv
func ValidateProducerTimestamp(timestamp int64, walletAddr []byte, interval time.Duration) (bool, error) {
	// Open Producer table
	db, _ := sql.Open("sqlite3", constants.ProducerTable)

	// search for wallet address in table and return timestamp
	row, err := db.Query("SELECT timestamp FROM producer WHERE public_key_hash = ?", walletAddr)
	if err != nil {
		return false, err
	}
	defer row.Close()

	// verify row was found
	if !row.Next() {
		return false, nil
	}

	// Scan for timestamp value in database row
	var wT time.Duration
	row.Scan(&wT)

	// check if p < t + itv
	if time.Duration(timestamp) < (wT + interval) {
		return false, nil
	}

	return true, nil
}
