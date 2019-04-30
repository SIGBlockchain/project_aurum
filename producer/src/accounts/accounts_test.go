package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"strconv"
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
		// {
		// 	name: "Signed contract", // WILL IMPLEMENT AFTER SIGN CONTRACT FUNCTION
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Serialize(tt.args.withSignature)
			switch tt.name {
			case "Minting contract":
				if !bytes.Equal(got[2:180], bytes.Repeat([]byte(strconv.Itoa(0)), 178)) {
					t.Errorf("Non null sender public key for minting contract")
				}
				if got[180] != 0 {
					t.Errorf("Non-zero signature length in minting contract: %s", got[180])
				}
				if !bytes.Equal(got[181:213], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in minting contract")
				}
				break
			case "Unsigned contract":
				if got[180] != 0 {
					t.Errorf("Non-zero signature length in unsigned contract: %s", got[180])
				}
				if !bytes.Equal(got[2:180], keys.EncodePublicKey(tt.c.SenderPubKey)) {
					t.Errorf("Invalid encoded public key for unsigned contract")
				}
				if !bytes.Equal(got[181:213], tt.c.RecipPubKeyHash) {
					t.Errorf("Invalid recipient public key hash in unsigned contract")
				}
			default:
			}
		})
	}
}

// func TestContract_Deserialize(t *testing.T) {
// 	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	testContract, _ := MakeContract(1, *senderPrivateKey, senderPrivateKey.PublicKey, 25, 0)
// 	type args struct {
// 		b []byte
// 	}
// 	tests := []struct {
// 		name string
// 		c    *Contract
// 		args args
// 	}{
// 		{
// 			c: &Contract{},
// 			args: args{
// 				testContract.Serialize(false),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.c.Deserialize(tt.args.b)
// 			if !reflect.DeepEqual(tt.c.Version, testContract.Version) {
// 				t.Errorf("Contract versions do not match; c = %v, testContract = %v", tt.c.Version, testContract.Version)
// 			}
// 			if !reflect.DeepEqual(tt.c.SenderPubKey, testContract.SenderPubKey) {
// 				t.Errorf("Contract sender public keys do not match; c = %v, testContract = %v", tt.c.SenderPubKey, testContract.SenderPubKey)
// 			}
// 			if !reflect.DeepEqual(tt.c.SigLen, testContract.SigLen) {
// 				t.Errorf("Contract signature lengths do not match; c = %v, testContract = %v", tt.c.SigLen, testContract.SigLen)
// 			}
// 			if tt.c.Signature != nil {
// 				t.Errorf("Contract signatures do not match; c = %v, testContract = %v", tt.c.Signature, testContract.Signature)
// 			}
// 			if !reflect.DeepEqual(tt.c.RecipPubKeyHash, testContract.RecipPubKeyHash) {
// 				t.Errorf("Contract recipient public key hashes do not match; c = %v, testContract = %v", tt.c.RecipPubKeyHash, testContract.RecipPubKeyHash)
// 			}
// 			if !reflect.DeepEqual(tt.c.Value, testContract.Value) {
// 				t.Errorf("Contract values do not match; c = %v, testContract = %v", tt.c.Value, testContract.Value)
// 			}
// 			if !reflect.DeepEqual(tt.c.Nonce, testContract.Nonce) {
// 				t.Errorf("Contract nonces do not match; c = %v, testContract = %v", tt.c.Nonce, testContract.Nonce)
// 			}
// 		})
// 	}
// }

// func TestContract_SignContract(t *testing.T) {
// 	senderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	testContract, _ := MakeContract(1, *senderPrivateKey, senderPrivateKey.PublicKey, 25, 0)
// 	type args struct {
// 		sender ecdsa.PrivateKey
// 	}
// 	tests := []struct {
// 		name string
// 		c    *Contract
// 		args args
// 	}{
// 		{
// 			c: &testContract,
// 			args: args{
// 				sender: *senderPrivateKey,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			copyOfContract := testContract
// 			tt.c.SignContract(&tt.args.sender)
// 			serializedTestContract := block.HashSHA256(copyOfContract.Serialize(false))
// 			var esig struct {
// 				R, S *big.Int
// 			}
// 			if _, err := asn1.Unmarshal(tt.c.Signature, &esig); err != nil {
// 				t.Errorf("Failed to unmarshall signature")
// 			}
// 			if !ecdsa.Verify(&tt.c.SenderPubKey, serializedTestContract, esig.R, esig.S) {
// 				t.Errorf("Failed to verify valid signature")
// 			}
// 			maliciousPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 			if ecdsa.Verify(&maliciousPrivateKey.PublicKey, serializedTestContract, esig.R, esig.S) {
// 				t.Errorf("Failed to reject invalid signature")
// 			}
// 		})
// 	}
// }

// func TestContract_UpdateAccountBalanceTable(t *testing.T) {
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
// 	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)")
// 	statement.Exec()
// 	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
// 	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))), 350, 3)
// 	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
// 	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))), 200, 2)
// 	testContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 4)
// 	type args struct {
// 		table string
// 	}
// 	tests := []struct {
// 		name string
// 		c    *Contract
// 		args args
// 	}{
// 		{
// 			c: &testContract,
// 			args: args{
// 				table: dbName,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// defer func() {
// 			// 	if r := recover(); r != nil {
// 			// 		t.Errorf("Recovered from panic: %s", r)
// 			// 	}
// 			// }()
// 			tt.c.UpdateAccountBalanceTable(database)
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
