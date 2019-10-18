package jsonify

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// load file into interface
func LoadJSON(file *os.File, iFace interface{}) error {
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(fileData, &iFace); err != nil {
		return err
	}

	return nil
}

// DumpJSON places interface info into a new file and return that file
func DumpJSON(file *os.File, iFace interface{}) error {
	bytes, err := json.Marshal(iFace)
	if err != nil {
		return err
	}
	_, err = file.Write(bytes)
	if err != nil {
		return errors.New("Failed to write to file")
	}
	return nil
}
