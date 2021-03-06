package accountstable

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/SIGBlockchain/project_aurum/internal/accountinfo"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

// structs

// Connection struct used receiver call on a sql.DB with mutex support
type Connection struct {
	lock   sync.RWMutex
	dbconn *sql.DB
}

// Lock locks Connection for writing. If the lock is already locked for writing,
// Lock blocks until the lock is available.
func (c *Connection) Lock() {
	c.lock.Lock()
}

// Unlock unlocks Connection for writing. It is a run-time error if at is not locked for
// writing on entry to Unlock.
func (c *Connection) Unlock() {
	c.lock.Unlock()
}

// InsertAccountIntoAccountBalanceTable calls Connection.InsertAccountIntoAccountBalanaceTable using struct connection
func (c *Connection) InsertAccountIntoAccountBalanceTable(pkhash []byte, value uint64) error {
	return InsertAccountIntoAccountBalanceTable(c.dbconn, pkhash, value)
}

// ExchangeAndUpdateAccounts calls Connection.ExchangeAndUpdateAccounts using struct connection
func (c *Connection) ExchangeAndUpdateAccounts(contract *contracts.Contract) error {
	return ExchangeAndUpdateAccounts(c.dbconn, contract)
}

// MintAurumUpdateAccountBalanceTable calls Connection.MintAurumUpdateAccountBalanceTable using struct connection
func (c *Connection) MintAurumUpdateAccountBalanceTable(pkhash []byte, value uint64) error {
	return MintAurumUpdateAccountBalanceTable(c.dbconn, pkhash, value)
}

// GetBalance locks access and calls Connection.GetBalance using struct connection
func (c *Connection) GetBalance(pkhash []byte) (uint64, error) {
	c.Lock()
	bal, err := GetBalance(c.dbconn, pkhash)
	c.Unlock()
	if err != nil {
		return 0, errors.New("cannot access balance: " + err.Error())
	}
	return bal, nil
}

// GetStateNonce calls Connection.GetStateNonce using struct connection
func (c *Connection) GetStateNonce(pkhash []byte) (uint64, error) {
	c.Lock()
	bal, err := GetStateNonce(c.dbconn, pkhash)
	c.Unlock()
	if err != nil {
		return 0, errors.New("cannot access balance: " + err.Error())
	}
	return bal, nil
}

// GetAccountInfo calls Connection.GetAccountInfo using struct connection
func (c *Connection) GetAccountInfo(pkhash []byte) (*accountinfo.AccountInfo, error) {
	c.Lock()
	bal, err := GetAccountInfo(c.dbconn, pkhash)
	c.Unlock()
	if err != nil {
		return nil, errors.New("cannot access balance: " + err.Error())
	}
	return bal, nil
}

// UpdateAccountTable calls Connection.UpdateAccountTable using struct connection
func (c *Connection) UpdateAccountTable(b *block.Block) error {
	return UpdateAccountTable(c.dbconn, b)
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
	statement, err := dbConnection.Prepare(sqlstatements.INSERT_VALUES_INTO_ACCOUNT_BALANCES)
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
func ExchangeAndUpdateAccounts(dbConnection *sql.DB, c *contracts.Contract) error {
	encodedCContractSenderPublicKey, err := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return err
	}
	senderPKH := hashing.New(encodedCContractSenderPublicKey)
	recipPKH := c.RecipPubKeyHash
	value := c.Value

	// retrieve both sender's and recipient's balance and nonce
	senderAccountInfo, errSenderAccount := GetAccountInfo(dbConnection, senderPKH)
	recipientAccountInfo, errRecipientAccount := GetAccountInfo(dbConnection, recipPKH)

	if errSenderAccount == nil {
		// update sender's balance by subtracting the amount indicated by value and adding one to nonce
		_, err := dbConnection.Exec(sqlstatements.UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH,
			int(senderAccountInfo.Balance-value), int(senderAccountInfo.StateNonce+1), hex.EncodeToString(senderPKH))
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
	_, err = dbConnection.Exec(sqlstatements.UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, updatedBal, updatedNonce, hex.EncodeToString(recipPKH))
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
	accountInfo, errAccount := GetAccountInfo(dbConnection, pkhash)

	if errAccount == nil {
		// update pkhash's balance by adding the amount indicated by value, and add one to nonce
		_, err := dbConnection.Exec(sqlstatements.UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH,
			int(accountInfo.Balance)+int(value), int(accountInfo.StateNonce)+1, hex.EncodeToString(pkhash))
		if err != nil {
			return errors.New("Failed to update phash's balance")
		}
		return nil
	}

	return errors.New("Failed to find row")
}

func GetBalance(dbConnection *sql.DB, pkhash []byte) (uint64, error) {
	// search for pkhash's balance
	row, err := dbConnection.Query(sqlstatements.GET_BALANCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, hex.EncodeToString(pkhash))
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

func GetStateNonce(dbConnection *sql.DB, pkhash []byte) (uint64, error) {
	// search for pkhash's stateNonce
	row, err := dbConnection.Query(sqlstatements.GET_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, hex.EncodeToString(pkhash))
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

func GetAccountInfo(dbConnection *sql.DB, pkhash []byte) (*accountinfo.AccountInfo, error) {
	// retrieve pkhash's balance
	balance, err := GetBalance(dbConnection, pkhash)
	if err != nil {
		return nil, errors.New("Failed to retreive balance: " + err.Error())
	}

	// retrieve pkhash's stateNonce
	stateNonce, err := GetStateNonce(dbConnection, pkhash)
	if err != nil {
		return nil, errors.New("Failed to retreive stateNonce: " + err.Error())
	}

	return &accountinfo.AccountInfo{Balance: balance, StateNonce: stateNonce}, nil
}

/*calculates and inserts accounts' balance and nonce into the account balance table
  NOTE: the db connection passed in should be open
*/
func UpdateAccountTable(db *sql.DB, b *block.Block) error {

	//retrieve contracts
	contrcts := make([]*contracts.Contract, len(b.Data))
	for i, data := range b.Data {
		contrcts[i] = &contracts.Contract{}
		err := contrcts[i].Deserialize(data)
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
	for _, contract := range contrcts {
		var encodedContractSenderPublicKey []byte
		if contract.SenderPubKey != nil {
			var err error
			encodedContractSenderPublicKey, err = publickey.Encode(contract.SenderPubKey)
			if err != nil {
				return err
			}

		}
		addRecip := true
		addSender := true

		if contract.SenderPubKey == nil { // minting contracts
			minting = true
			err := InsertAccountIntoAccountBalanceTable(db, contract.RecipPubKeyHash, contract.Value)
			if err != nil {
				return err
			}
			continue
		}

		for i := 0; i < len(totalBalances); i++ {
			if bytes.Compare(totalBalances[i].accountPKH, hashing.New(encodedContractSenderPublicKey)) == 0 {
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
				accountInfo{accountPKH: hashing.New(encodedContractSenderPublicKey), balance: -1 * int64(contract.Value), nonce: 1})
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

			row, _ := db.Query(sqlstatements.GET_BALANCE_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH, hex.EncodeToString(acc.accountPKH))
			if row.Next() {
				row.Scan(&balance, &nonce) // retrieve balance and nonce from account_balances
				row.Close()

				// update balance and nonce
				_, err := db.Exec(sqlstatements.UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH,
					acc.balance+int64(balance), acc.nonce+uint64(nonce), hex.EncodeToString(acc.accountPKH))
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
