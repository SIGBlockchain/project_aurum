package contractstable

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
)

func InsertContractIntoContractsTable(dbConnection *sql.DB, c *contracts.Contract) error {
	// create prepared statement to insert into contracts table
	statement,err := dbConnection.Prepare(sqlstatements.INSERT_VALUES_INTO_CONTRACTS)
	if err != nil {
		return errors.New("Unable to prepare sql statement to insert into contracts table")
	}
	defer statement.Close()

	// get senders pub key hash-------------------------------------------------------------------------- work on this....
	encodedSenderPubKey,_ := publickey.Encode(c.SenderPubKey)
	_,err = statement.Exec(hashing.New(encodedSenderPubKey),c.RecipPubKeyHash,c.StateNonce)
	if err != nil{
		errors.New("Failed to execute statement to insert into contracts")
	}

	return nil
}

func InsertContractsIntoContractsTable(dbConnection *sql.DB, c []*contracts.Contract) error {
	// Not yet implemented
	return errors.New("Not yet implemented")
}

func GetContractsFromContractsTable(dbConn *sql.DB, senderPubKey *ecdsa.PublicKey, nonce uint64) ([]*contracts.Contract, error) {
	//Not yet implemented
	return nil, errors.New("Not yet implemented")
}
