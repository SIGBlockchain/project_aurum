package accounts

import (
	"crypto/ecdsa"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"

	"github.com/SIGBlockchain/project_aurum/producer/src/keys"
)

/*
Version
Sender Public Key
Signature Length
Signature
Recipient Public Key Hash
Value
Nonce
*/
type Contract struct {
	Version         uint16
	SenderPubKey    ecdsa.PublicKey
	SigLen          uint8  // len of the signature
	Signature       []byte // size varies
	RecipPubKeyHash []byte // 32 bytes
	Value           uint64
	Nonce           uint64
}

// 		CLIENT FUNCTION*************************** PRIVATE KEY IS OKAY
// Fills struct fields with parameters given
// (with the exception of the signature field)
// Calls sign contract
// Returns contract

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

	// private key is of size 40 bytes
	// public key is of size 32 bytes
	// total size of struct is 112 bytes
	c := Contract{
		Version:         version,
		SenderPubKey:    sender.PublicKey,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&recipient)), // size 32 bytes
		Value:           value,
		Nonce:           nonce,
	}
	c.SignContract(sender) // passing in the senders private key to get sig

	return c
}

/*
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
signature = Sign ( hashed contract, sender private key )
sig len = signature length
siglen and sig go into respective fields in contract
*/
func (c *Contract) SignContract(sender ecdsa.PrivateKey) {

	spubkey := keys.EncodePublicKey(&c.SenderPubKey)

	// fmt.Println(len(c.RecipPubKeyHash))

	preSerial := make([]byte, 374)

	binary.LittleEndian.PutUint16(preSerial[0:2], c.Version)   // 2
	copy(preSerial[2:180], spubkey)                            //178
	copy(preSerial[180:212], c.RecipPubKeyHash)                //32
	binary.LittleEndian.PutUint64(preSerial[212:220], c.Value) //8
	binary.LittleEndian.PutUint64(preSerial[220:228], c.Nonce) //8

	preHash := block.HashSHA256(preSerial)

	c.Signature, _ = sender.Sign(rand.Reader, preHash, nil)
	c.SigLen = uint8(len(c.Signature))
}

/*
Check balance (ideal scenario):
Open table
Get hash of contract
Verify signature with hash and public key
Go to table and find sender
Confirm balance is sufficient
Update Account Balances (S & R)		// ony updated when true
Increment Table Nonce

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
	// fmt.Println("ABOUT TO VALID")

	table, err := sql.Open("sqlite3", tableName)
	if err != nil {
		//"Failed to open sqlite3 table"
		return false
	}

	defer table.Close()
	// statement, err := table.Prepare(
	// 	`CREATE TABLE IF NOT EXISTS account_balances (
	// 	public_key_hash TEXT,
	// 	balance INTEGER,
	// 	nonce INTEGER);`)

	// if err != nil {
	// 	return false // error stmt?
	// }
	// statement.Exec()
	// table.Close()

	// 	1. verify signature
	// hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value + nonce )
	// verify (hashed contract, spubkey, signature) (T)

	spubkey := keys.EncodePublicKey(&c.SenderPubKey)

	preSerial := make([]byte, 374)

	binary.LittleEndian.PutUint16(preSerial[0:2], c.Version)   // 2
	copy(preSerial[2:180], spubkey)                            //178
	copy(preSerial[180:212], c.RecipPubKeyHash)                //32
	binary.LittleEndian.PutUint64(preSerial[212:220], c.Value) //8
	binary.LittleEndian.PutUint64(preSerial[220:228], c.Nonce) //8

	hashedContract := block.HashSHA256(preSerial)
	if ecdsa.Verify(&c.SenderPubKey, hashedContract, c.SenderPubKey.X, c.SenderPubKey.Y) {
		fmt.Println("ecdsa.verify true")
	}

	// //var pubKey string
	// rows, err := table.Query("SELECT public_key_hash FROM acccount_balances")
	// if err != nil {
	// 	//"Failed to create rows to look for public key"
	// 	return false
	// }

	// var pk string
	// for rows.Next() {
	// 	rows.Scan(&pk)
	// 	if pk == string(keys.EncodePublicKey(&c.SenderPubKey)) {
	// 		return true
	// 	}
	// 	fmt.Printf(pk)
	// 	fmt.Println(" pubKey")

	// }
	// if true && false {
	// 	return true
	// }
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
	/*
		0-2 version
		2-180 spubkey
		180-181 siglen
		181 - 181+c.siglen signature
		181+c.siglen - (181+c.siglen + 32) rpkh
		(181+c.siglen + 32) - (181+c.siglen + 32+ 8) value
		(181+c.siglen + 32+ 8) - (181+c.siglen + 32 + 8 + 8) nonce

	*/

	spubkey := keys.EncodePublicKey(&c.SenderPubKey) //size 178
	totalSize := (2 + 178 + 1 + int(c.SigLen) + 32 + 16)
	serializedContract := make([]byte, totalSize)

	binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
	copy(serializedContract[2:180], spubkey)
	serializedContract[180] = c.SigLen
	copy(serializedContract[181:(181+int(c.SigLen))], c.Signature)
	copy(serializedContract[(181+int(c.SigLen)):(181+int(c.SigLen)+32)], c.RecipPubKeyHash)
	binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32):(181+int(c.SigLen)+32+8)], c.Value)
	binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32+8):(181+int(c.SigLen)+32+8+8)], c.Nonce)

	return serializedContract
}

// Deserialize into a struct
func (c Contract) Deserialize(b []byte) Contract {
	spubkeydecoded := keys.DecodePublicKey(b[2:180])
	siglen := int(b[180])

	fmt.Println("siglen")
	fmt.Println(siglen)
	fmt.Println("sig")
	fmt.Println(b[181 : 181+siglen])
	fmt.Println("rpkh")
	fmt.Println(b[181+int(siglen) : 181+int(siglen)+32])
	fmt.Println("val")
	fmt.Println(b[181+int(siglen)+32 : 181+int(siglen)+32+8])
	fmt.Println("nonce")
	fmt.Println(b[181+int(siglen)+32+8 : 181+int(siglen)+32+8+8])

	c2 := Contract{
		Version:         binary.LittleEndian.Uint16(b[0:2]),
		SenderPubKey:    *spubkeydecoded,
		SigLen:          b[180],
		Signature:       b[181:(181 + siglen)],
		RecipPubKeyHash: b[(181 + siglen):(181 + siglen + 32)],
		Value:           binary.LittleEndian.Uint64(b[(181 + siglen + 32):(181 + siglen + 32 + 8)]),
		Nonce:           binary.LittleEndian.Uint64(b[(181 + siglen + 32 + 8):(181 + siglen + 32 + 8 + 8)]),
	}
	fmt.Println(c2.SenderPubKey)
	fmt.Println(c2.SigLen)
	fmt.Println(c2.Signature)
	fmt.Println(c2.RecipPubKeyHash)
	fmt.Println(c2.Value)
	fmt.Println(c2.Nonce)
	return c2
}

/*
FOR 1ST TEST... VERIFY
appears that the account balances are not written so we cannot know if the correct amount is available in an account.
premade database to use? or add a balance in and use that?

NONCE HAS TO BE 1 + WHATS IN THE TABLE...
UPDATE TABLE WHEN VALIDATE CONTRACTS IS TRUE (VALIDATE CONTRACTS)		// PASS CONTRACT INTO UPDATE FUNCTION


*/
