package jsonify

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

//TODO - improve this description: stores file information as structured json
type Config struct {
	Version                 uint16
	InitialAurumSupply      uint64
	Port                    string
	BlockProductionInterval string
	Localhost               bool
}

// is the assumption that the file is already open?
// if the function is being returned and the interface needs to be defined locally, why pass an interface?
// ok to return an error?
// convert file to interface 
func toJSON(file *os.File, config interface{}) (*Config, error) {
	cfgData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Failed to read configuration file : " + err.Error())
	}

	cfg := Config{}
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, errors.New("Failed to unmarshall configuration data : " + err.Error())
	}

	return &cfg, nil
}