package jsonify

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// is the assumption that the file is already open?
// ok to return an error?
// convert file to interface
func toInterface(file *os.File, iFace interface{}) (*interface{}, error) {
	cfgData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Failed to read configuration file : " + err.Error())
	}

	cfg := iFace
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, errors.New("Failed to unmarshall configuration data : " + err.Error())
	}

	return &cfg, nil
}
