package contractstable

import (
	"crypto/ecdsa"
	"database/sql"
	"errors"

	"github.com/SIGBlockchain/project_aurum/internal/contracts"
)

func InsertContractIntoContractsTable(c *contracts.Contract) error {
	// Not yet implemented
	return errors.New("Not yet implemented")
}

func InsertContractsIntoContractsTable(c []*contracts.Contract) error {
	// Not yet implemented
	return errors.New("Not yet implemented")
}

func GetContractsFromContractsTable(dbConn *sql.DB, senderPubKey *ecdsa.PublicKey, nonce uint64) ([]*contracts.Contract, error) {
	//Not yet implemented
	return nil, errors.New("Not yet implemented")
}
