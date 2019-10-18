package jsonify

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

type book struct {
	Author string
	Title  string
}

func TestLoadJSON(t *testing.T) {
	//arrange
	var testBook book
	tmpfile, err := ioutil.TempFile("", "jsontest") //create tempfile and open it
	data := book{
		Author: "steve",
		Title:  "world",
	}
	file, _ := json.MarshalIndent(data, "", " ")       //converts structs into array of bytes form
	err = ioutil.WriteFile(tmpfile.Name(), file, 0644) // writes a array of bytes(the data)
	if err != nil {
		t.Errorf("failed to write file %v", err)
	}

	defer func() {
		err := tmpfile.Close() //closes file
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)
		}

		err = os.Remove(tmpfile.Name()) //deletes the file
		if err != nil {
			t.Errorf("Failed to remove database: %s", err)

		}
	}()

	//act
	LoadJSON(tmpfile, &testBook)
	//assert
	if testBook.Author != "steve" {
		t.Errorf("Expected: %v , got: %v", "steve", testBook.Author)
	}

	if testBook.Title != "world" {
		t.Errorf("Expected: %v , got: %v", "world", testBook.Title)
	}
}

func TestDumpJSON(t *testing.T) {
	// arrange
	iFace := book{
		Author: "Dump",
		Title:  "JSON",
	}
	file, err := os.Create("temp") // created and opened
	data, _ := json.Marshal(file)  //converts structs into array of bytes form
	_, err = file.Write(data)      // writes a array of bytes(the data)
	if err != nil {
		t.Errorf("failed to write file %v", err)
	}

	defer func() {
		err := file.Close() //closes file
		if err != nil {
			t.Errorf("Failed to close file: %s", err)
		}

		err = os.Remove(file.Name()) //deletes the file
		if err != nil {
			t.Errorf("Failed to delete file: %s", err)

		}
	}()
	//act
	err = DumpJSON(file, iFace)
	// assert
	if "Dump" != iFace.Author {
		t.Errorf("the strings are not equal")
	}
	if "JSON" != iFace.Title {
		t.Errorf("the strings are not equal")
	}

}
