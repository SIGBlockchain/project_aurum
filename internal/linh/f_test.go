package linh

import (
	"fmt"
	"testing"
)

func TestFunc(t *testing.T) {
	append2()
	fmt.Print(append2())
	if len(append2()) != 6 {
		t.Errorf("failure")
	}
	//t.Errorf("failure")
}
