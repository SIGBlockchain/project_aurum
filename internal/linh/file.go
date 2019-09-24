package linh

func append2() []int {
	a := []int{1, 2, 4}
	b := []int{5, 6, 7}
	var c []int
	for i := 0; i < len(a); i++ {
		c = append(c, a[i])
	}

	for i := 0; i < len(b); i++ {
		c = append(c, b[i])
	}
	return c
}
