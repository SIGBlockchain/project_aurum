package sqlqueries

const (
	// starting from the top did handler...

	// ASK ABOUT PREPARE STMTS... CREATING TABLES... INSERTING in this case, left off in handlers_test.go
	GET_EVERYTHING_BY_WALLETADDRESS        = `SELECT * FROM account_balances WHERE public_key_hash = "`
	GET_BALANCE_BY_PUB_KEY_HASH            = "SELECT balance FROM account_balances WHERE public_key_hash = \""
	GET_NONCE_BY_PUB_KEY_HASH              = "SELECT nonce FROM account_balances WHERE public_key_hash= \""
	GET_HEIGHT_POSITION_SIZE_FROM_METADATA = "SELECT height, position, size FROM metadata"
	GET_POSITION_SIZE_FROM_METADATA        = "SELECT position, size FROM metadata"
	GET_POSITION_SIZE_HASH_FROM_METADATA   = "SELECT position, size, hash FROM metadata"
	GET_HEIGHT_FROM_METADATA               = "SELECT height FROM metadata"
	//finished blockchain.go...
)
