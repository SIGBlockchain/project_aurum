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

	"github.com/SIGBlockchain/project_aurum/producer/src/block"
	"github.com/SIGBlockchain/project_aurum/producer/src/keys"
	_ "github.com/mattn/go-sqlite3"
)

func TestMakeContract(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	type args struct {
		version   uint16
		sender    ecdsa.PrivateKey
		recipient ecdsa.PublicKey
		value     uint64
		nonce     uint64
	}
	tests := []struct {
		name    string
		args    args
		want    Contract
		wantErr bool
	}{
		{
			name: "Minting contract",
			args: args{
				version:   1,
				sender:    *senderPrivateKey,
				recipient: senderPrivateKey.PublicKey,
				value:     1000000000,
				nonce:     0,
			},
			want: Contract{
				Version:         1,
				SenderPubKey:    senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey)),
				Value:           1000000000,
				Nonce:           0,
			},
			wantErr: false,
		},
		{
			name: "Normal contract",
			args: args{
				version:   1,
				sender:    *senderPrivateKey,
				recipient: recipientPrivateKey.PublicKey,
				value:     1000000000,
				nonce:     0,
			},
			want: Contract{
				Version:         1,
				SenderPubKey:    senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
				Value:           1000000000,
				Nonce:           0,
			},
			wantErr: false,
		},
		{
			name: "Normal contract",
			args: args{
				version:   0,
				sender:    *senderPrivateKey,
				recipient: recipientPrivateKey.PublicKey,
				value:     1000000000,
				nonce:     0,
			},
			want: Contract{
				Version:         0,
				SenderPubKey:    senderPrivateKey.PublicKey,
				SigLen:          0,
				Signature:       nil,
				RecipPubKeyHash: block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey)),
				Value:           1000000000,
				Nonce:           0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeContract(tt.args.version, tt.args.sender, tt.args.recipient, tt.args.value, tt.args.nonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeContract() = %v, want %v", got, tt.want)
			}
			// if tt.wantErr == true && err != errors.New("Invalid version; must be >= 1") {
			// 	t.Errorf("Invalid error return. Should be %s, instead got: %s", "\"Invalid version; must be >= 1\"", err)
			// }
		})
	}
}

func TestContract_Serialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 1000, 5)
	copyOfContract := testContract
	testContract.SignContract(senderPrivateKey)
	type args struct {
		withSignature bool
	}
	tests := []struct {
		name string
		c    Contract
		args args
		want []byte
	}{
		{
			c: testContract,
			args: args{
				withSignature: false,
			},
			want: copyOfContract.Serialize(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Serialize(tt.args.withSignature); !bytes.Equal(got, tt.want) {
				t.Errorf("Contract.Serialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContract_Deserialize(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testContract, _ := MakeContract(1, *senderPrivateKey, senderPrivateKey.PublicKey, 25, 0)
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			c: &Contract{},
			args: args{
				testContract.Serialize(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Deserialize(tt.args.b)
			if !reflect.DeepEqual(tt.c.Version, testContract.Version) {
				t.Errorf("Contract versions do not match; c = %v, testContract = %v", tt.c.Version, testContract.Version)
			}
			if !reflect.DeepEqual(tt.c.SenderPubKey, testContract.SenderPubKey) {
				t.Errorf("Contract sender public keys do not match; c = %v, testContract = %v", tt.c.SenderPubKey, testContract.SenderPubKey)
			}
			if !reflect.DeepEqual(tt.c.SigLen, testContract.SigLen) {
				t.Errorf("Contract signature lengths do not match; c = %v, testContract = %v", tt.c.SigLen, testContract.SigLen)
			}
			if tt.c.Signature != nil {
				t.Errorf("Contract signatures do not match; c = %v, testContract = %v", tt.c.Signature, testContract.Signature)
			}
			if !reflect.DeepEqual(tt.c.RecipPubKeyHash, testContract.RecipPubKeyHash) {
				t.Errorf("Contract recipient public key hashes do not match; c = %v, testContract = %v", tt.c.RecipPubKeyHash, testContract.RecipPubKeyHash)
			}
			if !reflect.DeepEqual(tt.c.Value, testContract.Value) {
				t.Errorf("Contract values do not match; c = %v, testContract = %v", tt.c.Value, testContract.Value)
			}
			if !reflect.DeepEqual(tt.c.Nonce, testContract.Nonce) {
				t.Errorf("Contract nonces do not match; c = %v, testContract = %v", tt.c.Nonce, testContract.Nonce)
			}
		})
	}
}

func TestContract_SignContract(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	testContract, _ := MakeContract(1, *senderPrivateKey, senderPrivateKey.PublicKey, 25, 0)
	type args struct {
		sender ecdsa.PrivateKey
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			c: &testContract,
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
			if !ecdsa.Verify(&tt.c.SenderPubKey, serializedTestContract, esig.R, esig.S) {
				t.Errorf("Failed to verify valid signature")
			}
			maliciousPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if ecdsa.Verify(&maliciousPrivateKey.PublicKey, serializedTestContract, esig.R, esig.S) {
				t.Errorf("Failed to reject invalid signature")
			}
		})
	}
}

func TestContract_UpdateAccountBalanceTable(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	dbName := "test.db"
	database, _ := sql.Open("sqlite3", dbName)
	defer func() {
		database.Close()
		err := os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()
	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))), 350, 3)
	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))), 200, 2)
	testContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 4)
	type args struct {
		table string
	}
	tests := []struct {
		name string
		c    *Contract
		args args
	}{
		{
			c: &testContract,
			args: args{
				table: dbName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// defer func() {
			// 	if r := recover(); r != nil {
			// 		t.Errorf("Recovered from panic: %s", r)
			// 	}
			// }()
			tt.c.UpdateAccountBalanceTable(tt.args.table)
			var pkhash string
			var balance uint64
			var nonce uint64
			rows, _ := database.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
			for rows.Next() {
				rows.Scan(&pkhash, &balance, &nonce)
				decodedPkhash, _ := hex.DecodeString(pkhash)
				if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))) {
					if balance != 0 {
						t.Errorf("Invalid sender balance: %d", balance)
					}
					if nonce != 4 {
						t.Errorf("Invalid sender nonce: %d", nonce)
					}
				} else if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))) {
					if balance != 550 {
						t.Errorf("Invalid recipient balance: %d", balance)
					}
					if nonce != 3 {
						t.Errorf("Invalid recipient nonce: %d", nonce)
					}
				}
			}
		})
	}
}

func TestValidateContract(t *testing.T) {
	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	dbName := "test.db"
	database, _ := sql.Open("sqlite3", dbName)
	defer func() {
		database.Close()
		err := os.Remove(dbName)
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Recovered from panic: %s", r)
		}
	}()
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
	statement.Exec()
	statement, err = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	if err != nil {
		t.Errorf("Insertion statement failed: %s", err)
	}
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))), 350, 3) // give sender initially 350
	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))), 200, 2) // give recip initially 200
	validContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 4)
	copyOfContract := validContract
	validContract.SignContract(senderPrivateKey)
	// invalidContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 5)
	// invalidContract.SignContract(senderPrivateKey)
	// invalidNonceContract, _ := MakeContract(1, *recipientPrivateKey, senderPrivateKey.PublicKey, 250, 3)
	// invalidNonceContract.SignContract(recipientPrivateKey)
	// zeroValueContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 0, 5)
	// zeroValueContract.SignContract(senderPrivateKey)
	type args struct {
		c         Contract
		tableName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid contract",
			args: args{
				c: validContract,
			},
			want: true,
		},
		// {
		// 	name: "Invalid contract by value",
		// 	args: args{
		// 		c: invalidContract,
		// 	},
		// 	want: false,
		// },
		// {
		// 	name: "Invalid contract by nonce",
		// 	args: args{
		// 		c: invalidNonceContract,
		// 	},
		// 	want: false,
		// },
		// {
		// 	name: "Zero value contract (spam control)",
		// 	args: args{
		// 		c: zeroValueContract,
		// 	},
		// 	want: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got, err := ValidateContract(tt.args.c, tt.args.tableName); got != tt.want {
			// 	t.Errorf("ValidateContract() = %v, want %v. Error: %s", got, tt.want, err)
			// }
			serializedCopy := block.HashSHA256(copyOfContract.Serialize(false))
			signaturelessValidContract := block.HashSHA256(validContract.Serialize(false))
			if !reflect.DeepEqual(serializedCopy, signaturelessValidContract) {
				t.Errorf("Contracts do not match. Wanted: %v, got %v", serializedCopy, signaturelessValidContract)
			}

			// var pkhash string
			// var balance uint64
			// var nonce uint64
			// rows, _ := database.Query("SELECT public_key_hash, balance, nonce FROM account_balances")
			// for rows.Next() {
			// 	rows.Scan(&pkhash, &balance, &nonce)
			// 	decodedPkhash, _ := hex.DecodeString(pkhash)
			// 	if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))) {
			// 		if balance != 0 {
			// 			t.Errorf("Invalid sender balance: %d", balance)
			// 		}
			// 		if nonce != 4 {
			// 			t.Errorf("Invalid sender nonce: %d", nonce)
			// 		}
			// 	} else if bytes.Equal(decodedPkhash, block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))) {
			// 		if balance != 550 {
			// 			t.Errorf("Invalid recipient balance: %d", balance)
			// 		}
			// 		if nonce != 3 {
			// 			t.Errorf("Invalid recipient nonce: %d", nonce)
			// 		}
			// 	}
			// }
		})
	}
}
