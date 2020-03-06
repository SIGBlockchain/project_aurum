package sqlstatements

const (
	GET_EVERYTHING_FROM_ACCOUNT_BALANCE_BY_WALLETADDRESS    = "SELECT * FROM account_balances WHERE public_key_hash = ?"
	GET_COUNT_EVERYTHING_FROM_METADATA                      = "SELECT COUNT(*) FROM METADATA"
	CREATE_ACCOUNT_BALANCES_TABLE                           = "CREATE TABLE IF NOT EXISTS account_balances (public_key_hash TEXT, balance INTEGER, nonce INTEGER)"
	CREATE_METADATA_TABLE                                   = "CREATE TABLE IF NOT EXISTS metadata (height INTEGER PRIMARY KEY, position INTEGER, size INTEGER, hash TEXT)"
	CREATE_PRODUCER_TABLE                                   = "CREATE TABLE IF NOT EXISTS producer (public_key_hash TEXT PRIMARY KEY, timestamp INTEGER)"
	CREATE_CONTRACT_TABLE                                   = "CREATE TABLE IF NOT EXISTS contracts (serialized_contract BINARY PRIMARY KEY, sender_pub_key TEXT, nonce INTEGER)"
	INSERT_VALUES_INTO_ACCOUNT_BALANCES                     = "INSERT INTO account_balances (public_key_hash, balance, nonce) VALUES(?, ?, ?)"
	INSERT_VALUES_INTO_METADATA                             = "INSERT INTO metadata (height, position, size, hash) VALUES (?, ?, ?, ?)"
	INSERT_VALUES_INTO_PRODUCER                             = "INSERT INTO producer (public_key_hash, timestamp) VALUES (?, ?)"
	UPDATE_ACCOUNT_BALANCES_BY_PUB_KEY_HASH                 = "UPDATE account_balances set balance = ?, nonce = ? WHERE public_key_hash = ?"
	GET_PUB_KEY_HASH_BALANCE_NONCE_FROM_ACCOUNT_BALANCES    = "SELECT public_key_hash, balance, nonce FROM account_balances"
	GET_BALANCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH       = "SELECT balance FROM account_balances WHERE public_key_hash = ?"
	GET_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH         = "SELECT nonce FROM account_balances WHERE public_key_hash = ?"
	GET_HEIGHT_POSITION_SIZE_FROM_METADATA                  = "SELECT height, position, size FROM metadata"
	GET_POSITION_SIZE_FROM_METADATA                         = "SELECT position, size FROM metadata"
	GET_POSITION_SIZE_HASH_FROM_METADATA                    = "SELECT position, size, hash FROM metadata"
	GET_HEIGHT_FROM_METADATA                                = "SELECT height FROM metadata"
	GET_BALANCE_NONCE_FROM_ACCOUNT_BALANCES_BY_PUB_KEY_HASH = "SELECT balance, nonce FROM account_balances WHERE public_key_hash = ?"
	GET_BATCH_OF_BLOCKS_FROM_METADATA                       = "SELECT height, position, size FROM metadata WHERE height BETWEEN ? and ? + ? - 1 ORDER BY height"
	// GET_BATCH_OF_BLOCKS_FROM_METADATA variables: (startHeight, startHeight, numBlocks)
	GET_CONTRACT_BY_SENDER_PUBLIC_KEY_AND_NONCE = "SELECT serialized_contract from contracts WHERE sender_pub_key = ? and nonce = ?"
)
