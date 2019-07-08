package pendingpool

import (
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/validation"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

//PendingData contains pending balance and pending nonce
type PendingData struct {
	PendingBal   uint64
	PendingNonce uint64
}

//PendingMap contains a map that maps a hex encoded string of a wallet address to a pointer of PendingData
type PendingMap struct {
	Sender map[string]*PendingData
}

//NewPendingData returns an instance of pendingData given pending balance and pending nonce
func NewPendingData(pendingBal uint64, pendingNonce uint64) PendingData {
	return PendingData{pendingBal, pendingNonce}
}

//NewPendingMap returns an instance of pendingMap given a wallet address and an instance of pendingData
func NewPendingMap() PendingMap {
	m := make(map[string]*PendingData)
	return PendingMap{m}
}

//Add returns an error if the process of validating the given contract has failed.
//Otherwise, Add either inserts the sender's PKhash and the PendingData struct into the map,
//or updates the pending balance and pending nonce for that sender's PKhash in the map
func (m *PendingMap) Add(c *contracts.Contract, accDB *sql.DB) error {
	senderPKHash := hashing.New(publickey.Encode(c.SenderPubKey))
	senderPKStr := hex.EncodeToString(senderPKHash) // hex encoded sender PKhash string for the key

	senderPD, inMap := m.Sender[senderPKStr]
	if !inMap { // if the key is not in the map
		err := validation.ValidateContract(c)
		if err != nil {
			return errors.New("Failed to validate contract: " + err.Error())
		}

		balance, err := accountstable.GetBalance(senderPKHash)
		if err != nil {
			return errors.New("Failed to find sender public key hash in accounts_balance")
		}

		pendingD := NewPendingData(uint64(balance)-c.Value, c.StateNonce) // create new pendingData struct for this sender
		m.Sender[senderPKStr] = &pendingD                                 // insert key and pendingData struct into the map
	} else if inMap { // if key is in the map
		err := validation.ValidatePending(c, &(senderPD.PendingBal), &(senderPD.PendingNonce))
		if err != nil {
			return errors.New("Failed to validate contract with pending balance and pending state nonce: " + err.Error())
		}
	}

	return nil
}
