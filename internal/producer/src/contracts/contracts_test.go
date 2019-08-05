package contracts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

func TestNew(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	type args struct {
		version       uint16
		sender        *ecdsa.PrivateKey
		recipient     []byte
		value         uint64
		newStateNonce uint64
	}
	tests := []struct {
		name    string
		args    args
		want    *Contract
		wantErr bool
	}{
		{
			name: "Unsigned Minting contract",
			args: args{
				version:       1,
				sender:        nil,
				recipient:     hashing.New(encodedPublicKey),
				value:         1000000000,
				newStateNonce: 1,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    nil,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: hashing.New(encodedPublicKey),
				Value:           1000000000,
				StateNonce:      1,
			},
			wantErr: false,
		},
		{
			name: "Unsigned Normal contract",
			args: args{
				version:       1,
				sender:        senderPrivateKey,
				recipient:     hashing.New(encodedRecipientKey),
				value:         1000000000,
				newStateNonce: 1,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    &senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: hashing.New(encodedRecipientKey),
				Value:           1000000000,
				StateNonce:      1,
			},
			wantErr: false,
		},
		{
			name: "Version 0 contract",
			args: args{
				version:       0,
				sender:        senderPrivateKey,
				recipient:     hashing.New(encodedPublicKey),
				value:         1000000000,
				newStateNonce: 1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.version, tt.args.sender, tt.args.recipient, tt.args.value, tt.args.newStateNonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContract_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	nullSenderContract, _ := New(1, nil, hashing.New(encodedPublicKey), 1000, 0)
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	unsignedContract, _ := New(1, senderPrivateKey, hashing.New(encodedRecipientKey), 1000, 0)
	signedContract, _ := New(1, senderPrivateKey, hashing.New(encodedRecipientKey), 1000, 0)
	signedContract.Sign(senderPrivateKey)
	tests := []struct {
		name string
		c    *Contract
	}{
		{
			name: "Minting contract",
			c:    nullSenderContract,
		},
		{
			name: "Unsigned contract",
			c:    unsignedContract,
		},
		{
			name: "Signed contract",
			c:    signedContract,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.c.Serialize()
			sigLen := got[180]
			testSendPublicKey := tt.c.SenderPubKey
			testEncodeSenderPubKey, _ := publickey.Encode(testSendPublicKey)
			switch tt.name {
			case "Minting contract":
				if !bytes.Equal(got[2:180], make([]byte, 178)) {
					t.Errorf("Non null sender public key for minting contract")
				}
				if sigLen != 0 {
					t.Errorf("Non-zero signature length in minting contract: %v", sigLen)
				}
				if !bytes.Equal(got[181:213], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in minting contract")
				}
				break
			case "Unsigned contract":
				if sigLen != 0 {
					t.Errorf("Non-zero signature length in unsigned contract: %v", sigLen)
				}
				if !bytes.Equal(got[2:180], testEncodeSenderPubKey) {
					t.Errorf("Invalid encoded public key for unsigned contract")
				}
				if !bytes.Equal(got[181:213], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in unsigned contract")
				}
			case "Signed Contract":
				if sigLen == 0 {
					t.Errorf("Zero length signature in signed contract: %v", sigLen)
				}
				if !bytes.Equal(got[2:180], testEncodeSenderPubKey) {
					t.Errorf("Invalid encoded public key for signed contract")
				}
				if !bytes.Equal(got[(181+int(sigLen)):(181+int(sigLen)+32)], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in signed contract")
				}
			}
		})
	}
}

func TestContract_Deserialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	nullSenderContract, _ := New(1, nil, hashing.New(encodedPublicKey), 1000, 1)
	nullSenderContractSerialized, _ := nullSenderContract.Serialize()
	encodedRecipientKey, _ := publickey.Encode(&recipientPrivateKey.PublicKey)
	unsignedContract, _ := New(1, senderPrivateKey, hashing.New(encodedRecipientKey), 1000, 1)
	unsignedContractSerialized, _ := unsignedContract.Serialize()
	signedContract, _ := New(1, senderPrivateKey, hashing.New(encodedRecipientKey), 1000, 1)
	signedContract.Sign(senderPrivateKey)
	signedContractSerialized, _ := signedContract.Serialize()
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			name: "Minting contract",
			c:    &Contract{},
			args: args{
				nullSenderContractSerialized,
			},
		},
		{
			name: "Unsigned contract",
			c:    &Contract{},
			args: args{
				unsignedContractSerialized,
			},
		},
		{
			name: "Signed contract",
			c:    &Contract{},
			args: args{
				signedContractSerialized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Deserialize(tt.args.b)
			switch tt.name {
			case "Minting contract":
				if tt.c.Version != nullSenderContract.Version {
					t.Errorf("Invalid field on nullSender contract: version")
				}
				if tt.c.SigLen != nullSenderContract.SigLen {
					t.Errorf("Invalid field on nullSender contract: signature length")
				}
				if tt.c.Value != nullSenderContract.Value {
					t.Errorf("Invalid field on nullSender contract: value")
				}
				if tt.c.Signature != nil {
					t.Errorf("Invalid field on nullSender contract: signature")
				}
				if tt.c.SenderPubKey != nil {
					t.Errorf("Invalid field on nullSender contract: sender public key")
				}
				if tt.c.StateNonce != nullSenderContract.StateNonce {
					t.Errorf(fmt.Sprintf("Invalid field on nullSender contract: state nonce. Want: %d, got %d", nullSenderContract.StateNonce, tt.c.StateNonce))
				}
				break
			case "Unsigned contract":
				if tt.c.Version != unsignedContract.Version {
					t.Errorf("Invalid field on unsigned contract: version")
				}
				if tt.c.SigLen != unsignedContract.SigLen {
					t.Errorf("Invalid field on unsigned contract: signature length")
				}
				if tt.c.Value != unsignedContract.Value {
					t.Errorf("Invalid field on unsigned contract: value")
				}
				if tt.c.Signature != nil {
					t.Errorf("Invalid field on unsigned contract: signature")
				}
				if !reflect.DeepEqual(tt.c.SenderPubKey, &senderPrivateKey.PublicKey) {
					t.Errorf("Invalid field on unsigned contract: sender public key")
				}
				if tt.c.StateNonce != unsignedContract.StateNonce {
					t.Errorf("Invalid field on unsigned contract: state nonce")
				}
				break
			case "Signed contract":
				if tt.c.Version != signedContract.Version {
					t.Errorf("Invalid field on signed contract: version")
				}
				if tt.c.SigLen != signedContract.SigLen {
					t.Errorf("Invalid field on signed contract: signature length")
				}
				if tt.c.Value != signedContract.Value {
					t.Errorf("Invalid field on signed contract: value")
				}
				if !bytes.Equal(tt.c.Signature, signedContract.Signature) {
					t.Errorf("Invalid field on signed contract: signature")
				}
				if !reflect.DeepEqual(tt.c.SenderPubKey, &senderPrivateKey.PublicKey) {
					t.Errorf("Invalid field on signed contract: sender public key")
				}
				if tt.c.StateNonce != signedContract.StateNonce {
					t.Errorf("Invalid field on signed contract: state nonce")
				}
				break
			}
		})
	}
}

func TestContract_Sign(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	testContract, _ := New(1, senderPrivateKey, hashing.New(encodedPublicKey), 1000, 0)
	type args struct {
		sender ecdsa.PrivateKey
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			c: testContract,
			args: args{
				sender: *senderPrivateKey,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyOfContract := testContract
			serializedTestContract, _ := copyOfContract.Serialize()
			hashedContract := hashing.New(serializedTestContract)
			tt.c.Sign(&tt.args.sender)
			var esig struct {
				R, S *big.Int
			}
			if _, err := asn1.Unmarshal(tt.c.Signature, &esig); err != nil {
				t.Errorf("Failed to unmarshall signature")
			}
			if !ecdsa.Verify(tt.c.SenderPubKey, hashedContract, esig.R, esig.S) {
				t.Errorf("Failed to verify valid signature")
			}
			maliciousPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if ecdsa.Verify(&maliciousPrivateKey.PublicKey, hashedContract, esig.R, esig.S) {
				t.Errorf("Failed to reject invalid signature")
			}
		})
	}
}

func TestEquals(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	contract1 := Contract{
		Version:         1,
		SenderPubKey:    &senderPrivateKey.PublicKey,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: hashing.New(encodedPublicKey),
		Value:           1000000000,
		StateNonce:      1,
	}

	contracts := make([]Contract, 7)
	for i := 0; i < 7; i++ {
		contracts[i] = contract1
	}
	contracts[0].Version = 9001
	anotherSenderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	contracts[1].SenderPubKey = &anotherSenderPrivateKey.PublicKey
	contracts[2].SigLen = 9
	contracts[3].Signature = make([]byte, 100)
	encodedAnotherSenderPublicKey, _ := publickey.Encode(&anotherSenderPrivateKey.PublicKey)
	contracts[4].RecipPubKeyHash = hashing.New(encodedAnotherSenderPublicKey)
	contracts[5].Value = 9002
	contracts[6].StateNonce = 9

	tests := []struct {
		name string
		c1   Contract
		c2   Contract
		want bool
	}{
		{
			name: "equal contracts",
			c1:   contract1,
			c2:   contract1,
			want: true,
		},
		{
			name: "different contract version",
			c1:   contract1,
			c2:   contracts[0],
			want: false,
		},
		{
			name: "different contract SenderPubKey",
			c1:   contract1,
			c2:   contracts[1],
			want: false,
		},
		{
			name: "different contract signature lengths",
			c1:   contract1,
			c2:   contracts[2],
			want: false,
		},
		{
			name: "different contract signatures",
			c1:   contract1,
			c2:   contracts[3],
			want: false,
		},
		{
			name: "different contract RecipPubKeyHash",
			c1:   contract1,
			c2:   contracts[4],
			want: false,
		},
		{
			name: "different contract Values",
			c1:   contract1,
			c2:   contracts[5],
			want: false,
		},
		{
			name: "different contract StateNonce",
			c1:   contract1,
			c2:   contracts[6],
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := Equals(tt.c1, tt.c2); result != tt.want {
				t.Errorf("Error: Equals() returned %v for %s\n Wanted: %v", result, tt.name, tt.want)
			}
		})
	}

}

func TestContractToString(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, _ := publickey.Encode(&senderPrivateKey.PublicKey)
	testContract := Contract{
		Version:         1,
		SenderPubKey:    &senderPrivateKey.PublicKey,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: hashing.New(encodedSenderPublicKey),
		Value:           1000000000,
		StateNonce:      1,
	}
	encodedTestContractSenderPublicKey, _ := publickey.Encode(testContract.SenderPubKey)
	stringOfTheContract := fmt.Sprintf("Version: %v\nSenderPubKey: %v\nSigLen: %v\nSignature: %v\nRecipPubKeyHash: %v\nValue: %v\nStateNonce: %v\n", testContract.Version,
		hex.EncodeToString(encodedTestContractSenderPublicKey), testContract.SigLen, hex.EncodeToString(testContract.Signature),
		hex.EncodeToString(testContract.RecipPubKeyHash), testContract.Value, testContract.StateNonce)

	if result := testContract.ToString(); result != stringOfTheContract {
		t.Error("Contract String is not equal to test String")
	}
}
