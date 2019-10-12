package scanner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
)

func TestContractHistory(t *testing.T) {
	// get time stamp
	ti := time.Now()
	nowTime := ti.UnixNano()

	// create a sender wallet address
	senderWalletAddress, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// create a recipient wallet address
	recipientWalletAddress, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientKey, _ := publickey.Encode(&recipientWalletAddress.PublicKey)

	// create a test block
	contractHistory := History{
		Timestamp:       nowTime,
		SenderPubKey:    &senderWalletAddress.PublicKey,
		RecipPubKeyHash: hashing.New(encodedRecipientKey),
		Value:           90001,
	}

	// create a series of blocks for testing
	blocks := make([]History, 7)
	for i := 0; i < 7; i++ {
		blocks[i] = contractHistory
	}

	// call the function
	resultHistoryScanner, err := ContractHistory(&senderWalletAddress.PublicKey)
	if err != nil {
		t.Errorf("An error occured when calling the function ContractHistory()")
	}

	if contractHistory.Timestamp != resultHistoryScanner.Timestamp {
		t.Errorf("Result timestamp does not match the expected timestamp!")
	}

	if !reflect.DeepEqual(contractHistory.SenderPubKey, resultHistoryScanner.SenderPubKey) {
		t.Errorf("Result sender public key does not match expected sender public key!")
	}

	if !reflect.DeepEqual(contractHistory.RecipPubKeyHash, resultHistoryScanner.RecipPubKeyHash) {
		t.Errorf("Result recipient key does not match expected receipient key!")
	}

	if contractHistory.Value != resultHistoryScanner.Value {
		t.Errorf("Result value does not match expected value!")
	}
}
