package jsonify

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// load file into interface
func loadJSON(file *os.File, inrface interface{}) (*interface{}, error) {
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fileData, &inrface); err != nil {
		return nil, err
	}

	return &inrface, nil
}
