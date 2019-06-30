package validation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/accounts/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/producer/src/hashing"
	"github.com/SIGBlockchain/project_aurum/pkg/publickey"
	_ "github.com/mattn/go-sqlite3"
)

// Test cases for validation (next issue)
//// Zero value contracts
//// Minting contracts
//// Invalid signature contracts
//// Insufficient balance contracts
//// Completely valid contract

func TestValidateContract(t *testing.T) {
	dbName := constants.AccountsTable
	dbc, _ := sql.Open("sqlite3", dbName)
	defer func() {
		err := dbc.Close()
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
		err = os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPKH := hashing.New(publickey.Encode(&sender.PublicKey))
	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPKH := hashing.New(publickey.Encode(&recipient.PublicKey))
	err := accountstable.InsertAccountIntoAccountBalanceTable(dbc, senderPKH, 1000)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}
	err = accountstable.InsertAccountIntoAccountBalanceTable(dbc, recipientPKH, 1000)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}
	zeroValueContract, _ := contracts.New(1, sender, recipientPKH, 0, 1)
	zeroValueContract.Sign(sender)

	nilSenderContract, _ := contracts.New(1, nil, senderPKH, 500, 1)

	senderRecipContract, _ := contracts.New(1, sender, senderPKH, 500, 1)
	senderRecipContract.Sign(sender)

	invalidSignatureContract, _ := contracts.New(1, sender, recipientPKH, 500, 1)
	invalidSignatureContract.Sign(recipient)

	insufficentFundsContract, _ := contracts.New(1, sender, recipientPKH, 2000000, 1)
	insufficentFundsContract.Sign(sender)

	invalidNonceContract, _ := contracts.New(1, sender, recipientPKH, 20, 0)
	invalidNonceContract.Sign(sender)

	invalidNonceContract2, _ := contracts.New(1, sender, recipientPKH, 20, 2)
	invalidNonceContract2.Sign(sender)

	validTwoExistingAccountsContract, _ := contracts.New(1, sender, recipientPKH, 500, 1)
	validTwoExistingAccountsContract.Sign(sender)

	keyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	keyNotInTablePKH := hashing.New(publickey.Encode(&keyNotInTable.PublicKey))

	validOneExistingAccountsContract, _ := contracts.New(1, sender, keyNotInTablePKH, 500, 1)
	validOneExistingAccountsContract.Sign(sender)
	accountstable.InsertAccountIntoAccountBalanceTable(dbc, keyNotInTablePKH, 500)

	anotherKeyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	anotherKeyNotInTablePKH := hashing.New(publickey.Encode(&anotherKeyNotInTable.PublicKey))

	newAccountToANewerAccountContract, _ := contracts.New(1, keyNotInTable, anotherKeyNotInTablePKH, 500, 1)
	newAccountToANewerAccountContract.Sign(keyNotInTable)

	tests := []struct {
		name    string
		c       *contracts.Contract
		wantErr bool
	}{
		{
			name:    "Zero value",
			c:       zeroValueContract,
			wantErr: true,
		},
		{
			name:    "Nil sender",
			c:       nilSenderContract,
			wantErr: true,
		},
		{
			name:    "Sender == Recipient",
			c:       senderRecipContract,
			wantErr: true,
		},
		{
			name:    "Invalid signature",
			c:       invalidSignatureContract,
			wantErr: true,
		},
		{
			name:    "Insufficient funds",
			c:       insufficentFundsContract,
			wantErr: true,
		},
		{
			name:    "Invalid nonce",
			c:       invalidNonceContract,
			wantErr: true,
		},
		{
			name:    "Invalid nonce 2",
			c:       invalidNonceContract2,
			wantErr: true,
		},
		{
			name:    "Totally valid with old accounts",
			c:       validTwoExistingAccountsContract,
			wantErr: false,
		},
		{
			name:    "Totally valid with a new account",
			c:       validOneExistingAccountsContract,
			wantErr: false,
		},
		{
			name:    "Totally valid with a new account to newer account",
			c:       newAccountToANewerAccountContract,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContract(tt.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// func checkBalanceAndNonce(queryBalance uint64, wantBalance uint64, queryNonce uint64, wantNonce uint64) error {
// 	if queryBalance != wantBalance {
// 		return fmt.Errorf("balance does not match; wanted: %d, got: %d", wantBalance, queryBalance)
// 	}
// 	if queryNonce != wantNonce {
// 		return fmt.Errorf("nonce does not match; wanted: %d, got: %d", wantNonce, queryNonce)
// 	}
// 	return nil
// }
