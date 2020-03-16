package publickey

var (
	Info = map[byte]publicKeyInfo{0x01: {Length: 65}}
)

type publicKeyInfo struct {
	Length uint
}
