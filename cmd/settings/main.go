package main

import (
	"encoding/json"
	"errors"
	"flag"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/SIGBlockchain/project_aurum/internal/config"
	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/jsonify"
)

// getBinDir returns a string of the project root bin directory
// TODO should this be placed in config or another package that is more appropraite
func getBinDir() string {
	return build.Default.GOPATH + constants.ProjectRoot + "bin/"
}

// getConfigFile opens a config file based on a filepath
// TODO even though this is one line it might increase readability - too much overhead?
func getConfigFile(path string) (*os.File, error) {
	return os.Open(path)
}

// setConfigFlags loads a configuration file into a Config struct, modifies the struct according to flags,
// and returns the updated struct
func setConfigFlags(configFile *os.File) (config.Config, error) {
	cfg := config.Config{}
	err := jsonify.LoadJSON(configFile, &cfg)
	if err != nil {
		return cfg, errors.New("Failed to unmarshall configuration data : " + err.Error())
	}

	//specify flags
	versionU64 := flag.Uint("version", uint(cfg.Version), "enter version number")
	cfg.Version = uint16(*versionU64) // ideally this and the above line would be combined
	flag.Uint64Var(&cfg.InitialAurumSupply, "supply", cfg.InitialAurumSupply, "enter a number for initial aurum supply")
	flag.StringVar(&cfg.Port, "port", cfg.Port, "enter port number")
	flag.StringVar(&cfg.BlockProductionInterval, "interval", cfg.BlockProductionInterval, "enter a time for block production interval\n(assuming seconds if units are not provided)")
	flag.BoolVar(&cfg.Localhost, "localhost", cfg.Localhost, "syntax: -localhost=/boolean here/")
	flag.StringVar(&cfg.MintAddr, "mint", cfg.MintAddr, "enter a mint address (64 characters hex string)")

	//read flags
	flag.Parse()

	// get units of interval
	intervalSuffix := strings.TrimLeftFunc(cfg.BlockProductionInterval, func(r rune) bool {
		return !unicode.IsLetter(r) && unicode.IsDigit(r)
	})
	// check units are valid
	hasSuf := false
	for _, s := range [7]string{"ns", "us", "µs", "ms", "s", "m", "h"} {
		if intervalSuffix == s {
			hasSuf = true
			break
		}
	}
	if !hasSuf {
		log.Fatalf("Failed to enter a valid interval suffix\nBad input: %v\n"+
			"Format should be digits and unit with no space e.g. 1h or 20s",
			cfg.BlockProductionInterval)
	}

	if len(cfg.MintAddr) != 64 && len(cfg.MintAddr) != 0 {
		log.Fatalf("Failed to enter a valid 64 character hex string for mint address.\n"+
			"Bad input: %v (len: %v)\n"+"The mint address must have 64 characters", cfg.MintAddr, len(cfg.MintAddr))
	}
	return cfg, nil
}

// main does runs four tasks: open a config file, set configuration flags, setup configuration as json,
// and write configurations into open file
func main() {
	// open config file
	configFile, err := getConfigFile(getBinDir() + constants.ConfigurationFile)
	if err != nil {
		log.Fatal("Failed to open configuration file: " + err.Error())
	}
	defer configFile.Close()

	// update config interface based on flags
	cfg, err := setConfigFlags(configFile)
	if err != nil {
		log.Fatal("Failed to set configuration: " + err.Error())
	}

	// convert cfg to bytes
	marshalledJSON, err := json.Marshal(cfg)
	if err != nil {
		log.Fatalf("Failed to marshal new config: %v", err)
	}

	// write bytes to file
	if err := ioutil.WriteFile(getBinDir(), marshalledJSON, 0644); err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}
