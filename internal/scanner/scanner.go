package scanner

import "crypto/ecdsa"
import "errors"

// History struct contains list of metadata-objects describing
// contract associated with that public key and its hash
type History struct {
	Timestamp       int64
	SenderPubKey    *ecdsa.PublicKey
	RecipPubKeyHash []byte // 32 bytes
	Value           uint64
}

// ContractHistory return a list of metadata-objects describing every contract
// associated with that public key
func ContractHistory(publickey *ecdsa.PublicKey) (History, error) {
	return History{}, errors.New("Function has not been implemented")
}
