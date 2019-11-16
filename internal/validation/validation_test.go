package validation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/accountstable"
	"github.com/SIGBlockchain/project_aurum/internal/block"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/contracts"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
	"github.com/SIGBlockchain/project_aurum/internal/wallet"
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
	statement, _ := dbc.Prepare(sqlstatements.CREATE_ACCOUNT_BALANCES_TABLE)
	statement.Exec()

	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, _ := publickey.Encode(&sender.PublicKey)
	senderPKH := hashing.New(encodedSenderPublicKey)
	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientPublicKey, _ := publickey.Encode(&recipient.PublicKey)
	recipientPKH := hashing.New(encodedRecipientPublicKey)
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
	encodedKeyNotInTablePublicKey, _ := publickey.Encode(&keyNotInTable.PublicKey)
	keyNotInTablePKH := hashing.New(encodedKeyNotInTablePublicKey)

	validOneExistingAccountsContract, _ := contracts.New(1, sender, keyNotInTablePKH, 500, 1)
	validOneExistingAccountsContract.Sign(sender)
	accountstable.InsertAccountIntoAccountBalanceTable(dbc, keyNotInTablePKH, 500)

	anotherKeyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedAnotherKeyNotInTablePublicKey, _ := publickey.Encode(&anotherKeyNotInTable.PublicKey)
	anotherKeyNotInTablePKH := hashing.New(encodedAnotherKeyNotInTablePublicKey)

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
			err := ValidateContract(dbc, tt.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test cases for ValidatePending
//// Zero value
//// Nil Sender
//// Sender == Recipient
//// Invalid signature
//// Insufficient balance
//// Invalid state nonce
//// Completely valid

func TestValidatePending(t *testing.T) {
	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, _ := publickey.Encode(&sender.PublicKey)
	senderPKH := hashing.New(encodedSenderPublicKey)
	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedRecipientPublicKey, _ := publickey.Encode(&recipient.PublicKey)
	recipientPKH := hashing.New(encodedRecipientPublicKey)

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

	// Start: pBalance = 100, pNonce = 0
	validFirstContract, _ := contracts.New(1, sender, recipientPKH, 50, 1)
	validFirstContract.Sign(sender)

	// pBalance = 50, pNonce = 1
	keyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	encodedSenderPublicKey, err := publickey.Encode(&keyNotInTable.PublicKey)
	if err != nil {
		t.Errorf("failure to encode Sender Public Key: %v", err)
	}
	keyNotInTablePKH := hashing.New(encodedSenderPublicKey)

	InvalidBalanceContract, _ := contracts.New(1, sender, keyNotInTablePKH, 51, 2)
	InvalidBalanceContract.Sign(sender)

	InvalidNonceContract, _ := contracts.New(1, sender, keyNotInTablePKH, 20, 3)
	InvalidNonceContract.Sign(sender)

	ValidSecondContract, _ := contracts.New(1, sender, keyNotInTablePKH, 50, 2)
	ValidSecondContract.Sign(sender)

	tests := []struct {
		name     string
		c        *contracts.Contract
		pBalance uint64
		pNonce   uint64
		wantErr  bool
	}{
		{
			name:     "Zero value",
			c:        zeroValueContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Nil sender",
			c:        nilSenderContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Sender == Recipient",
			c:        senderRecipContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Invalid signature",
			c:        invalidSignatureContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Insufficient funds",
			c:        insufficentFundsContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Invalid nonce",
			c:        invalidNonceContract,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Invalid nonce 2",
			c:        invalidNonceContract2,
			pBalance: 1000,
			pNonce:   0,
			wantErr:  true,
		},
		{
			name:     "Totally valid",
			c:        validFirstContract,
			pBalance: 100,
			pNonce:   0,
			wantErr:  false,
		},
		{
			name:    "Invalid balance",
			c:       InvalidBalanceContract,
			wantErr: true,
		},
		{
			name:    "Invalid state nonce",
			c:       InvalidNonceContract,
			wantErr: true,
		},
		{
			name:    "Totally valid 2",
			c:       ValidSecondContract,
			wantErr: false,
		},
	}

	var updatedBal uint64
	var updatedNonce uint64
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i > 7 {
				tt.pBalance = updatedBal
				tt.pNonce = updatedNonce
			}

			var err error
			err = ValidatePending(tt.c, &tt.pBalance, &tt.pNonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePending() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			updatedBal = tt.pBalance
			updatedNonce = tt.pNonce
		})
	}
}

func TestValidateBlock(t *testing.T) {
	baseBlk := block.Block{
		Version:        1,
		Height:         0,
		PreviousHash:   hashing.New([]byte{'x'}),
		MerkleRootHash: hashing.New([]byte{'q'}),
		Timestamp:      time.Now().UnixNano() - 10,
		Data:           [][]byte{hashing.New([]byte{'r'})},
	}
	baseBlk.DataLen = uint16(len(baseBlk.Data))
	tests := []struct {
		name  string
		block block.Block
		want  bool
	}{
		{
			"Valid Block",
			block.Block{
				Version:        1,
				Height:         baseBlk.Height + 1,
				PreviousHash:   block.HashBlock(baseBlk),
				MerkleRootHash: hashing.GetMerkleRootHash(baseBlk.Data),
				Timestamp:      time.Now().UnixNano(),
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			true,
		},
		{
			"Invalid version",
			block.Block{
				Version:        10000,
				Height:         baseBlk.Height,
				PreviousHash:   baseBlk.PreviousHash,
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      baseBlk.Timestamp,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid height",
			block.Block{
				Version:        baseBlk.Version,
				Height:         999999999,
				PreviousHash:   baseBlk.PreviousHash,
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      baseBlk.Timestamp,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid previous hash",
			block.Block{
				Version:        baseBlk.Version,
				Height:         baseBlk.Height,
				PreviousHash:   hashing.New([]byte{'a'}),
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      baseBlk.Timestamp,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid previous hash (nil)",
			block.Block{
				Version:        baseBlk.Version,
				Height:         baseBlk.Height,
				PreviousHash:   nil,
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      baseBlk.Timestamp,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid timestamp",
			block.Block{
				Version:        baseBlk.Version,
				Height:         baseBlk.Height,
				PreviousHash:   baseBlk.PreviousHash,
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      baseBlk.Timestamp,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid timestamp (future)",
			block.Block{
				Version:        baseBlk.Version,
				Height:         baseBlk.Height,
				PreviousHash:   baseBlk.PreviousHash,
				MerkleRootHash: baseBlk.MerkleRootHash,
				Timestamp:      time.Now().UnixNano() + 100,
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
		{
			"Invalid Merkle-root Hash",
			block.Block{
				Version:        1,
				Height:         baseBlk.Height + 1,
				PreviousHash:   block.HashBlock(baseBlk),
				MerkleRootHash: hashing.New([]byte{'y'}),
				Timestamp:      time.Now().UnixNano(),
				Data:           baseBlk.Data,
				DataLen:        baseBlk.DataLen,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := ValidateBlock(tt.block, baseBlk.Version, baseBlk.Height, block.HashBlock(baseBlk),
				baseBlk.Timestamp); result != tt.want {
				t.Errorf("Validate returned the wrong boolean. Want: %v Got: %v", tt.want, result)
			}
		})
	}
}

func TestValidateProducerTimestamp(t *testing.T) {
	wallet.CreateProducerTable()
	db, err := sql.Open("sqlite3", constants.ProducerTable)
	if err != nil {
		t.Error("Failed to open database for test")
	}
	defer func() {
		db.Close()
		os.Remove("producer.db")
	}()

	walletAddr := []byte{'a'}
	tableTimestamp := int64(time.Now().Nanosecond())
	hashedWalletAddr := hashing.New(walletAddr)
	_, err = db.Exec(sqlstatements.INSERT_VALUES_INTO_PRODUCER, hashedWalletAddr, int(tableTimestamp))
	if err != nil {
		t.Error("Failed to execute statement for database")
	}

	tests := []struct {
		name       string
		timeStamp  int64
		walletAddr []byte
		interval   time.Duration
		want       bool
	}{
		{
			"Valid producer timestamp",
			tableTimestamp + time.Second.Nanoseconds() + 1,
			hashedWalletAddr,
			time.Second,
			true,
		},
		{
			"Valid producer timestamp (Equal)",
			tableTimestamp + time.Second.Nanoseconds(),
			hashedWalletAddr,
			time.Second,
			true,
		},
		{
			"Invalid timestamp",
			tableTimestamp,
			hashedWalletAddr,
			time.Second,
			false,
		},
		{
			"Invalid wallet address",
			tableTimestamp + time.Second.Nanoseconds(),
			hashing.New([]byte{'b'}),
			time.Second,
			false,
		},
		{
			"Nil wallet address",
			tableTimestamp + time.Second.Nanoseconds(),
			nil,
			time.Second,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateProducerTimestamp(db, tt.timeStamp, tt.walletAddr, tt.interval)
			if err != nil {
				t.Errorf("ValidateProducerTimestamp returned err: %v", err)
			}
			if result != tt.want {
				t.Errorf("ValidateProducerTimestamp returned the wrong boolean. Want: %v Got: %v", tt.want, result)
			}
		})
	}

}

func TestSetConfigFromFlags(t *testing.T) {

}
