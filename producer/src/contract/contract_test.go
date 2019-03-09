package contract

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	block "../block"
)

func TestMakeYield(t *testing.T) {
	testPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testPubKey := testPrivKey.PublicKey
	pubKeyBytes := testPubKey.X.Bytes()
	pubKeyBytes = append(pubKeyBytes, testPubKey.Y.Bytes()...)
	hashedKey := block.HashSHA256(pubKeyBytes)
	var expectedRecipient []byte = hashedKey[:]
	expectedValue := 1000000
	actual := MakeYield(testPubKey, 1000000)
	if bytes.Equal(expectedRecipient, actual.Recipient) {
		return t.Errorf("Recipients do not match")
	}
	if expectedValue != actual.Value {
		return t.Errorf("Values do not match")
	}
}
