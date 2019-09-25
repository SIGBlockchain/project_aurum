package test

import "testing"

func TestTestFunc(t *testing.T) {
	rslt := subtart()

	if rslt != 1 {
		t.Errorf("wrong value")
	}
}
