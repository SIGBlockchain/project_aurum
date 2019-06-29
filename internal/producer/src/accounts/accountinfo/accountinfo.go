package accountinfo

import "encoding/binary"

type AccountInfo struct {
	Balance    uint64
	StateNonce uint64
}

func NewAccountInfo(balance uint64, stateNonce uint64) *AccountInfo {
	return &AccountInfo{Balance: balance, StateNonce: stateNonce}
}

func (accInfo *AccountInfo) Serialize() ([]byte, error) {
	serializedAccount := make([]byte, 16) // 8 + 8 bytes for balance and stateNonce
	binary.LittleEndian.PutUint64(serializedAccount[:8], accInfo.Balance)
	binary.LittleEndian.PutUint64(serializedAccount[8:], accInfo.StateNonce)
	return serializedAccount, nil
}

func (accInfo *AccountInfo) Deserialize(serializedAccountInfo []byte) error {
	accInfo.Balance = binary.LittleEndian.Uint64(serializedAccountInfo[:8])
	accInfo.StateNonce = binary.LittleEndian.Uint64(serializedAccountInfo[8:])
	return nil
}
