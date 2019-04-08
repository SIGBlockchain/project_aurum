package accounts

import "crypto/ecdsa"

/*
Version
Sender Public Key
Signature Length
Signature
Recipient Public Key Hash
Value
Nonce
*/
type Contract struct{}

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
func MakeContract(version uint16, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64, nonce uint64) Contract {
	return Contract{}
}

/*
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
signature = Sign ( hashed contract, sender private key )
sig len = signature length
siglen and sig go into respective fields in contract
*/
func (c *Contract) SignContract() {}

/*
1. verify signature
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
verify (hashed contract, spubkey, signature) (T)
2. validate amount
check table to see if sender's balance >= contract amount (T)
3. validate nonce
check to see that nonce is 1 + table nonce for that account (T)

If all 3 are true, update table
*/
func ValidateContract(c Contract, tableName string) bool {
	if true && false {
		return true
	}
	return false
}

/*
spkh = sha256 ( serialized sender pub key )
find sender public key hash
decrease value from sender public key hash account
increment their nonce by one
increase value of recipient public key hash account by contract value
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
