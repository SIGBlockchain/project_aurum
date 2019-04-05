package accounts

import (
	"crypto/ecdsa"
	"fmt"
	"unsafe"
)

/*
Version
Sender Public Key
Signature
Recipient Public Key Hash
Value
Nonce
*/
type Contract struct {
	Version      uint16
	SenderPubKey ecdsa.PublicKey
	Signature    []byte
	RecipPubKey  ecdsa.PublicKey
	Value        uint64
	Nonce        uint64
}

/*
Fills struct fields with parameters given
(with the exception of the signature field)
Calls sign contract
Returns contract
*/
func MakeContract(version uint16, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64, nonce uint64) Contract {

	c := Contract{
		Version:      version,
		SenderPubKey: sender.PublicKey,
		Signature:    nil,
		RecipPubKey:  recipient,
		Value:        value,
		Nonce:        nonce,
	}
	c.SignContract()
	//unsafe.Sizeof(c)
	fmt.Println(unsafe.Sizeof(c))
	return c
}

/*
Take the contract by reference
(with a null signature field)
Serialize it
Hash it
Generate a signature with the sender public key
Update the signature field
*/
func (c *Contract) SignContract() {

}

/*
Check balance (ideal scenario):
Open table
Get hash of contract
Verify signature with hash and public key
Go to table and find sender
Confirm balance is sufficient
Update Account Balances (S & R)
Increment Table Nonce
*/
func ValidateContract(c Contract, tableName string) bool {
	if true && false {
		return true
	}
	return false
}

/*
Open table
Find public key
subtract from signing key
Fields:
public key, balance, nonce
*/
func UpdateAccountBalanceTable(table string) {}

// Serialize all fields of the contract
func (c Contract) Serialize() []byte {

	return []byte{}
}

// Deserialize into a struct
func (c Contract) Deserialize(b []byte) Contract {
	return Contract{}
}
