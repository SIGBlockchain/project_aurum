package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"

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
Nonce
*/
type Contract struct {
	Version         uint16
	SenderPubKey    *ecdsa.PublicKey
	SigLen          uint8  // len of the signature
	Signature       []byte // size varies
	RecipPubKeyHash []byte // 32 bytes
	Value           uint64
	Nonce           uint64
}

/*
version field comes from version parameter
sender public key comes from sender private key
signature comes from calling sign contract
signature length comes from signature
recipient pk hash comes from sha-256 hash of rpk
value is value parameter
nonce is nonce parameter
returns contract struct
*/
func MakeContract(version uint16, sender *ecdsa.PrivateKey, recipient []byte, value uint64, nonce uint64) (*Contract, error) {

	if version == 0 {
		return nil, errors.New("Invalid version; must be >= 1")
	}

	c := Contract{
		Version:         version,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: recipient,
		Value:           value,
		Nonce:           nonce,
	}

	if sender == nil {
		c.SenderPubKey = nil
	} else {
		c.SenderPubKey = &(sender.PublicKey)
	}

	return &c, nil
}

// // Serialize all fields of the contract
func (c *Contract) Serialize(withSignature bool) []byte {
	/*
		0-2 version
		2-180 spubkey
		180-181 siglen
		181 - 181+c.siglen signature
		181+c.siglen - (181+c.siglen + 32) rpkh
		(181+c.siglen + 32) - (181+c.siglen + 32+ 8) value
		(181+c.siglen + 32+ 8) - (181+c.siglen + 32 + 8 + 8) nonce

	*/

	// if contract's sender pubkey is nil, make 178 zeros in its place instead
	var spubkey []byte
	if c.SenderPubKey == nil {
		spubkey = make([]byte, 178)
	} else {
		spubkey = keys.EncodePublicKey(c.SenderPubKey) //size 178
	}

	//unsigned contract
	if withSignature == false {
		totalSize := (2 + 178 + 1 + 32 + 16)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = 0
		copy(serializedContract[181:213], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[213:221], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[221:229], c.Nonce)

		return serializedContract
	} else { //signed contract
		totalSize := (2 + 178 + 1 + int(c.SigLen) + 32 + 16)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = c.SigLen
		copy(serializedContract[181:(181+int(c.SigLen))], c.Signature)
		copy(serializedContract[(181+int(c.SigLen)):(181+int(c.SigLen)+32)], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32):(181+int(c.SigLen)+32+8)], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32+8):(181+int(c.SigLen)+32+8+8)], c.Nonce)

		return serializedContract
	}
}

// Deserialize into a struct
func (c *Contract) Deserialize(b []byte) {
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
		c.Nonce = binary.LittleEndian.Uint64(b[221:229])
	} else {
		c.Version = binary.LittleEndian.Uint16(b[0:2])
		c.SenderPubKey = spubkeydecoded
		c.SigLen = b[180]
		c.Signature = b[181:(181 + siglen)]
		c.RecipPubKeyHash = b[(181 + siglen):(181 + siglen + 32)]
		c.Value = binary.LittleEndian.Uint64(b[(181 + siglen + 32):(181 + siglen + 32 + 8)])
		c.Nonce = binary.LittleEndian.Uint64(b[(181 + siglen + 32 + 8):(181 + siglen + 32 + 8 + 8)])
	}
}

/*
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
signature = Sign ( hashed contract, sender private key )
sig len = signature length
siglen and sig go into respective fields in contract
*/
func (c *Contract) SignContract(sender *ecdsa.PrivateKey) {
	serializedTestContract := block.HashSHA256(c.Serialize(false))
	c.Signature, _ = sender.Sign(rand.Reader, serializedTestContract, nil)
	c.SigLen = uint8(len(c.Signature))
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
	rows, err := dbConnection.Query("SELECT public_key_hash , balance, nonce FROM account_balances")
	if err != nil {
		return errors.New("Failed to create rows to look for public key")
	}

	// search for the senders public key hash that belongs to the contract and update its fields
	var pkh string
	var tblBal int
	var tblNonce int
	var sqlQuery string
	for rows.Next() {
		err = rows.Scan(&pkh, &tblBal, &tblNonce)
		if err != nil {
			return errors.New("Failed to scan rows")
		}

		if reflect.DeepEqual(pkh, (hex.EncodeToString(senderPKH))) {
			rows.Close()
			compareVal := hex.EncodeToString(senderPKH)
			sqlQuery = fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", tblBal-int(value), tblNonce+1, compareVal)
		}
	}
	_, err = dbConnection.Exec(sqlQuery)
	if err != nil {
		return errors.New("Failed to execute sql query")
	}

	// new query to update the receiver
	rows, err = dbConnection.Query("SELECT public_key_hash , balance, nonce FROM account_balances")
	if err != nil {
		return errors.New("Failed the update the receiver query")
	}

	for rows.Next() {
		err = rows.Scan(&pkh, &tblBal, &tblNonce)
		if err != nil {
			return errors.New("Failed to scan rows")
		}

		if reflect.DeepEqual(pkh, (hex.EncodeToString(recipPKH))) {
			rows.Close()
			compareVal := hex.EncodeToString(recipPKH)
			sqlQuery = fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", tblBal+int(value), tblNonce+1, compareVal)
		}
	}
	_, err = dbConnection.Exec(sqlQuery)
	if err != nil {
		return errors.New("Failed to update recipient after searching in rows")
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

	var balance int
	var nonce int
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

	// check for unauthorized minting contracts
	if c.SenderPubKey == nil {
		validMinting := false
		for _, mintersPubKHash := range authorizedMinters {
			if bytes.Equal(c.RecipPubKeyHash, mintersPubKHash) {
				validMinting = true
				break
			}
		}
		if !validMinting {
			return false, nil
		}
	}
	return false, errors.New("Incomplete function")
}

// /*
// Check balance (ideal scenario):
// Open table
// Get hash of contract
// Verify signature with hash and public key
// Go to table and find sender
// Confirm balance is sufficient
// Update Account Balances (S & R)		// ony updated when true
// Increment Table Nonce

// 1. verify signature
// hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
// verify (hashed contract, spubkey, signature) (T)
// 2. validate amount
// check table to see if sender's balance >= contract amount (T)
// 3. validate nonce
// check to see that nonce is 1 + table nonce for that account (T)

// If all 3 are true, update table
// */
// func ValidateContract(c Contract, dbConnection *sql.DB) (bool, error) {
// 	// check for zero value transaction
// 	if c.Value == 0 {
// 		return false, errors.New("the transaction was zero value")
// 	}

// 	// Serialize the Contract
// 	serializedContract := block.HashSHA256(c.Serialize(false))

// 	// stores r and s values needed for ecdsa.Verify
// 	var esig struct {
// 		R, S *big.Int
// 	}
// 	if _, err := asn1.Unmarshal(c.Signature, &esig); err != nil {
// 		return false, errors.New("Failed to unmarshal signature")
// 	}

// 	// if the ecdsa.Verify is true then check the rest of the contract against whats in the databasei
// 	if !ecdsa.Verify(&c.SenderPubKey, serializedContract, esig.R, esig.S) {
// 		return false, errors.New("failed to verify signature")
// 	}

// 	rows, err := dbConnection.Query("SELECT public_key_hash , balance, nonce FROM account_balances")
// 	if err != nil {
// 		return false, err
// 	}
// 	defer rows.Close()

// 	//look for the public key that pertains to the contract and verify its balance and nonce
// 	var pkh string
// 	var tblBal int
// 	var tblNonce int
// 	for rows.Next() {
// 		err = rows.Scan(&pkh, &tblBal, &tblNonce)
// 		if err != nil {
// 			return false, errors.New("Failed ot scan rows")
// 		}
// 		if reflect.DeepEqual(pkh, (hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey))))) {
// 			if !(tblBal >= int(c.Value)) {
// 				return false, errors.New("tblBal is less than c.Value")
// 			}
// 			if !(tblNonce+1 == int(c.Nonce)) {
// 				return false, errors.New("couldn't validate")
// 			}
// 		}
// 	}

// 	err = c.UpdateAccountBalanceTable(dbConnection)
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }

// /*
// spkh = sha256 ( serialized sender pub key )
// find sender public key hash
// decrease value from sender public key hash account
// increment their nonce by one
// increase value of recipient public key hash account by contract value
// */
// func (c *Contract) UpdateAccountBalanceTable(dbConnection *sql.DB) error {
// 	rows, err := dbConnection.Query("SELECT public_key_hash , balance, nonce FROM account_balances")
// 	if err != nil {
// 		return errors.New("Failed to create rows to look for public key")
// 	}

// 	// search for the senders public key hash that belongs to the contract and update its fields
// 	var pkh string
// 	var tblBal int
// 	var tblNonce int
// 	var sqlQuery string
// 	for rows.Next() {
// 		rows.Scan(&pkh, &tblBal, &tblNonce)
// 		if reflect.DeepEqual(pkh, (hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey))))) {
// 			rows.Close()
// 			compareVal := hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey)))
// 			sqlQuery = fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", tblBal-int(c.Value), tblNonce+1, compareVal)
// 		}
// 	}

// 	_, err = dbConnection.Exec(sqlQuery)
// 	if err != nil {
// 		return errors.New("Failed to execute sql query")
// 	}

// 	// new query to update the receiver
// 	rows, err = dbConnection.Query("SELECT public_key_hash , balance, nonce FROM account_balances")
// 	if err != nil {
// 		return errors.New("Failed the update the receiver query")
// 	}

// 	for rows.Next() {
// 		err = rows.Scan(&pkh, &tblBal, &tblNonce)
// 		if err != nil {
// 			return errors.New("Failed to scan rows")
// 		}
// 		if reflect.DeepEqual(pkh, (hex.EncodeToString(c.RecipPubKeyHash))) {
// 			rows.Close()
// 			compareVal := hex.EncodeToString(c.RecipPubKeyHash)
// 			sqlQuery = fmt.Sprintf("UPDATE account_balances set balance=%d, nonce=%d WHERE public_key_hash= \"%s\"", tblBal+int(c.Value), tblNonce+1, compareVal)
// 		}
// 	}

// 	_, err = dbConnection.Exec(sqlQuery)
// 	if err != nil {
// 		return errors.New("Failed to update recipient after searching in rows")
// 	}

// 	return nil
// }
