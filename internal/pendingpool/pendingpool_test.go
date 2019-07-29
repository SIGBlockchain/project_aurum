package pendingpool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
	_ "github.com/mattn/go-sqlite3"
)

func TestAdd(t *testing.T) {
	dbc, _ := sql.Open("sqlite3", constants.AccountsTable)
	defer func() {
		err := dbc.Close()
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
		err = os.Remove(constants.AccountsTable)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	statement, _ := dbc.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	statement.Exec()

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	senderPKH := hashing.New(publickey.Encode(&sender.PublicKey))
	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPKH := hashing.New(publickey.Encode(&recipient.PublicKey))

	err := accountstable.InsertAccountIntoAccountBalanceTable(dbc, senderPKH, 100)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}
	err = accountstable.InsertAccountIntoAccountBalanceTable(dbc, recipientPKH, 99)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}

	validFirstContract, _ := contracts.New(1, sender, recipientPKH, 50, 1)
	validFirstContract.Sign(sender)

	invalidBalanceContract, _ := contracts.New(1, sender, recipientPKH, 51, 2)
	invalidBalanceContract.Sign(sender)

	invalidNonceContract, _ := contracts.New(1, sender, recipientPKH, 20, 3)
	invalidNonceContract.Sign(sender)

	validSecondContract, _ := contracts.New(1, sender, recipientPKH, 50, 2)
	validSecondContract.Sign(sender)

	recipientToSender1, _ := contracts.New(1, recipient, senderPKH, 33, 1)
	recipientToSender1.Sign(recipient)

	recipientToSender2, _ := contracts.New(1, recipient, senderPKH, 33, 2)
	recipientToSender2.Sign(recipient)

	recipientToSender3, _ := contracts.New(1, recipient, senderPKH, 33, 3)
	recipientToSender3.Sign(recipient)

	tests := []struct {
		name    string
		c       *contracts.Contract
		wantErr bool
	}{
		{
			"valid first contract",
			validFirstContract,
			false,
		},
		{
			"invalid balance contract",
			invalidBalanceContract,
			true,
		},
		{
			"invalid state nonce contract",
			invalidNonceContract,
			true,
		},
		{
			"valid second contract",
			validSecondContract,
			false,
		},
		{
			"recipient to sender contract 1",
			recipientToSender1,
			false,
		},
		{
			"recipient to sender contract 2",
			recipientToSender2,
			false,
		},
		{
			"recipient to sender contract 3",
			recipientToSender3,
			false,
		},
	}

	m := NewPendingMap()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.name)
			if err := m.Add(tt.c, dbc); (err != nil) != tt.wantErr {
				t.Errorf("Add() returned error: " + err.Error())
			}
		})
	}
}
