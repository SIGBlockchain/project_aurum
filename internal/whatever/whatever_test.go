package whatever

import "testing"

func TestCheck(t *testing.T) {
	s := [1]string{"Kyle"}
	c := check(s)
	if c != true {
		t.Error()
	}
}
