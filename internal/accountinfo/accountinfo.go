package accountinfo

import (
	"encoding/binary"
	"encoding/hex"
)


type AccountInfo struct {
	WalletAddress string
	Balance    uint64
	StateNonce uint64
}

func New(walletAddress string, balance uint64, stateNonce uint64) *AccountInfo {
	return &AccountInfo{Balance: balance, StateNonce: stateNonce, WalletAddress: walletAddress}
}

func (accInfo *AccountInfo) Serialize() ([]byte, error) {
	addressBytes, err := hex.DecodeString(accInfo.WalletAddress)
	serializedAccount := make([]byte, 16 + len(addressBytes)) // 8 + 8  + addressBytes bytes for walletAddress, balance, stateNonce
	if err != nil {
		return nil, err
	}
	copy(serializedAccount[:len(addressBytes)], addressBytes)
	binary.LittleEndian.PutUint64(serializedAccount[len(addressBytes):len(addressBytes)+8], accInfo.Balance)
	binary.LittleEndian.PutUint64(serializedAccount[len(addressBytes)+8:], accInfo.StateNonce)
	return serializedAccount, nil
}

func (accInfo *AccountInfo) Deserialize(serializedAccountInfo []byte) error {
    
	accInfo.WalletAddress = hex.EncodeToString(serializedAccountInfo[:178])
	accInfo.Balance = binary.LittleEndian.Uint64(serializedAccountInfo[178:186])
	accInfo.StateNonce = binary.LittleEndian.Uint64(serializedAccountInfo[186:])
	return nil
}
