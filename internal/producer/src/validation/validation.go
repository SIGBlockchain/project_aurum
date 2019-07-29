package validation

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"encoding/asn1"
	"errors"
	"math/big"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

func ValidateContract(dbConnection *sql.DB, c *contracts.Contract) error {
	// check for zero value transaction
	if c.Value == 0 {
		return errors.New("Invalid contract: zero value transaction")
	}

	// check for nil sender public key and recip == sha-256 hash of senderPK
	if c.SenderPubKey == nil || bytes.Equal(c.RecipPubKeyHash, hashing.New(publickey.Encode(c.SenderPubKey))) {
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
	senderPubKeyHash := hashing.New(publickey.Encode(c.SenderPubKey))
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
	if c.SenderPubKey == nil || recipPKhash.Equals(publickey.Encode(c.SenderPubKey)) {
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
