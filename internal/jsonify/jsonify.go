package jsonify

import (
	"encoding/json"
	"errors"
	"io"
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

func DumpJSON(writer io.Writer, iface interface{}) error {
	bytes, err := json.Marshal(iface)
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return errors.New("Failed to write to file")
	}
	return nil
}
