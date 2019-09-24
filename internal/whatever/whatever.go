package whatever

func check(s [1]string) bool {
	for i := 0; i < len(s); i++ {
		if len(s[i]) < 5 {
			return true
		}
	}
	return false
}
