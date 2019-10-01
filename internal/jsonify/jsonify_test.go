package jsonify

import (
	"os"
	"testing"
)

type book struct {
	Author string
	Title  string
}

func TestLoadJSON(t *testing.T) {
	file, err := os.Open("test.json")
	if err != nil {
		t.Errorf("failed to open file %v", err)
	}
	var testBook book
	LoadJSON(file, &testBook)
	if testBook.Author != "steve" {
		t.Errorf("Expected: %v , got: %v", "steve", testBook.Author)
	}

	if testBook.Title != "world" {
		t.Errorf("Expected: %v , got: %v", "world", testBook.Title)
	}

}
