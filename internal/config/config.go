package config

import (
	"encoding/json"
	"errors"
	"go/build"
	"io/ioutil"
	"os"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
)

type Config struct {
	Version                 uint16
	InitialAurumSupply      uint64
	Port                    string
	BlockProductionInterval string
	Localhost               bool
	MintAddr                string
}

// GetBinDir returns a string of the project root bin directory
// TODO does this need a test?
func GetBinDir() string {
	return build.Default.GOPATH + constants.ProjectRoot + "bin/"
}

// GetConfigFile opens a config file based on a filepath and returns an error if error occurs
// TODO does this need a test?
func GetConfigFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Failed to load configuration file : " + err.Error())
	}
	return file, nil
}

// LoadConfiguration fills a Config struct with configuration info from a file and returns
// the struct
func LoadConfiguration() (*Config, error) {
	configFile, err := GetConfigFile(constants.ConfigurationFile)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	cfgData, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, errors.New("Failed to read configuration file : " + err.Error())
	}

	cfg := Config{}
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return nil, errors.New("Failed to unmarshall configuration data : " + err.Error())
	}

	return &cfg, nil
}
