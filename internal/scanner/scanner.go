package scanner

import (
	"crypto/ecdsa"

	"github.com/SIGBlockchain/project_aurum/internal/ifaces"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

// ContractHistory struct contains  the wallet address of the sender and the recipient,
// value, and timestamp of the containing block
type SenderHistory struct {
	Timestamp              int64
	SenderWalletAddress    *ecdsa.PublicKey
	RecipientWalletAddress *ecdsa.PublicKey
	Value                  uint64
}



// GetContractHistory returns a list ofunc ontractHistory objects tht contain a history of contracts that
// are associated with the Sender Key and the hash of the Sender Key
func GetContractHistory(senderWalletAddress publickey.AurumPublicKey, scanner ifaces.IBlockchainStreamer) ([]ContractHistory, error) {
	block, err := scanner.GetBlockFromPublicKey(senderWalletAddress)
	if err != nil {
		return nil, err
	}

	return SenderHistory{
		Timestamp: block.Timestamp,
		SenderWalletAddress: senderWalletAddress.Key,
		RecipientWalletAddress: ,
		Value: ,
	}, nil
}
