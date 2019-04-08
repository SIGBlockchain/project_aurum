package accounts

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/SIGBlockchain/project_aurum/producer/src/block"

	"github.com/SIGBlockchain/project_aurum/producer/src/keys"
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
	SigLen       uint8  // len of the signature
	Signature    []byte // size of 22
	RecipPubKey  ecdsa.PublicKey
	Value        uint64
	Nonce        uint64
}

/*		CLIENT FUNCTION*************************** PRIVATE KEY IS OKAY
Fills struct fields with parameters given
(with the exception of the signature field)
Calls sign contract
Returns contract
*/
func MakeContract(version uint16, sender ecdsa.PrivateKey, recipient ecdsa.PublicKey, value uint64, nonce uint64) Contract {

	// private key is of size 40 bytes
	// public key is of size 32 bytes
	// total size of struct is 112 bytes
	c := Contract{
		Version:      version,
		SenderPubKey: sender.PublicKey,
		SigLen:       0,
		Signature:    nil,
		RecipPubKey:  recipient,
		Value:        value,
		Nonce:        nonce,
	}
	fmt.Println("ABOUT TO SIGN")
	c.SignContract(sender) // passing in the senders private key to get sig
	//unsafe.Sizeof(c)
	// fmt.Println("SIZE")
	// fmt.Println(unsafe.Sizeof(c))
	// fmt.Println("SIZE SENDER")
	// fmt.Println(unsafe.Sizeof(sender))

	// fmt.Println("SIZE recp")
	// fmt.Println(unsafe.Sizeof(recipient))

	// senderSlice := keys.EncodePublicKey(&c.SenderPubKey)
	// fmt.Println("SIZE send")
	// fmt.Println(unsafe.Sizeof(senderSlice))

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
func (c *Contract) SignContract(sender ecdsa.PrivateKey) {

	senderSlice := keys.EncodePublicKey(&c.SenderPubKey)
	recipSlice := keys.EncodePublicKey(&c.RecipPubKey)

	preSerial := make([]byte, 374)

	binary.LittleEndian.PutUint16(preSerial[0:2], c.Version)
	copy(preSerial[2:180], senderSlice)
	copy(preSerial[180:358], recipSlice)
	binary.LittleEndian.PutUint64(preSerial[358:366], c.Value)
	binary.LittleEndian.PutUint64(preSerial[366:374], c.Nonce)

	preHash := block.HashSHA256(preSerial)

	c.Signature, _ = sender.Sign(rand.Reader, preHash, nil)
	c.SigLen = uint8(len(c.Signature))

	fmt.Println("SIGNING******")
	fmt.Println("SIGN sender")
	fmt.Println(len(senderSlice))
	fmt.Println("recip slice")
	fmt.Println(len(recipSlice))
	fmt.Println("sig")
	fmt.Println(len(c.Signature))
	fmt.Println("sigLen")
	fmt.Println(c.SigLen)

	fmt.Println("AT ENDDD of signing****")
	fmt.Println(len(senderSlice) + len(recipSlice) + int(c.SigLen) + 2 + 16 + 1)

	fmt.Println(c.Signature)

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
*/
func ValidateContract(c Contract, tableName string) bool {
	fmt.Println("ABOUT TO VALID")

	// table, err := sql.Open("sqlite3", tableName)
	// if err != nil {
	// 	//"Failed to open sqlite3 table"
	// 	return false
	// }

	// defer table.Close()

	// //var pubKey string
	// rows, err := table.Query("SELECT public_key FROM acccount_balances")
	// if err != nil {
	// 	//"Failed to create rows to look for public key"
	// 	return false
	// }

	// //fmt.Println("ABOUT TO VALID")

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
Open table
Find public key
subtract from signing key
Fields:
public key, balance, nonce
*/
func UpdateAccountBalanceTable(table string) {}

// Serialize all fields of the contract
func (c Contract) Serialize() []byte {
	//senderSlice := keys.EncodePublicKey(&c.SenderPubKey) // THESE TWO SEEM TO CAUSE A PROBELM.... WHY!?
	// recipSlice := keys.EncodePublicKey(&c.RecipPubKey)   // size 178

	fmt.Println("INSIDE SERIALIZE")

	// // size of signature VARIES*********************
	// fmt.Println("SIGN send")
	//fmt.Println(len(senderSlice))
	// fmt.Println("rec")
	// fmt.Println(len(recipSlice))
	// fmt.Println("sig")
	// fmt.Println(len(c.Signature))
	// fmt.Println("sigLen")
	// fmt.Println(c.SigLen)

	// fmt.Printf("********TOTAL")
	// fmt.Println(len(senderSlice) + len(recipSlice) + int(c.SigLen) + 2 + 16 + 1)

	serializedContract := make([]byte, 447)

	// binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
	// copy(serializedContract[2:180], senderSlice)
	// serializedContract = append(serializedContract, c.SigLen)	//180:181
	// copy(serializedContract[181:(181+c.SigLen)], c.Signature)
	// copy(serializedContract[(181+c.SigLen):430], recipSlice)
	// binary.LittleEndian.PutUint64(serializedContract[430:438], c.Value)
	// binary.LittleEndian.PutUint64(serializedContract[438:446], c.Nonce)

	return serializedContract
}

// Deserialize into a struct
func (c Contract) Deserialize(b []byte) Contract {
	// c2 := Contract{
	// 	Version:      binary.LittleEndian.Uint16(b[0:2]),
	// 	SenderPubKey: *(keys.DecodePublicKey(b[2:26])),
	// 	Signature:    b[26:48],
	// 	RecipPubKey:  *(keys.DecodePublicKey(b[48:80])),
	// 	Value:        binary.LittleEndian.Uint64(b[80:88]),
	// 	Nonce:        binary.LittleEndian.Uint64(b[88:96]),
	// }
	return Contract{} //c2
}

/*
FOR 1ST TEST... VERIFY
appears that the account balances are not written so we cannot know if the correct amount is available in an account.
premade database to use? or add a balance in and use that?

NONCE HAS TO BE 1 + WHATS IN THE TABLE...
UPDATE TABLE WHEN VALIDATE CONTRACTS IS TRUE (VALIDATE CONTRACTS)		// PASS CONTRACT INTO UPDATE FUNCTION


*/
