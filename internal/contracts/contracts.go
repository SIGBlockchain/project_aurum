package contracts

import (
	"bytes"
	"crypto/ecdsa"
	mrand "math/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	
 
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

/*
Version
Sender Public Key
Signature Length
Signature
Recipient Public Key Hash
Value
*/
type Contract struct {
	Version         uint16
	SenderPubKey    *ecdsa.PublicKey
	SigLen          uint8  // len of the signature
	Signature       []byte // size varies
	RecipPubKeyHash []byte // 32 bytes
	Value           uint64
	StateNonce      uint64
}

type JSONContract struct {
	Version                uint16
	SenderPublicKey        string
	SignatureLength        uint8
	Signature              string
	RecipientWalletAddress string
	Value                  uint64
	StateNonce             uint64
}

/*
version field comes from version parameter
sender public key comes from sender private key
signature comes from calling sign contract
signature length comes from signature
recipient pk hash comes from sha-256 hash of rpk
value is value parameter
returns contract struct
*/
func New(version uint16, sender *ecdsa.PrivateKey, recipient []byte, value uint64, nextStateNonce uint64) (*Contract, error) {

	if version == 0 {
		return nil, errors.New("Invalid version; must be >= 1")
	}               

	c := Contract{
		Version:         version,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: recipient,
		Value:           value,
		StateNonce:      nextStateNonce,
	}

	if sender == nil {
		c.SenderPubKey = nil
	} else {
		c.SenderPubKey = &(sender.PublicKey)
	}

	return &c, nil
}

// // Serialize all fields of the contract
func (c *Contract) Serialize() ([]byte, error) {
	/*
		0-2 version
		2-180 spubkey
		180-181 siglen
		181 - 181+c.siglen signature
		181+c.siglen - (181+c.siglen + 32) rpkh
		(181+c.siglen + 32) - (181+c.siglen + 32+ 8) value

	*/

	// if contract's sender pubkey is nil, make 178 zeros in its place instead
	var spubkey []byte
	var err error
	if c.SenderPubKey == nil {
		spubkey = make([]byte, 178)
	} else {
		spubkey, err = publickey.Encode(c.SenderPubKey) //size 178
		if err != nil {
			return nil, err
		}
	}

	//unsigned contract
	if c.SigLen == 0 {
		totalSize := (2 + 178 + 1 + 32 + 8 + 8)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = 0
		copy(serializedContract[181:213], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[213:221], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[221:229], c.StateNonce)

		return serializedContract, nil
	} else { //signed contract
		totalSize := (2 + 178 + 1 + int(c.SigLen) + 32 + 8 + 8)
		serializedContract := make([]byte, totalSize)
		binary.LittleEndian.PutUint16(serializedContract[0:2], c.Version)
		copy(serializedContract[2:180], spubkey)
		serializedContract[180] = c.SigLen
		copy(serializedContract[181:(181+int(c.SigLen))], c.Signature)
		copy(serializedContract[(181+int(c.SigLen)):(181+int(c.SigLen)+32)], c.RecipPubKeyHash)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32):(181+int(c.SigLen)+32+8)], c.Value)
		binary.LittleEndian.PutUint64(serializedContract[(181+int(c.SigLen)+32+8):(181+int(c.SigLen)+32+8+8)], c.StateNonce)

		return serializedContract, nil
	}
}

// Deserialize into a struct
func (c *Contract) Deserialize(b []byte) error {
	var spubkeydecoded *ecdsa.PublicKey
	var err error

	// if serialized sender public key contains only zeros, sender public key is nil
	if bytes.Equal(b[2:180], make([]byte, 178)) {
		spubkeydecoded = nil
	} else {
		spubkeydecoded, err = publickey.Decode(b[2:180])
		if err != nil {
			return err
		}
	}
	siglen := int(b[180])

	// unsigned contract
	if siglen == 0 {
		c.Version = binary.LittleEndian.Uint16(b[0:2])
		c.SenderPubKey = spubkeydecoded
		c.SigLen = b[180]
		c.RecipPubKeyHash = b[181:213]
		c.Value = binary.LittleEndian.Uint64(b[213:221])
		c.StateNonce = binary.LittleEndian.Uint64(b[221:229])
	} else {
		c.Version = binary.LittleEndian.Uint16(b[0:2])
		c.SenderPubKey = spubkeydecoded
		c.SigLen = b[180]
		c.Signature = b[181:(181 + siglen)]
		c.RecipPubKeyHash = b[(181 + siglen):(181 + siglen + 32)]
		c.Value = binary.LittleEndian.Uint64(b[(181 + siglen + 32):(181 + siglen + 32 + 8)])
		c.StateNonce = binary.LittleEndian.Uint64(b[(181 + siglen + 32 + 8):(181 + siglen + 32 + 8 + 8)])
	}
	return nil
}

/*
hashed contract = sha 256 hash ( version + spubkey + rpubkeyhash + value)
signature = Sign ( hashed contract, sender private key )
sig len = signature length
siglen and sig go into respective fields in contract
*/
func (c *Contract) Sign(sender *ecdsa.PrivateKey) error {
	serializedTestContract, err := c.Serialize()
	if err != nil {
		return errors.New("Failed to serialize contract")
	}
	hashedContract := hashing.New(serializedTestContract)
	c.Signature, _ = sender.Sign(rand.Reader, hashedContract, nil)
	c.SigLen = uint8(len(c.Signature))
	return nil
}

// compare two contracts and return true only if all fields match
func (contract1 *Contract) Equals(contract2 Contract) bool {
	// copy both contracts
	c1val := reflect.ValueOf(*contract1)
	c2val := reflect.ValueOf(contract2)

	// loops through fields
	for i := 0; i < c1val.NumField(); i++ {
		finterface1 := c1val.Field(i).Interface() // value assignment from c1 as interface
		finterface2 := c2val.Field(i).Interface() // value assignment from c2 as interface

		switch finterface1.(type) { // switch on type
		case uint8, uint16, uint64, int64:
			if finterface1 != finterface2 {
				return false
			}
		case []byte:
			if !bytes.Equal(finterface1.([]byte), finterface2.([]byte)) {
				return false
			}
		case [][]byte:
			for i := 0; i < len(finterface1.([][]byte)); i++ {
				if !bytes.Equal(finterface1.([][]byte)[i], finterface2.([][]byte)[i]) {
					return false
				}
			}
		case *ecdsa.PublicKey:
			if !reflect.DeepEqual(finterface1, finterface2) {
				return false
			}
		}
	}
	return true
}

// ToString takes in a Contract and return a string version
func (c Contract) ToString() string {
	encodedSenderPublicKey, err := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return "Error on encoding Sender Public Key"
	}
	return fmt.Sprintf("Version: %v\nSenderPubKey: %v\nSigLen: %v\nSignature: %v\nRecipPubKeyHash: %v\nValue: %v\nStateNonce: %v\n",
		c.Version, hex.EncodeToString(encodedSenderPublicKey), c.SigLen, hex.EncodeToString(c.Signature), hex.EncodeToString(c.RecipPubKeyHash),
		c.Value, c.StateNonce)
}

// Marshal takes a Contract and returns a JSONContract
func (c *Contract) Marshal() (JSONContract, error) {
	encodedSender, err := publickey.Encode(c.SenderPubKey)
	if err != nil {
		return JSONContract{}, errors.New("Failed to encode sender pubkey: " + err.Error())
	}

	var newJSONContract = JSONContract{
		Version:                c.Version,
		SenderPublicKey:        hex.EncodeToString(encodedSender),
		SignatureLength:        c.SigLen,
		Signature:              hex.EncodeToString(c.Signature),
		RecipientWalletAddress: hex.EncodeToString(c.RecipPubKeyHash),
		Value:                  c.Value,
		StateNonce:             c.StateNonce,
	}

	return newJSONContract, nil
}

// Unmarshal takes a JSONContract and returns a Contract
func (mc *JSONContract) Unmarshal() (Contract, error) {
	encodedSender, err := hex.DecodeString(mc.SenderPublicKey)
	if err != nil {
		return Contract{}, errors.New("Failed to decode sender string: " + err.Error())
	}
	senderPB, err := publickey.Decode(encodedSender)
	if err != nil {
		return Contract{}, errors.New("Failed to decode sender: " + err.Error())
	}
	signature, err := hex.DecodeString(mc.Signature)
	if err != nil {
		return Contract{}, errors.New("Failed to decode signature: " + err.Error())
	}
	recip, err := hex.DecodeString(mc.RecipientWalletAddress)
	if err != nil {
		return Contract{}, errors.New("Failed to decode recipient wallet address: " + err.Error())
	}

	c := Contract{
		mc.Version,
		senderPB,
		mc.SignatureLength,
		signature,
		recip,
		mc.Value,
		mc.StateNonce,
	}
	return c, nil
	
}



func GenerateRandomContract() (*Contract){
	min := 1
	maxVer := 65535
	b := make([]byte, 32)
	mrand.Read(b)
	genRecipPubKeyHash := b 
	genVersion := mrand.Intn(maxVer - min) + min
	genValue := mrand.Uint64()+uint64(min)
	genStateNonce := mrand.Uint64()+uint64(min)
	
	
	
	c:= &Contract{
		RecipPubKeyHash: genRecipPubKeyHash,
		Version: uint16(genVersion),
		Value: genValue,
		StateNonce: genStateNonce,
	
	}
	return c
	
}
	


