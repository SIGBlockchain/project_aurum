package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/SIGBlockchain/project_aurum/internal/config"
	"github.com/SIGBlockchain/project_aurum/internal/jsonify"
)

func main() {
	//open configuration file
	configFile, err := os.Open("../config.json")
	if err != nil {
		log.Fatal("Failed to open configuration file : " + err.Error())
	}
	defer configFile.Close()

	cfg := config.Config{}
	err = jsonify.LoadJSON(configFile, &cfg)
	if err != nil {
		log.Fatal("Failed to unmarshall configuration data : " + err.Error())
	}

	//specify flags
	versionU64 := flag.Uint("version", uint(cfg.Version), "enter version number")
	flag.Uint64Var(&cfg.InitialAurumSupply, "supply", cfg.InitialAurumSupply, "enter a number for initial aurum supply")
	flag.StringVar(&cfg.Port, "port", cfg.Port, "enter port number")
	flag.StringVar(&cfg.BlockProductionInterval, "interval", cfg.BlockProductionInterval, "enter a time for block production interval\n(assuming seconds if units are not provided)")
	flag.BoolVar(&cfg.Localhost, "localhost", cfg.Localhost, "syntax: -localhost=/boolean here/")

	//read flags
	flag.Parse()

	cfg.Version = uint16(*versionU64)

	// check that interval is valid
	validInterval, _ := regexp.MatchString("[0-9]", string([]rune(cfg.BlockProductionInterval)[0]))
	if !validInterval {
		log.Fatalf("Interval input did not start with digit")
	}

	validInterval, _ = regexp.MatchString("[a-s]", string([]rune(cfg.BlockProductionInterval)[len(cfg.BlockProductionInterval)-1]))
	if !validInterval {
		log.Fatalf("Interval input did not end with letter")
	}
	//check block production interval suffix
	hasSuf := false

	for _, s := range [7]string{"ns", "us", "µs", "ms", "s", "m", "h"} {
		if strings.HasSuffix(cfg.BlockProductionInterval, s) {
			hasSuf = true
			break
		}
	}
	if !hasSuf {
		log.Fatalf("Failed to enter a valid interval suffix\nBad input: %v\n"+
			"Format should be digits and unit with no space e.g. 1h or 20s",
			cfg.BlockProductionInterval)
	}

	//write into configuration file
	marshalledJSON, err := json.Marshal(cfg)
	if err != nil {
		log.Fatalf("Failed to marshal new config: %v", err)
	}

	if err := ioutil.WriteFile("../config.json", marshalledJSON, 0644); err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}
