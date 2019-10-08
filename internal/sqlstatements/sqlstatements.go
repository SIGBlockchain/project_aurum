package sqlstatements

const (
	GET_EVERYTHING_FROM_ACCOUNT_BALANCE_BY_WALLETADDRESS    = "SELECT * FROM account_balances WHERE public_key_hash = ?"
	GET_COUNT_EVERYTHING_FROM_METADATA                      = "SELECT COUNT(*) FROM METADATA"
	CREATE_ACCOUNT_BALANCES_TABLE                           = "CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)"
	CREATE_METADATA_TABLE                                   = "CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)"
	CREATE_PRODUCER_TABLE                                   = "CREATE TABLE IF NOT EXISTS producer (public_key_hash TEXT PRIMARY KEY, timestamp INTEGER)"
	INSERT_VALUES_INTO_ACCOUNT_BALANCES                     = "INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES(?, ?, ?)"
	INSERT_VALUES_INTO_METADATA                             = "INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)"
	UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH                 = "UPDATE account_balances set balance = ?, nonce = ? WHERE public_key_hash = ?"
	GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES    = "SELECT public_key_hash, balance, nonce FROM account_balances"
	GET_BALANCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH       = "SELECT balance FROM account_balances WHERE public_key_hash = ?"
	GET_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH         = "SELECT nonce FROM account_balances WHERE public_key_hash = ?"
	GET_HEIGHT_POSITION_SIZE_FROM_METADATA                  = "SELECT height, position, size FROM metadata"
	GET_POSITION_SIZE_FROM_METADATA                         = "SELECT position, size FROM metadata"
	GET_POSITION_SIZE_HASH_FROM_METADATA                    = "SELECT position, size, hash FROM metadata"
	GET_HEIGHT_FROM_METADATA                                = "SELECT height FROM metadata"
	GET_BALANCE_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH = "SELECT balance, nonce FROM account_balances WHERE public_key_hash = ?"
)
