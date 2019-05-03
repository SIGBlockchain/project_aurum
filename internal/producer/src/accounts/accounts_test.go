package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/asn1"
	"encoding/hex"
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
		version   uint16
		sender    *ecdsa.PrivateKey
		recipient []byte
		value     uint64
		nonce     uint64
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
				version:   1,
				sender:    nil,
				recipient: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				value:     1000000000,
				nonce:     0,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    nil,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				Value:           1000000000,
				Nonce:           0,
			},
			wantErr: false,
		},
		{
			name: "Unsigned Normal contract",
			args: args{
				version:   1,
				sender:    senderPrivateKey,
				recipient: block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
				value:     1000000000,
				nonce:     0,
			},
			want: &Contract{
				Version:         1,
				SenderPubKey:    &senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
				Value:           1000000000,
				Nonce:           0,
			},
			wantErr: false,
		},
		{
			name: "Version 0 contract",
			args: args{
				version:   0,
				sender:    senderPrivateKey,
				recipient: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				value:     1000000000,
				nonce:     0,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeContract(tt.args.version, tt.args.sender, tt.args.recipient, tt.args.value, tt.args.nonce)
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
	type args struct {
		withSignature bool
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			name: "Minting contract",
			c:    nullSenderContract,
			args: args{
				withSignature: false,
			},
		},
		{
			name: "Unsigned contract",
			c:    unsignedContract,
			args: args{
				withSignature: false,
			},
		},
		{
			name: "Signed contract",
			c:    signedContract,
			args: args{
				withSignature: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Serialize(tt.args.withSignature)
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
			default:
			}
		})
	}
}

func TestContract_Deserialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	nullSenderContract, _ := MakeContract(1, nil, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)), 1000, 0)
	nullSenderContractSerialized := nullSenderContract.Serialize(false)
	unsignedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 0)
	unsignedContractSerialized := unsignedContract.Serialize(false)
	signedContract, _ := MakeContract(1, senderPrivateKey, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)), 1000, 0)
	signedContract.SignContract(senderPrivateKey)
	signedContractSerialized := signedContract.Serialize(true)
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
				if tt.c.Nonce != nullSenderContract.Nonce {
					t.Errorf("Invalid field on nullSender contract: nonce")
				}
				if tt.c.Signature != nil {
					t.Errorf("Invalid field on nullSender contract: signature")
				}
				if tt.c.SenderPubKey != nil {
					t.Errorf("Invalid field on nullSender contract: sender public key")
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
				if tt.c.Nonce != unsignedContract.Nonce {
					t.Errorf("Invalid field on unsigned contract: nonce")
				}
				if tt.c.Signature != nil {
					t.Errorf("Invalid field on unsigned contract: signature")
				}
				if !reflect.DeepEqual(tt.c.SenderPubKey, &senderPrivateKey.PublicKey) {
					t.Errorf("Invalid field on unsigned contract: sender public key")
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
				if tt.c.Nonce != signedContract.Nonce {
					t.Errorf("Invalid field on signed contract: nonce")
				}
				if !bytes.Equal(tt.c.Signature, signedContract.Signature) {
					t.Errorf("Invalid field on signed contract: signature")
				}
				if !reflect.DeepEqual(tt.c.SenderPubKey, &senderPrivateKey.PublicKey) {
					t.Errorf("Invalid field on signed contract: sender public key")
				}
				break
			default:
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
			tt.c.SignContract(&tt.args.sender)
			serializedTestContract := block.HashSHA256(copyOfContract.Serialize(false))
			var esig struct {
				R, S *big.Int
			}
			if _, err := asn1.Unmarshal(tt.c.Signature, &esig); err != nil {
				t.Errorf("Failed to unmarshall signature")
			}
			if !ecdsa.Verify(tt.c.SenderPubKey, serializedTestContract, esig.R, esig.S) {
				t.Errorf("Failed to verify valid signature")
			}
			maliciousPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if ecdsa.Verify(&maliciousPrivateKey.PublicKey, serializedTestContract, esig.R, esig.S) {
				t.Errorf("Failed to reject invalid signature")
			}
		})
	}
}

func TestInsertAccountIntoAccountBalanceTable(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	dbName := "accountBalanceTable.tab"
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
	dbName := "accountBalanceTable.tab"
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
	dbName := "accountBalanceTable.tab"
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
//// Invalid nonce contracts
//// Completely valid contract

func TestValidateContract(t *testing.T) {
	dbName := "accountBalanceTable.tab"
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

	minter, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	minterPKH := block.HashSHA256(keys.EncodePublicKey(&minter.PublicKey))
	authMinters := [][]byte{minterPKH}

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

	falseMintingContract, _ := MakeContract(1, nil, senderPKH, 500, 1)

	type args struct {
		c                 *Contract
		table             string
		authorizedMinters [][]byte
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Zero value",
			args: args{
				c:                 zeroValueContract,
				table:             dbName,
				authorizedMinters: authMinters,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Unauthorized minting",
			args: args{
				c:                 falseMintingContract,
				table:             dbName,
				authorizedMinters: authMinters,
			},
			want:    false,
			wantErr: false,
		},
		// {
		// 	name: "Authorized minting",
		// 	args: args{
		// 		c:                 mintingContract,
		// 		table:             dbName,
		// 		authorizedMinters: authMinters,
		// 	},
		// 	want:    true,
		// 	wantErr: false,
		// },
		// {
		// 	name: "Invalid signature",
		// 	args: args{
		// 		c:                 invalidSigContract,
		// 		table:             dbName,
		// 		authorizedMinters: authMinters,
		// 	},
		// 	want:    false,
		// 	wantErr: false,
		// },
		// {
		// 	name:    "Insufficient funds",
		// 	want:    false,
		// 	wantErr: false,
		// },
		// {
		// 	name:    "Invalid nonce",
		// 	want:    false,
		// 	wantErr: false,
		// },
		// {
		// 	name:    "Totally valid with old accounts",
		// 	want:    true,
		// 	wantErr: false,
		// },
		// {
		// 	name:    "Totally valid with a new account",
		// 	want:    true,
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateContract(tt.args.c, tt.args.table, tt.args.authorizedMinters)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateContract() = %v, want %v", got, tt.want)
			}
			var pkhash string
			var balance uint64
			var nonce uint64
			rows, err := dbc.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
			if err != nil {
				t.Errorf("Failed to acquire rows from table")
			}
			switch tt.name {
			case "Zero value":
				for rows.Next() {
					err = rows.Scan(&pkhash, &balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan rows: %s", err)
					}
					decodedPkhash, _ := hex.DecodeString(pkhash)
					if !bytes.Equal(decodedPkhash, senderPKH) {
						if balance != 1000 {
							t.Errorf("Invalid balance on zero value sender")
						}
						if nonce != 0 {
							t.Error("Invalid nonce on zero value sender")
						}
					}
					if !bytes.Equal(decodedPkhash, recipientPKH) {
						if balance != 1000 {
							t.Errorf("Invalid balance on zero value recipient")
						}
						if nonce != 0 {
							t.Error("Invalid nonce on zero value recipient")
						}
					}

				}
			case "Unauthorized minting":
				for rows.Next() {
					err = rows.Scan(&pkhash, &balance, &nonce)
					if err != nil {
						t.Errorf("failed to scan rows: %s", err)
					}
					decodedPkhash, _ := hex.DecodeString(pkhash)
					if !bytes.Equal(decodedPkhash, senderPKH) {
						if balance != 1000 {
							t.Errorf("Invalid balance on false minting")
						}
						if nonce != 0 {
							t.Error("Invalid nonce on false minting")
						}
					}
				}
			default:
			}
		})
	}
}

// func TestValidateContract(t *testing.T) {
// 	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	dbName := "test.db"
// 	database, _ := sql.Open("sqlite3", dbName)
// 	defer func() {
// 		database.Close()
// 		err := os.Remove(dbName)
// 		if err != nil {
// 			t.Errorf("Failed to remove database: %s", err)
// 		}
// 	}()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Errorf("Recovered from panic: %s", r)
// 		}
// 	}()
// 	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// 	statement.Exec()
// 	statement, err = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
// 	if err != nil {
// 		t.Errorf("Insertion statement failed: %s", err)
// 	}
// 	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))), 350, 3) // give sender initially 350
// 	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
// 	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))), 200, 2) // give recip initially 200
// 	validContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 4)
// 	copyOfContract := validContract
// 	validContract.SignContract(senderPrivateKey)
// 	invalidContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 5)
// 	invalidContract.SignContract(senderPrivateKey)
// 	invalidNonceContract, _ := MakeContract(1, *recipientPrivateKey, senderPrivateKey.PublicKey, 250, 3)
// 	invalidNonceContract.SignContract(recipientPrivateKey)
// 	zeroValueContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 0, 5)
// 	zeroValueContract.SignContract(senderPrivateKey)
// 	type args struct {
// 		c         Contract
// 		tableName string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "Valid contract",
// 			args: args{
// 				c:         validContract,
// 				tableName: "account_balances",
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 		{
// 			name: "Invalid contract by value",
// 			args: args{
// 				c:         invalidContract,
// 				tableName: "account_balances",
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "Invalid contract by nonce",
// 			args: args{
// 				c:         invalidNonceContract,
// 				tableName: "account_balances",
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "Zero value contract (spam control)",
// 			args: args{
// 				c:         zeroValueContract,
// 				tableName: "account_balances",
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		// {
// 		// 	name: "Same sender and recipient",
// 		// 	args: args {

// 		// 	},
// 		// 	want: false,
// 		// wantErr: true,
// 		// },
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := ValidateContract(tt.args.c, database)
// 			if got != tt.want {
// 				t.Errorf("ValidateContract() = %v, want %v. Error: %v", got, tt.want, err)
// 			}
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ValidateContract() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			serializedCopy := block.HashSHA256(copyOfContract.Serialize(false))
// 			signaturelessValidContract := block.HashSHA256(validContract.Serialize(false))
// 			if !reflect.DeepEqual(serializedCopy, signaturelessValidContract) {
// 				t.Errorf("Contracts do not match. Wanted: %v, got %v", serializedCopy, signaturelessValidContract)
// 			}
// 			var pkhash string
// 			var balance uint64
// 			var nonce uint64
// 			rows, _ := database.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
// 			for rows.Next() {
// 				rows.Scan(&pkhash, &balance, &nonce)
// 				decodedPkhash, _ := hex.DecodeString(pkhash)
// 				if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))) {
// 					if balance != 0 {
// 						t.Errorf("Invalid sender balance: %d", balance)
// 					}
// 					if nonce != 4 {
// 						t.Errorf("Invalid sender nonce: %d", nonce)
// 					}
// 				} else if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))) {
// 					if balance != 550 {
// 						t.Errorf("Invalid recipient balance: %d", balance)
// 					}
// 					if nonce != 3 {
// 						t.Errorf("Invalid recipient nonce: %d", nonce)
// 					}
// 				}
// 			}
// 		})
// 	}
// }
