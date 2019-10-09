package jsonify

import (
	"io/ioutil"
	"testing"

	"os"
)

type nilInterface interface {
	M()
}

func TestLoadJSON(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "jsontest")
	var testNilInterface nilInterface
	file, err := os.Open("tmpfile")
	if err != nil {
		t.Errorf("failed to open file %v", err)
	}

	LoadJSON(file, &testNilInterface)
	if testNilInterface == nil {
		t.Errorf("empty interface")
	}

	file.Close()
	os.Remove(tmpfile.Name())
}
