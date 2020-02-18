package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/SIGBlockchain/project_aurum/internal/config"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/validation"
)

// main does runs four tasks: open a config file, set configuration flags, setup configuration as json,
// and write configurations into open file
func main() {
	// open config file
	configFile, err := config.GetConfigFile(config.GetBinDir() + constants.ConfigurationFile)
	if err != nil {
		log.Fatal("Failed to open configuration file: " + err.Error())
	}
	defer configFile.Close()

	// update config interface based on flags
	cfg, err := validation.SetConfigFromFlags(configFile)
	if err != nil {
		log.Fatal("Failed to set configuration: " + err.Error())
	}

	// convert cfg to bytes
	marshalledJSON, err := json.Marshal(cfg)
	if err != nil {
		log.Fatalf("Failed to marshal new config: %v", err)
	}

	// write bytes to file
	if err := ioutil.WriteFile(config.GetBinDir()+constants.ConfigurationFile, marshalledJSON, 0644); err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}
