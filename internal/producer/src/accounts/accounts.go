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

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
)

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
	senderQuery := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash= \"%s\"", hex.EncodeToString(senderPKH))
	recipientQuery := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash= \"%s\"", hex.EncodeToString(recipPKH))
	var tblBal, tblNonce int

	// search for the sender's account
	row, err := dbConnection.Query(senderQuery)
	if err != nil {
		return errors.New("Failed to create row to look for sender")
	}

	if row.Next() {
		// if sender's account is found, retrieve the balance and nonce and close the query
		err = row.Scan(&tblBal, &tblNonce)
		if err != nil {
			return errors.New("Failed to scan row")
		}
		row.Close()

		// update sender's balance by subtracting the amount indicated by value and adding one to nonce
		sqlUpdate := fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", tblBal-int(value), tblNonce+1, hex.EncodeToString(senderPKH))
		_, err = dbConnection.Exec(sqlUpdate)
		if err != nil {
			return errors.New("Failed to execute sqlUpdate for sender")
		}

	} else {
		row.Close()
		return errors.New("Cannot find Sender's account")
	}

	// search for the recipient's account
	row, err = dbConnection.Query(recipientQuery)
	if err != nil {
		return errors.New("Failed to create row to look for recipient")
	}

	var updatedNonce, updatedBal int
	if row.Next() {
		// if recipient's account is found, retrieve the balance and nonce and close the query
		err = row.Scan(&tblBal, &tblNonce)
		if err != nil {
			return errors.New("Failed to scan row")
		}
		row.Close()
		updatedBal = tblBal + int(value)
		updatedNonce = tblNonce + 1
	} else {
		// if recipient's account is not found, close the query and insert recipient's account into table
		row.Close()
		err = InsertAccountIntoAccountBalanceTable(dbConnection, recipPKH, 0)
		if err != nil {
			return errors.New("Failed to insert recipient's account into table: " + err.Error())
		}
		updatedBal = int(value)
		updatedNonce = 0
	}

	// update recipient's balance with updatedBal and nonce with updatedNonce
	sqlUpdate := fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", updatedBal, updatedNonce, hex.EncodeToString(recipPKH))
	_, err = dbConnection.Exec(sqlUpdate)
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
	// create a query for the row that contains the pkhash in the table
	sqlQuery := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash= \"%s\"", hex.EncodeToString(pkhash))
	row, err := dbConnection.Query(sqlQuery)
	if err != nil {
		return errors.New("Failed to create rows for query")
	}

	var balance, nonce int
	if row.Next() {
		// if row is found, retrieve balance and nonce and close the query
		err = row.Scan(&balance, &nonce)
		if err != nil {
			return errors.New("Failed to scan row")
		}
		row.Close()

		// update pkhash's balance by adding the amount indicated by value, and add one to nonce
		sqlUpdate := fmt.Sprintf("UPDATE account_balances SET balance= %d, nonce= %d WHERE public_key_hash= \"%s\"", balance+int(value), nonce+1, hex.EncodeToString(pkhash))
		_, err = dbConnection.Exec(sqlUpdate)
		if err != nil {
			return errors.New("Failed to update phash's balance")
		}
		return nil
	}

	return errors.New("Failed to find row")
}

func ValidateContract(c *Contract, table string, authorizedMinters [][]byte) (bool, error) {
	db, err := sql.Open("sqlite3", table)
	if err != nil {
		return false, errors.New("Failed to open table")
	}
	defer db.Close()

	// check for zero value transaction
	if c.Value == 0 {
		return false, nil
	}

	// if contract is for minting
	if c.SenderPubKey == nil {
		// check for unauthorized minting contracts
		for _, mintersPubKHash := range authorizedMinters {
			if bytes.Equal(c.RecipPubKeyHash, mintersPubKHash) {
				// authorized minting
				err = MintAurumUpdateAccountBalanceTable(db, c.RecipPubKeyHash, c.Value)
				if err != nil {
					return false, errors.New("Failed to mint aurum with a valid minting contract: " + err.Error())
				}

				return true, nil
			}
		}
		// unauthorized minting
		return false, nil
	}

	// verify the signature in the contract
	// Serialize the Contract
	copyOfSigLen := c.SigLen
	c.SigLen = 0
	serializedContract, err := c.Serialize()
	if err != nil {
		return false, errors.New(err.Error())
	}
	hashedContract := block.HashSHA256(serializedContract)

	// stores r and s values needed for ecdsa.Verify
	var esig struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(c.Signature, &esig); err != nil {
		return false, errors.New("Failed to unmarshal signature")
	}

	// if ecdsa.Verify returns false, the signature is invalid
	if !ecdsa.Verify(c.SenderPubKey, hashedContract, esig.R, esig.S) {
		return false, nil
	}

	// create a query for the row that contains the sender's pkhash in the table
	senderPubKeyStr := hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(c.SenderPubKey)))
	sqlQuery := fmt.Sprintf("SELECT balance, nonce FROM account_balances WHERE public_key_hash= \"%s\"", senderPubKeyStr)
	row, err := db.Query(sqlQuery)
	if err != nil {
		return false, errors.New("Failed to create row for query")
	}
	defer row.Close()

	if row.Next() {
		// if row is found, retrieve the sender's balance and close the query
		var tblBalance int
		var tblNonce int
		row.Scan(&tblBalance, &tblNonce)
		row.Close()

		// check insufficient funds
		if tblBalance < int(c.Value) {
			// invalid contract because the sender's balance is less than the contract amount
			return false, nil
		}

		if (tblNonce + 1) != int(c.StateNonce) {
			// invalid contract because contract state nonce is not the expected number
			return false, nil
		}

		// V--- valid contract ---V
		senderPubKeyHash, err := hex.DecodeString(senderPubKeyStr)
		if err != nil {
			return false, errors.New("Failed to decode senderPubKeyStr")
		}

		// update both the sender's and recipient's accounts
		err = ExchangeBetweenAccountsUpdateAccountBalanceTable(db, senderPubKeyHash, c.RecipPubKeyHash, c.Value)
		if err != nil {
			return false, errors.New("Failed to exchange between acccounts: " + err.Error())
		}

		c.SigLen = copyOfSigLen
		return true, nil
	}

	return false, errors.New("Failed to validate contract")
}
