package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/pkg/keys"
	_ "github.com/mattn/go-sqlite3"
)

func TestMakeContract(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
				recipient:     block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				value:         1000000000,
				newStateNonce: 1,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    nil,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
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
				recipient:     block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
				value:         1000000000,
				newStateNonce: 1,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    &senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
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
				recipient:     block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				value:         1000000000,
				newStateNonce: 1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeContract(tt.args.version, tt.args.sender, tt.args.recipient, tt.args.value, tt.args.newStateNonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeContract() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeContract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContract_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	nullSenderContract, _ := MakeContract(1, nil, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)), 1000, 0)
	unsignedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 0)
	signedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 0)
	signedContract.SignContract(senderPrivateKey)
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
				if !bytes.Equal(got[2:180], keys.EncodePublicKey(tt.c.SenderPubKey)) {
					t.Errorf("Invalid encoded public key for unsigned contract")
				}
				if !bytes.Equal(got[181:213], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in unsigned contract")
				}
			case "Signed Contract":
				if sigLen == 0 {
					t.Errorf("Zero length signature in signed contract: %v", sigLen)
				}
				if !bytes.Equal(got[2:180], keys.EncodePublicKey(tt.c.SenderPubKey)) {
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
	nullSenderContract, _ := MakeContract(1, nil, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)), 1000, 1)
	nullSenderContractSerialized, _ := nullSenderContract.Serialize()
	unsignedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 1)
	unsignedContractSerialized, _ := unsignedContract.Serialize()
	signedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 1)
	signedContract.SignContract(senderPrivateKey)
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

func TestContract_SignContract(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)), 1000, 0)
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
			hashedContract := block.HashSHA256(serializedTestContract)
			tt.c.SignContract(&tt.args.sender)
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

func TestInsertAccountIntoAccountBalanceTable(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	dbName := "accounts.tab"
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
	type args struct {
		dbConnection *sql.DB
		pkhash       []byte
		value        uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				dbc,
				block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey)),
				1000,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InsertAccountIntoAccountBalanceTable(tt.args.dbConnection, tt.args.pkhash, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("InsertAccountIntoAccountBalanceTable() error = %v, wantErr %v", err, tt.wantErr)
			}
			var pkhash string
			var balance uint64
			var nonce uint64
			rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
			if err != nil {
				t.Errorf("Failed to acquire rows from table")
			}
			for rows.Next() {
				err = rows.Scan(&pkhash, &balance, &nonce)
				if err != nil {
					t.Errorf("failed to scan rows: %s", err)
				}
				decodedPkhash, err := hex.DecodeString(pkhash)
				if err != nil {
					t.Errorf("failed to decode public key hash")
				}
				if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey))) {
					if balance != 1000 {
						t.Errorf("Invalid balance: %d", balance)
					}
					if nonce != 0 {
						t.Errorf("Invalid nonce: %d", nonce)
					}
				}
			}
		})
	}
}

func TestExchangeBetweenAccountsUpdateAccountBalanceTable(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))
	rpkh := block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))
	dbName := "accounts.tab"
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
	err := InsertAccountIntoAccountBalanceTable(dbc, spkh, 1000)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	err = InsertAccountIntoAccountBalanceTable(dbc, rpkh, 1000)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	type args struct {
		dbConnection *sql.DB
		senderPKH    []byte
		recipPKH     []byte
		value        uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				dbConnection: dbc,
				senderPKH:    spkh,
				recipPKH:     rpkh,
				value:        250,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExchangeBetweenAccountsUpdateAccountBalanceTable(tt.args.dbConnection, tt.args.senderPKH, tt.args.recipPKH, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ExchangeBetweenAccountsUpdateAccountBalanceTable() error = %v, wantErr %v", err, tt.wantErr)
				var pkhash string
				var balance uint64
				var nonce uint64
				rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
				if err != nil {
					t.Errorf("Failed to acquire rows from table")
				}
				for rows.Next() {
					err = rows.Scan(&pkhash, &balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan rows: %s", err)
					}
					decodedPkhash, err := hex.DecodeString(pkhash)
					if err != nil {
						t.Errorf("failed to decode public key hash")
					}
					if bytes.Equal(decodedPkhash, spkh) {
						if balance != 750 {
							t.Errorf("Invalid sender balance: %d", balance)
						}
						if nonce != 1 {
							t.Errorf("Invalid sender nonce: %d", nonce)
						}
					} else if bytes.Equal(decodedPkhash, rpkh) {
						if balance != 1250 {
							t.Errorf("Invalid recipient balance: %d", balance)
						}
						if nonce != 1 {
							t.Errorf("Invalid recipient nonce: %d", nonce)
						}
					}
				}
			}
		})
	}
}

func TestMintAurumUpdateAccountBalanceTable(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey))
	dbName := "accounts.tab"
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
	err := InsertAccountIntoAccountBalanceTable(dbc, spkh, 1000)
	if err != nil {
		t.Errorf("failed to insert account into balance table")
	}
	type args struct {
		dbConnection *sql.DB
		pkhash       []byte
		value        uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				dbConnection: dbc,
				pkhash:       spkh,
				value:        1500,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MintAurumUpdateAccountBalanceTable(tt.args.dbConnection, tt.args.pkhash, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MintAurumUpdateAccountBalanceTable() error = %v, wantErr %v", err, tt.wantErr)
				var pkhash string
				var balance uint64
				var nonce uint64
				rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
				if err != nil {
					t.Errorf("Failed to acquire rows from table")
				}
				for rows.Next() {
					err = rows.Scan(&pkhash, &balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan rows: %s", err)
					}
					decodedPkhash, err := hex.DecodeString(pkhash)
					if err != nil {
						t.Errorf("failed to decode public key hash")
					}
					if bytes.Equal(decodedPkhash, spkh) {
						if balance != 2500 {
							t.Errorf("Invalid balance: %d", balance)
						}
						if nonce != 1 {
							t.Errorf("Invalid nonce: %d", nonce)
						}
					}
				}
			}
		})
	}
}

// Test cases for validation (next issue)
//// Zero value contracts
//// Minting contracts
//// Invalid signature contracts
//// Insufficient balance contracts
//// Completely valid contract

func TestValidateContract(t *testing.T) {
	dbName := "accounts.tab"
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
	senderPKH := block.HashSHA256(keys.EncodePublicKey(&sender.PublicKey))
	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPKH := block.HashSHA256(keys.EncodePublicKey(&recipient.PublicKey))
	err := InsertAccountIntoAccountBalanceTable(dbc, senderPKH, 1000)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}
	err = InsertAccountIntoAccountBalanceTable(dbc, recipientPKH, 1000)
	if err != nil {
		t.Errorf("Failed to insert zero Sender account")
	}
	zeroValueContract, _ := MakeContract(1, sender, recipientPKH, 0, 1)
	zeroValueContract.SignContract(sender)

	nilSenderContract, _ := MakeContract(1, nil, senderPKH, 500, 1)

	senderRecipContract, _ := MakeContract(1, sender, senderPKH, 500, 1)
	senderRecipContract.SignContract(sender)

	invalidSignatureContract, _ := MakeContract(1, sender, recipientPKH, 500, 1)
	invalidSignatureContract.SignContract(recipient)

	insufficentFundsContract, _ := MakeContract(1, sender, recipientPKH, 2000000, 1)
	insufficentFundsContract.SignContract(sender)

	invalidNonceContract, _ := MakeContract(1, sender, recipientPKH, 20, 0)
	invalidNonceContract.SignContract(sender)

	invalidNonceContract2, _ := MakeContract(1, sender, recipientPKH, 20, 2)
	invalidNonceContract2.SignContract(sender)

	validTwoExistingAccountsContract, _ := MakeContract(1, sender, recipientPKH, 500, 1)
	validTwoExistingAccountsContract.SignContract(sender)

	keyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	keyNotInTablePKH := block.HashSHA256(keys.EncodePublicKey(&keyNotInTable.PublicKey))

	validOneExistingAccountsContract, _ := MakeContract(1, sender, keyNotInTablePKH, 500, 1)
	validOneExistingAccountsContract.SignContract(sender)
	InsertAccountIntoAccountBalanceTable(dbc, keyNotInTablePKH, 500)

	anotherKeyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	anotherKeyNotInTablePKH := block.HashSHA256(keys.EncodePublicKey(&anotherKeyNotInTable.PublicKey))

	newAccountToANewerAccountContract, _ := MakeContract(1, keyNotInTable, anotherKeyNotInTablePKH, 500, 1)
	newAccountToANewerAccountContract.SignContract(keyNotInTable)

	tests := []struct {
		name    string
		c       *Contract
		want    bool
		wantErr bool
	}{
		{
			name:    "Zero value",
			c:       zeroValueContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Nil sender",
			c:       nilSenderContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Sender == Recipient",
			c:       senderRecipContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid signature",
			c:       invalidSignatureContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Insufficient funds",
			c:       insufficentFundsContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid nonce",
			c:       invalidNonceContract,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Invalid nonce 2",
			c:       invalidNonceContract2,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Totally valid with old accounts",
			c:       validTwoExistingAccountsContract,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Totally valid with a new account",
			c:       validOneExistingAccountsContract,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Totally valid with a new account to newer account",
			c:       newAccountToANewerAccountContract,
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateContract(tt.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateContract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func checkBalanceAndNonce(queryBalance uint64, wantBalance uint64, queryNonce uint64, wantNonce uint64) error {
	if queryBalance != wantBalance {
		return fmt.Errorf("balance does not match; wanted: %d, got: %d", wantBalance, queryBalance)
	}
	if queryNonce != wantNonce {
		return fmt.Errorf("nonce does not match; wanted: %d, got: %d", wantNonce, queryNonce)
	}
	return nil
}

func TestAccountInfo_Deserialize(t *testing.T) {
	ac := NewAccountInfo(9001, 50)
	serAc, _ := ac.Serialize()
	type args struct {
		serializedAccountInfo []byte
	}
	tests := []struct {
		name    string
		accInfo *AccountInfo
		args    args
		wantErr bool
	}{
		{
			accInfo: &AccountInfo{},
			args:    args{serializedAccountInfo: serAc},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.accInfo.Deserialize(tt.args.serializedAccountInfo); (err != nil) != tt.wantErr {
				t.Errorf("AccountInfo.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.accInfo, ac) {
				t.Errorf("structs do not match")
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey))
	dbName := "accounts.tab"
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
	err := InsertAccountIntoAccountBalanceTable(dbc, spkh, 1000)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	type args struct {
		pkhash []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			args:    args{spkh},
			want:    1000,
			wantErr: false,
		},
		{
			args:    args{[]byte("doesn't exist in table")},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBalance(tt.args.pkhash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStateNonce(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey))
	dbName := "accounts.tab"
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
	err := InsertAccountIntoAccountBalanceTable(dbc, spkh, 1000)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	type args struct {
		pkhash []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			args:    args{spkh},
			want:    0,
			wantErr: false,
		},
		{
			args:    args{[]byte("doesn't exist in table")},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStateNonce(tt.args.pkhash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStateNonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStateNonce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAccountInfo(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := block.HashSHA256(keys.EncodePublicKey(&somePrivateKey.PublicKey))
	dbName := "accounts.tab"
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
	err := InsertAccountIntoAccountBalanceTable(dbc, spkh, 1000)
	if err != nil {
		t.Errorf("failed to insert sender account")
	}
	type args struct {
		pkhash []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *AccountInfo
		wantErr bool
	}{
		{
			args:    args{spkh},
			want:    &AccountInfo{1000, 0},
			wantErr: false,
		},
		{
			args:    args{[]byte("this account doesn't exit")},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAccountInfo(tt.args.pkhash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	contract1 := Contract{
		Version:         1,
		SenderPubKey:    &senderPrivateKey.PublicKey,
		SigLen:          0,
		Signature:       nil,
		RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
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
	contracts[4].RecipPubKeyHash = block.HashSHA256(keys.EncodePublicKey(&anotherSenderPrivateKey.PublicKey))
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

// func TestValidateContract(t *testing.T) {
// 	dbName := "accounts.tab"
// 	dbc, _ := sql.Open("sqlite3", dbName)
// 	defer func() {
// 		err := dbc.Close()
// 		if err != nil {
// 			t.Errorf("Failed to remove database: %s", err)
// 		}
// 		err = os.Remove(dbName)
// 		if err != nil {
// 			t.Errorf("Failed to remove database: %s", err)
// 		}
// 	}()
// 	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
// 	statement.Exec()

// 	minter, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	minterPKH := block.HashSHA256(keys.EncodePublicKey(&minter.PublicKey))
// 	err := InsertAccountIntoAccountBalanceTable(dbc, minterPKH, 1000)
// 	if err != nil {
// 		t.Errorf("Failed to insert minter account")
// 	}
// 	authMinters := [][]byte{minterPKH}

// 	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	senderPKH := block.HashSHA256(keys.EncodePublicKey(&sender.PublicKey))
// 	recipient, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	recipientPKH := block.HashSHA256(keys.EncodePublicKey(&recipient.PublicKey))
// 	err = InsertAccountIntoAccountBalanceTable(dbc, senderPKH, 1000)
// 	if err != nil {
// 		t.Errorf("Failed to insert zero Sender account")
// 	}
// 	err = InsertAccountIntoAccountBalanceTable(dbc, recipientPKH, 1000)
// 	if err != nil {
// 		t.Errorf("Failed to insert zero Sender account")
// 	}
// 	zeroValueContract, _ := MakeContract(1, sender, recipientPKH, 0, 1)
// 	zeroValueContract.SignContract(sender)

// 	falseMintingContract, _ := MakeContract(1, nil, senderPKH, 500, 1)

// 	validMintingContract, _ := MakeContract(1, nil, minterPKH, 500, 1)

// 	invalidSignatureContract, _ := MakeContract(1, sender, recipientPKH, 500, 1)
// 	invalidSignatureContract.SignContract(recipient)

// 	insufficentFundsContract, _ := MakeContract(1, sender, recipientPKH, 2000000, 1)
// 	insufficentFundsContract.SignContract(sender)

// 	invalidNonceContract, _ := MakeContract(1, sender, recipientPKH, 20, 0)
// 	invalidNonceContract.SignContract(sender)

// 	invalidNonceContract2, _ := MakeContract(1, sender, recipientPKH, 20, 2)
// 	invalidNonceContract2.SignContract(sender)

// 	validTwoExistingAccountsContract, _ := MakeContract(1, sender, recipientPKH, 500, 1)
// 	validTwoExistingAccountsContract.SignContract(sender)

// 	keyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	keyNotInTablePKH := block.HashSHA256(keys.EncodePublicKey(&keyNotInTable.PublicKey))

// 	validOneExistingAccountsContract, _ := MakeContract(1, sender, keyNotInTablePKH, 500, 2)
// 	validOneExistingAccountsContract.SignContract(sender)

// 	anotherKeyNotInTable, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	anotherKeyNotInTablePKH := block.HashSHA256(keys.EncodePublicKey(&anotherKeyNotInTable.PublicKey))

// 	newAccountToANewerAccountContract, _ := MakeContract(1, keyNotInTable, anotherKeyNotInTablePKH, 500, 1)
// 	newAccountToANewerAccountContract.SignContract(keyNotInTable)

// 	type args struct {
// 		c                 *Contract
// 		table             string
// 		authorizedMinters [][]byte
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "Zero value",
// 			args: args{
// 				c:                 zeroValueContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Unauthorized minting",
// 			args: args{
// 				c:                 falseMintingContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Authorized minting",
// 			args: args{
// 				c:                 validMintingContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Invalid signature",
// 			args: args{
// 				c:                 invalidSignatureContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Insufficient funds",
// 			args: args{
// 				c:                 insufficentFundsContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Invalid nonce",
// 			args: args{
// 				c:                 invalidNonceContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Invalid nonce 2",
// 			args: args{
// 				c:                 invalidNonceContract2,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Totally valid with old accounts",
// 			args: args{
// 				c:                 validTwoExistingAccountsContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Totally valid with a new account",
// 			args: args{
// 				c:                 validOneExistingAccountsContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Totally valid with a new account to newer account",
// 			args: args{
// 				c:                 newAccountToANewerAccountContract,
// 				table:             dbName,
// 				authorizedMinters: authMinters,
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := ValidateContract(tt.args.c, tt.args.table, tt.args.authorizedMinters)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("ValidateContract() = %v, want %v", got, tt.want)
// 			}
// 			var pkhash string
// 			var balance uint64
// 			var nonce uint64
// 			rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
// 			if err != nil {
// 				t.Errorf("Failed to acquire rows from table")
// 			}
// 			defer rows.Close()
// 			switch tt.name {
// 			case "Unauthorized minting":
// 				for rows.Next() {
// 					if err := rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, senderPKH) {
// 						if err := checkBalanceAndNonce(balance, 1000, nonce, 0); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			case "Authorized minting":
// 				for rows.Next() {
// 					if err = rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, minterPKH) {
// 						if err := checkBalanceAndNonce(balance, 1500, nonce, 1); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			case "Zero value":
// 			case "Invalid signature":
// 			case "Insufficient funds":
// 			case "Invalid nonce":
// 			case "Invalid nonce 2":
// 				for rows.Next() {
// 					if err := rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, senderPKH) || bytes.Equal(decodedPkhash, recipientPKH) {
// 						if err := checkBalanceAndNonce(balance, 1000, nonce, 0); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			case "Totally valid with old accounts":
// 				for rows.Next() {
// 					if err = rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, senderPKH) {
// 						if err := checkBalanceAndNonce(balance, 500, nonce, 1); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 					if bytes.Equal(decodedPkhash, recipientPKH) {
// 						if err := checkBalanceAndNonce(balance, 1500, nonce, 1); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			case "Totally valid with a new account":
// 				for rows.Next() {
// 					if err = rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, senderPKH) {
// 						if err := checkBalanceAndNonce(balance, 0, nonce, 2); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 					if bytes.Equal(decodedPkhash, keyNotInTablePKH) {
// 						if err := checkBalanceAndNonce(balance, 500, nonce, 0); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			case "Totally valid with a new account to newer account":
// 				for rows.Next() {
// 					if err = rows.Scan(&pkhash, &balance, &nonce); err != nil {
// 						t.Errorf("failed to scan rows: %s", err)
// 					}
// 					decodedPkhash, _ := hex.DecodeString(pkhash)
// 					if bytes.Equal(decodedPkhash, keyNotInTablePKH) {
// 						if err := checkBalanceAndNonce(balance, 0, nonce, 1); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 					if bytes.Equal(decodedPkhash, anotherKeyNotInTablePKH) {
// 						if err := checkBalanceAndNonce(balance, 500, nonce, 0); err != nil {
// 							t.Errorf(err.Error())
// 						}
// 					}
// 				}
// 				break
// 			}
// 		})
// 	}
// }
