package accountinfo

import (
	"reflect"
	"testing"
)

func TestAccountInfo_Deserialize(t *testing.T) {
	ac := New("wallet", 9001, 50)
	serAc, _ := ac.Serialize()
	type args struct {
		serializedAccountInfo []byte
	}
	tests := []struct {
		name    string
		accInfo *AccountInfo
		args    args
		wantErr bool
	}{
		{
			accInfo: &AccountInfo{},
			args:    args{serializedAccountInfo: serAc},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.accInfo.Deserialize(tt.args.serializedAccountInfo); (err != nil) != tt.wantErr {
				t.Errorf("AccountInfo.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.accInfo, ac) {
				t.Errorf("structs do not match")
			}
		})
	}
}
