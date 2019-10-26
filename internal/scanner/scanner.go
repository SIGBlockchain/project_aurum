package scanner

import (
	"crypto/ecdsa"
	"errors"

	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

// ContractHistory struct contains  the wallet address of the sender and the recipient,
// value, and timestamp of the containing block
type ContractHistory struct {
	Timestamp              int64
	SenderWalletAddress    *ecdsa.PublicKey
	RecipientWalletAddress *ecdsa.PublicKey
	Value                  uint64
}

// GetContractHistory returns a list of ContractHistory objects tht contain a history of contracts that
// are associated with the Sender Key and the hash of the Sender Key
func GetContractHistory(senderWalletAddress publickey.AurumPublicKey, blockchain interface{}) ([]ContractHistory, error) {
	return blockchain.ScanBlockChain(senderWalletAddress), errors.New("function not implemented")
}
