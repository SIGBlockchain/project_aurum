package contractstable

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

func InsertContractIntoContractsTable(dbConnection *sql.DB, c *contracts.Contract) error {
	// create prepared statement to insert into contracts table
	statement, err := dbConnection.Prepare(sqlstatements.INSERT_VALUES_INTO_CONTRACTS)
	serializedContract, _ := c.Serialize()
	encodedSenderPubKey, _ := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return errors.New("Unable to prepare sql statement to insert into contracts table")
	}
	defer statement.Close()

	_, err = statement.Exec(serializedContract, hex.EncodeToString(hashing.New(encodedSenderPubKey)), c.StateNonce)
	if err != nil {
		errors.New("Failed to execute statement to insert into contracts")
	}

	return nil
}

func InsertContractsIntoContractsTable(dbConnection *sql.DB, contracts []*contracts.Contract) error {
	for _, c := range contracts {
		err := InsertContractIntoContractsTable(dbConnection, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetContractsFromContractsTable(dbConn *sql.DB, senderPubKey *ecdsa.PublicKey, nonce uint64) ([]*contracts.Contract, error) {
	//Not yet implemented
	return nil, errors.New("Not yet implemented")
}
