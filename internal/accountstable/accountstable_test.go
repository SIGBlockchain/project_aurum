package accountstable

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/accountinfo"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/hashing"
	"github.com/SIGBlockchain/project_aurum/internal/publickey"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
	_ "github.com/mattn/go-sqlite3"
)

func TestInsertAccountIntoAccountBalanceTable(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
				hashing.New(publickey.Encode(&somePrivateKey.PublicKey)),
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
			rows, err := dbc.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
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
				if bytes.Equal(decodedPkhash, hashing.New(publickey.Encode(&somePrivateKey.PublicKey))) {
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
	spkh := hashing.New(publickey.Encode(&senderPrivateKey.PublicKey))
	rpkh := hashing.New(publickey.Encode(&recipientPrivateKey.PublicKey))
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
				rows, err := dbc.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
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
	spkh := hashing.New(publickey.Encode(&somePrivateKey.PublicKey))
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
				rows, err := dbc.Query(sqlstatements.GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES)
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

func TestGetBalance(t *testing.T) {
	somePrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spkh := hashing.New(publickey.Encode(&somePrivateKey.PublicKey))
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
			got, err := GetBalance(dbc, tt.args.pkhash)
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
	spkh := hashing.New(publickey.Encode(&somePrivateKey.PublicKey))
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
			got, err := GetStateNonce(dbc, tt.args.pkhash)
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
	spkh := hashing.New(publickey.Encode(&somePrivateKey.PublicKey))
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
		want    *accountinfo.AccountInfo
		wantErr bool
	}{
		{
			args:    args{spkh},
			want:    &accountinfo.AccountInfo{1000, 0},
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
			got, err := GetAccountInfo(dbc, tt.args.pkhash)
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
