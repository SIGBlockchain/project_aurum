package mock

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestMockIoReaderGenerator(t *testing.T) {
	// arrange
	m := MockIoReader{}
	testString := "abcde"

	// act
	m.When("Read").Given([]byte(testString)).Return(5, io.EOF)

	// assert
	if err := m.Check(); err != nil {
		t.Error("Failed check")
	}
	if !bytes.Equal(m.Buffer, []byte(testString)) {
		t.Errorf("Wrong buffer. Got %v want %v", m.Buffer, []byte(testString))
	}
	if m.NRead != 5 {
		t.Errorf("Wrong number bytes read. Got %v want %v", m.NRead, 5)
	}
	if m.Error != io.EOF {
		t.Errorf("Wrong error. Got %v want %v", m.Error, io.EOF)
	}
	res, err := ioutil.ReadAll(m)
	if err != nil {
		t.Errorf("Failed ReadAll call: " + err.Error())
	}
	if string(res) != testString {
		t.Errorf("Wrong output. Got %v want %v", string(res), testString)
	}
}
