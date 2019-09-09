package jsonify

import (
	"encoding/json"
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
