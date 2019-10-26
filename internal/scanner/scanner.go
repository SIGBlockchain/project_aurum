package scanner

import (
	"crypto/ecdsa"
	"errors"
)

// ContractHistory struct contains  the wallet address of the sender and the recipient,
// value, and timestamp of the containing block
type ContractHistory struct {
	Timestamp              int64
	SenderWalletAddress    *ecdsa.PublicKey
	RecipientWalletAddress *ecdsa.PublicKey
	Value                  uint64
}

// ScanContractHistory returns a list of ContractHistory objects tht contain a history of contracts that
// are associated with the Sender Key and the hash of the Sender Key
func ScanContractHistory() (ContractHistory, error) {
	return ContractHistory{}, errors.New("function not implemented")
}
