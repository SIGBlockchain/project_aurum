package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"encoding/asn1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

var accountBalanceTable = "accounts.tab"

/*
Version
Sender Public Key
Signature Length
Signature
Recipient Public Key Hash
Value
*/
type Contract struct {
	Version         uint16
	SenderPubKey    *ecdsa.PublicKey
	SigLen          uint8  // len of the signature
	Signature       []byte // size varies
	RecipPubKeyHash []byte // 32 bytes
	Value           uint64
	StateNonce      uint64
}

/*
version field comes from version parameter
sender public key comes from sender private key
signature comes from calling sign contract
signature length comes from signature
recipient pk hash comes from sha-256 hash of rpk
value is value parameter
returns contract struct
*/
func MakeContract(version uint16, sender *ecdsa.PrivateKey, recipient []byte, value uint64, nextStateNonce uint64) (*Contract, error) {

	if version == 0 {
		return nil, errors.New("Invalid version; must be >= 1")
	}

	c := Contract{
		Version:         version,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: recipient,
		Value:           value,
		StateNonce:      nextStateNonce,
	}

	if sender == nil {
		c.SenderPubKey = nil
	} else {
		c.SenderPubKey = &(sender.PublicKey)
	}

	return &c, nil
}

// // Serialize all fields of the contract
func (c *Contract) Serialize() ([]byte, error) {
	/*
		0-2 version
		2-180 spubkey
		180-181 siglen
		181 - 181+c.siglen signature
		181+c.siglen - (181+c.siglen + 32) rpkh
		(181+c.siglen + 32) - (181+c.siglen + 32+ 8) value

	*/

	// if contract's sender pubkey is nil, make 178 zeros in its place instead
	var spubkey []byte
	if c.SenderPubKey == nil {
		spubkey = make([]byte, 178)
	} else {
		spubkey = keys.EncodePublicKey(c.SenderPubKey) //size 178
	}

	//unsigned contract
	if c.SigLen == 0 {
		totalSize := (2 + 178 + 1 + 32 + 8 + 8)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = 0
		copy(serializedContract[181:213], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[213:221], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[221:229], c.StateNonce)

		return serializedContract, nil
	} else { //signed contract
		totalSize := (2 + 178 + 1 + int(c.SigLen) + 32 + 8 + 8)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = c.SigLen
		copy(serializedContract[181:(181+int(c.SigLen))], c.Signature)
		copy(serializedContract[(181+int(c.SigLen)):(181+int(c.SigLen)+32)], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32):(181+int(c.SigLen)+32+8)], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32+8):(181+int(c.SigLen)+32+8+8)], c.StateNonce)

		return serializedContract, nil
	}
}

// Deserialize into a struct
func (c *Contract) Deserialize(b []byte) error {
	var spubkeydecoded *ecdsa.PublicKey

	// if serialized sender public key contains only zeros, sender public key is nil
	if bytes.Equal(b[2:180], make([]byte, 178)) {
		spubkeydecoded = nil
	} else {
		spubkeydecoded = keys.DecodePublicKey(b[2:180])
	}
	siglen := int(b[180])

	// unsigned contract
	if siglen == 0 {
		c.Version = binary.LittleEndian.Uint16(b[0:2])
		c.SenderPubKey = spubkeydecoded
		c.SigLen = b[180]
		c.RecipPubKeyHash = b[181:213]
		c.Value = binary.LittleEndian.Uint64(b[213:221])
		c.StateNonce = binary.LittleEndian.Uint64(b[221:229])
	} else {
		c.Version = binary.LittleEndian.Uint16(b[0:2])
		c.SenderPubKey = spubkeydecoded
		c.SigLen = b[180]
		c.Signature = b[181:(181 + siglen)]
		c.RecipPubKeyHash = b[(181 + siglen):(181 + siglen + 32)]
		c.Value = binary.LittleEndian.Uint64(b[(181 + siglen + 32):(181 + siglen + 32 + 8)])
		c.StateNonce = binary.LittleEndian.Uint64(b[(181 + siglen + 32 + 8):(181 + siglen + 32 + 8 + 8)])
	}
	return nil
}

/*
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value)
signature = Sign ( hashed contract, sender private key )
sig len = signature length
siglen and sig go into respective fields in contract
*/
func (c *Contract) SignContract(sender *ecdsa.PrivateKey) error {
	serializedTestContract, err := c.Serialize()
	if err != nil {
		return errors.New("Failed to serialize contract")
	}
	hashedContract := block.HashSHA256(serializedTestContract)
	c.Signature, _ = sender.Sign(rand.Reader, hashedContract, nil)
	c.SigLen = uint8(len(c.Signature))
	return nil
}

/*
Insert into account balance table
Value set to value paramter
Nonce set to zero
Public Key Hash insert into pkhash column

Return every error possible with an explicit message
*/
func InsertAccountIntoAccountBalanceTable(dbConnection *sql.DB, pkhash []byte, value uint64) error {
	// create a prepared statement to insert into account_balances
	statement, err := dbConnection.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES(?, ?, ?)")
	if err != nil {
		return errors.New("Failed to prepare statement to insert account into table")
	}
	defer statement.Close()

	// execute the prepared statement to insert into account_balances
	_, err = statement.Exec(hex.EncodeToString(pkhash), value, 0)
	if err != nil {
		return errors.New("Failed to execute statement to insert account into table")
	}

	return nil
}

/*
Deduct value from sender's balance
Add value to recipient's balance
Increment both nonces by 1
*/
func ExchangeBetweenAccountsUpdateAccountBalanceTable(dbConnection *sql.DB, senderPKH []byte, recipPKH []byte, value uint64) error {
	// retrieve both sender's and recipient's balance and nonce
	senderAccountInfo, errSenderAccount := GetAccountInfo(senderPKH)
	recipientAccountInfo, errRecipientAccount := GetAccountInfo(recipPKH)

	if errSenderAccount == nil {
		// update sender's balance by subtracting the amount indicated by value and adding one to nonce
		sqlUpdate := fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"",
			int(senderAccountInfo.Balance-value), int(senderAccountInfo.StateNonce+1), hex.EncodeToString(senderPKH))
		_, err := dbConnection.Exec(sqlUpdate)
		if err != nil {
			return errors.New("Failed to execute sqlUpdate for sender")
		}

	} else {
		return errors.New("Cannot find Sender's account")
	}

	var updatedNonce, updatedBal int
	if errRecipientAccount == nil {
		// if recipient's account is found
		updatedBal = int(recipientAccountInfo.Balance + value)
		updatedNonce = int(recipientAccountInfo.StateNonce + 1)
	} else {
		// if recipient's account is not found, insert recipient's account into table
		err := InsertAccountIntoAccountBalanceTable(dbConnection, recipPKH, 0)
		if err != nil {
			return errors.New("Failed to insert recipient's account into table: " + err.Error())
		}
		updatedBal = int(value)
		updatedNonce = 0
	}

	// update recipient's balance with updatedBal and nonce with updatedNonce
	sqlUpdate := fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", updatedBal, updatedNonce, hex.EncodeToString(recipPKH))
	_, err := dbConnection.Exec(sqlUpdate)
	if err != nil {
		return errors.New("Failed to execute sqlUpdate for recipient")
	}

	return nil
}

/*
Add value to pkhash's balanace
Increment nonce by 1
*/
func MintAurumUpdateAccountBalanceTable(dbConnection *sql.DB, pkhash []byte, value uint64) error {
	// retrieve pkhash's balance and nonce
	accountInfo, errAccount := GetAccountInfo(pkhash)

	if errAccount == nil {
		// update pkhash's balance by adding the amount indicated by value, and add one to nonce
		sqlUpdate := fmt.Sprintf("UPDATE account_balances SET balance= %d, nonce= %d WHERE public_key_hash= \"%s\"",
			int(accountInfo.Balance)+int(value), int(accountInfo.StateNonce)+1, hex.EncodeToString(pkhash))
		_, err := dbConnection.Exec(sqlUpdate)
		if err != nil {
			return errors.New("Failed to update phash's balance")
		}
		return nil
	}

	return errors.New("Failed to find row")
}

func ValidateContract(c *Contract) error {
	// check for zero value transaction
	if c.Value == 0 {
		return errors.New("Invalid contract: zero value transaction")
	}

	// check for nil sender public key and recip == sha-256 hash of senderPK
	if c.SenderPubKey == nil || bytes.Equal(c.RecipPubKeyHash, block.HashSHA256(keys.EncodePublicKey(c.SenderPubKey))) {
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
	hashedContract := block.HashSHA256(serializedContract)

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
	senderPubKeyHash := block.HashSHA256(keys.EncodePublicKey(c.SenderPubKey))
	senderAccountInfo, errAccount := GetAccountInfo(senderPubKeyHash)

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

type AccountInfo struct {
	Balance    uint64
	StateNonce uint64
}

func NewAccountInfo(balance uint64, stateNonce uint64) *AccountInfo {
	return &AccountInfo{Balance: balance, StateNonce: stateNonce}
}

func (accInfo *AccountInfo) Serialize() ([]byte, error) {
	serializedAccount := make([]byte, 16) // 8 + 8 bytes for balance and stateNonce
	binary.LittleEndian.PutUint64(serializedAccount[:8], accInfo.Balance)
	binary.LittleEndian.PutUint64(serializedAccount[8:], accInfo.StateNonce)
	return serializedAccount, nil
}

func (accInfo *AccountInfo) Deserialize(serializedAccountInfo []byte) error {
	accInfo.Balance = binary.LittleEndian.Uint64(serializedAccountInfo[:8])
	accInfo.StateNonce = binary.LittleEndian.Uint64(serializedAccountInfo[8:])
	return nil
}

func GetBalance(pkhash []byte) (uint64, error) {
	// open account balance table
	db, err := sql.Open("sqlite3", "accounts.tab")
	if err != nil {
		return 0, errors.New("Failed to open account balance table")
	}
	defer db.Close()

	// search for pkhash's balance
	row, err := db.Query("SELECT balance FROM account_balances WHERE public_key_hash = \"" + hex.EncodeToString(pkhash) + "\"")
	if err != nil {
		return 0, errors.New("Failed to create row for query")
	}
	defer row.Close()

	if !row.Next() {
		return 0, errors.New("Failed to find row given pkHash")
	}

	var balance uint64
	err = row.Scan(&balance)
	if err != nil {
		return 0, errors.New("Failed to scan row")
	}
	return balance, nil
}

func GetStateNonce(pkhash []byte) (uint64, error) {
	// open account balance table
	db, err := sql.Open("sqlite3", "accounts.tab")
	if err != nil {
		return 0, errors.New("Failed to open account balance table")
	}
	defer db.Close()

	// search for pkhash's stateNonce
	row, err := db.Query("SELECT nonce FROM account_balances WHERE public_key_hash= \"" + hex.EncodeToString(pkhash) + "\"")
	if err != nil {
		return 0, errors.New("Failed to create row for query")
	}
	defer row.Close()

	if !row.Next() {
		return 0, errors.New("Failed to find row given pkHash")
	}

	var stateNonce uint64
	err = row.Scan(&stateNonce)
	if err != nil {
		return 0, errors.New("Failed to scan row")
	}
	return stateNonce, nil
}

func GetAccountInfo(pkhash []byte) (*AccountInfo, error) {
	// retrieve pkhash's balance
	balance, err := GetBalance(pkhash)
	if err != nil {
		return nil, errors.New("Failed to retreive balance: " + err.Error())
	}

	// retrieve pkhash's stateNonce
	stateNonce, err := GetStateNonce(pkhash)
	if err != nil {
		return nil, errors.New("Failed to retreive stateNonce: " + err.Error())
	}

	return &AccountInfo{Balance: balance, StateNonce: stateNonce}, nil
}

// compare two contracts and return true only if all fields match
func Equals(contract1 Contract, contract2 Contract) bool {
	// copy both contracts
	c1val := reflect.ValueOf(contract1)
	c2val := reflect.ValueOf(contract2)

	// loops through fields
	for i := 0; i < c1val.NumField(); i++ {
		finterface1 := c1val.Field(i).Interface() // value assignment from c1 as interface
		finterface2 := c2val.Field(i).Interface() // value assignment from c2 as interface

		switch finterface1.(type) { // switch on type
		case uint8, uint16, uint64, int64:
			if finterface1 != finterface2 {
				return false
			}
		case []byte:
			if !bytes.Equal(finterface1.([]byte), finterface2.([]byte)) {
				return false
			}
		case [][]byte:
			for i := 0; i < len(finterface1.([][]byte)); i++ {
				if !bytes.Equal(finterface1.([][]byte)[i], finterface2.([][]byte)[i]) {
					return false
				}
			}
		case *ecdsa.PublicKey:
			if !reflect.DeepEqual(finterface1, finterface2) {
				return false
			}
		}
	}
	return true
}

// func ValidateContract(c *Contract, table string, authorizedMinters [][]byte) (bool, error) {
// 	db, err := sql.Open("sqlite3", table)
// 	if err != nil {
// 		return false, errors.New("Failed to open table")
// 	}
// 	defer db.Close()

// 	// check for zero value transaction
// 	if c.Value == 0 {
// 		return false, nil
// 	}

// 	// if contract is for minting
// 	if c.SenderPubKey == nil {
// 		// check for unauthorized minting contracts
// 		for _, mintersPubKHash := range authorizedMinters {
// 			if bytes.Equal(c.RecipPubKeyHash, mintersPubKHash) {
// 				// authorized minting
// 				err = MintAurumUpdateAccountBalanceTable(db, c.RecipPubKeyHash, c.Value)
// 				if err != nil {
// 					return false, errors.New("Failed to mint aurum with a valid minting contract: " + err.Error())
// 				}

// 				return true, nil
// 			}
// 		}
// 		// unauthorized minting
// 		return false, nil
// 	}

// 	// verify the signature in the contract
// 	// Serialize the Contract
// 	copyOfSigLen := c.SigLen
// 	c.SigLen = 0
// 	serializedContract, err := c.Serialize()
// 	if err != nil {
// 		return false, errors.New(err.Error())
// 	}
// 	hashedContract := block.HashSHA256(serializedContract)

// 	// stores r and s values needed for ecdsa.Verify
// 	var esig struct {
// 		R, S *big.Int
// 	}
// 	if _, err := asn1.Unmarshal(c.Signature, &esig); err != nil {
// 		return false, errors.New("Failed to unmarshal signature")
// 	}

// 	// if ecdsa.Verify returns false, the signature is invalid
// 	if !ecdsa.Verify(c.SenderPubKey, hashedContract, esig.R, esig.S) {
// 		return false, nil
// 	}

// 	// retrieve sender's balance from account balance table
// 	senderPubKeyHash := block.HashSHA256(keys.EncodePublicKey(c.SenderPubKey))
// 	senderAccountInfo, errAccount := GetAccountInfo(senderPubKeyHash)

// 	if errAccount == nil {
// 		// check insufficient funds
// 		if senderAccountInfo.balance < c.Value {
// 			// invalid contract because the sender's balance is less than the contract amount
// 			return false, nil
// 		}

// 		if senderAccountInfo.stateNonce+1 != c.StateNonce {
// 			// invalid contract because contract state nonce is not the expected number
// 			return false, nil
// 		}

// 		// V--- valid contract ---V
// 		// update both the sender's and recipient's accounts
// 		err = ExchangeBetweenAccountsUpdateAccountBalanceTable(db, senderPubKeyHash, c.RecipPubKeyHash, c.Value)
// 		if err != nil {
// 			return false, errors.New("Failed to exchange between acccounts: " + err.Error())
// 		}

// 		c.SigLen = copyOfSigLen
// 		return true, nil
// 	}

// 	return false, errors.New("Failed to validate contract")
// }
