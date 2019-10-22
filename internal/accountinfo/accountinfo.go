package accountinfo

import (
	"encoding/binary"
	"encoding/hex"
)


type AccountInfo struct {
	Balance    uint64
	StateNonce uint64
	WalletAddress string
}

func New(walletAddress string, balance uint64, stateNonce uint64) *AccountInfo {
	return &AccountInfo{Balance: balance, StateNonce: stateNonce, WalletAddress: walletAddress}
}

func (accInfo *AccountInfo) Serialize() ([]byte, error) {
	addressBytes, err := hex.DecodeString(accInfo.WalletAddress)
	serializedAccount := make([]byte, 48) // 8 + 8  + 32 bytes for walletAddress, balance, stateNonce
	if err != nil {
		return nil, err
	}
	serializedAccount = append(serializedAccount, addressBytes...)
	binary.LittleEndian.PutUint64(serializedAccount[:40], accInfo.Balance)
	binary.LittleEndian.PutUint64(serializedAccount[40:], accInfo.StateNonce)
	return serializedAccount, nil
}

func (accInfo *AccountInfo) Deserialize(serializedAccountInfo []byte) error {
	accInfo.Balance = binary.LittleEndian.Uint64(serializedAccountInfo[:8])
	accInfo.StateNonce = binary.LittleEndian.Uint64(serializedAccountInfo[8:])
	return nil
}
