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

func TestMockBlockFetcherGenerator(t *testing.T) {
	// arrange
	m := MockBlockFetcher{}
	blocks := [][]byte{[]byte("test1"), []byte("test2")}
	errors := []error{nil, nil}

	// act
	m.When("FetchBlockByHeight").Given(1058).Return(blocks[0], errors[0])
	m.When("FetchBlockByHeight").Given(2043).Return(blocks[1], errors[1])

	//assert
	if !bytes.Equal(m.SerializedBlocks[0], blocks[0]) {
		t.Errorf("Did not set correct block return. Got %v wanted %v", m.SerializedBlocks[0], blocks[0])
	}
	if !bytes.Equal(m.SerializedBlocks[1], blocks[1]) {
		t.Errorf("Did not set correct block return. Got %v wanted %v", m.SerializedBlocks[0], blocks[0])
	}
	if m.Errors[0] != errors[0] {
		t.Errorf("Did not set correct error return. Got %v wanted %v", m.Errors[0], errors[0])
	}
	if m.Errors[1] != errors[1] {
		t.Errorf("Did not set correct error return. Got %v wanted %v", m.Errors[0], errors[0])
	}
	if b, e := m.FetchBlockByHeight(1058); (!bytes.Equal(b, blocks[0])) || (e != errors[0]) {
		t.Errorf("Got wrong returns from function call. Got %v, %v, wanted %v, %v", b, e, blocks[0], errors[0])
	}
	if b, e := m.FetchBlockByHeight(2043); (!bytes.Equal(b, blocks[1])) || (e != errors[1]) {
		t.Errorf("Got wrong returns from function call. Got %v, %v, wanted %v, %v", b, e, blocks[1], errors[1])
	}
}
