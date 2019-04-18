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

// func TestValidateContrac(t *testing.T) {
// 	table := "table.db"
// 	conn, err := sql.Open("sqlite3", table)
// 	defer conn.Close()

// 	defer func() {
// 		os.Remove(table)
// 	}()
// 	statement, err := conn.Prepare(
// 		`CREATE TABLE IF NOT EXISTS account_balances (
// 		public_key_hash TEXT,
// 		balance INTEGER,
// 		nonce INTEGER);`)

// 	if err != nil {
// 		t.Error(err)
// 	}
// 	statement.Exec()

// 	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	senderPublicKey := sender.PublicKey
// 	recipientPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	recipient := recipientPrivateKey.PublicKey
// 	c := MakeContract(1, *sender, senderPublicKey, 1000, 0)
// 	// INSERT ABOVE VALUES INTO TABLE
// 	sqlQuery := fmt.Sprintf("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (\"%s\", %d, %d)", hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&c.SenderPubKey))), 1000, 0)
// 	_, err = conn.Exec(sqlQuery)
// 	if err != nil {
// 		t.Errorf("Unable to insert values into table: %v", err)
// 	}

// 	// Query database here, make sure that the sender has 1000 aurum

// 	validContract := MakeContract(1, *sender, recipient, 1000, 1)
// 	falseContractInsufficientFunds := MakeContract(1, *recipientPrivateKey, senderPublicKey, 2000, 0)
// 	if !ValidateContract(validContract, table) {
// 		t.Errorf("Valid contract regarded as invalid")
// 	}

// 	// Query the database here, make sure the sender has 0 aurum
// 	// and the recipient has 1000 aurum

// 	if ValidateContract(falseContractInsufficientFunds, table) {
// 		t.Errorf("Invalid contract regarded as valid")
// 	}

// 	// Query the database again and make sure everything is still the same
// }

// func TestContractSerializatio(t *testing.T) {
// 	sender, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	senderPublicKey := sender.PublicKey
// 	contract := MakeContract(1, *sender, senderPublicKey, 1000, 20)
// 	serialized := contract.Serialize()
// 	deserialized := Contract{}.Deserialize(serialized)
// 	if !reflect.DeepEqual(contract, deserialized) {
// 		t.Errorf("Contracts (struct) do not match")
// 	}
// 	// if !cmp.Equal(contract, deserialized, cmp.AllowUnexported(Contract{})) {
// 	// 	t.Errorf("Contracts (struct) do not match")
// 	// }
// 	reserialized := deserialized.Serialize()
// 	if !bytes.Equal(serialized, reserialized) {
// 		t.Errorf("Contracts (byte slice) do not match")
// 	}
// }

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
				testContract.Serialize(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Deserialize(tt.args.b)
			// if !reflect.DeepEqual(tt.c, testContract) {
			// 	t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			// }
			if !reflect.DeepEqual(tt.c.Version, testContract.Version) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			}
			if !reflect.DeepEqual(tt.c.SenderPubKey, testContract.SenderPubKey) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			}
			if !reflect.DeepEqual(tt.c.SigLen, testContract.SigLen) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			}
			// if tt.c.Signature != nil {
			// 	t.Errorf("Contract signatures do not match; c = %v, testContract = %v", *tt.c, testContract)
			// }
			if !reflect.DeepEqual(tt.c.RecipPubKeyHash, testContract.RecipPubKeyHash) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			}
			if !reflect.DeepEqual(tt.c.Value, testContract.Value) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
			}
			if !reflect.DeepEqual(tt.c.Nonce, testContract.Nonce) {
				t.Errorf("Contracts do not match; c = %v, testContract = %v", *tt.c, testContract)
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
		},
	}
    for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// defer func() {
			// 	if r := recover(); r != nil {
			// 		t.Errorf("Recovered from panic: %s", r)
			// 	}
			// }()
			tt.c.SignContract(&tt.args.sender)
			hashedContract := block.HashSHA256(testContract.Serialize())
			var esig struct {
				R, S *big.Int
			}
			if _, err := asn1.Unmarshal(tt.c.Signature, &esig); err != nil {
				t.Errorf("Failed to unmarshall signature")
			}
			if !ecdsa.Verify(&tt.c.SenderPubKey, hashedContract, esig.R, esig.S) {
				t.Errorf("Failed to verify valid signature")
			}
			maliciousPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if ecdsa.Verify(&maliciousPrivateKey.PublicKey, hashedContract, esig.R, esig.S) {
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
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&senderPrivateKey.PublicKey))), 350, 3)
	statement, _ = database.Prepare("INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES (?, ?, ?)")
	statement.Exec(hex.EncodeToString(block.HashSHA256(keys.EncodePublicKey(&recipientPrivateKey.PublicKey))), 200, 2)
	validContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 4)
	validContract.SignContract(*senderPrivateKey)
	invalidContract, _ := MakeContract(1, *senderPrivateKey, recipientPrivateKey.PublicKey, 350, 5)
	invalidContract.SignContract(*senderPrivateKey)
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
		{
			name: "Invalid contract by value",
			args: args{
				c: invalidContract,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateContract(tt.args.c, tt.args.tableName); got != tt.want {
				t.Errorf("ValidateContract() = %v, want %v", got, tt.want)
			}
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
