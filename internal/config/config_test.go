package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
)

func TestLoadConfigurationFile(t *testing.T) {
	cfg := Config{1, 20, "5000", "40s", false}
	marshalledCfg, err := json.Marshal(cfg)
	if err != nil {
		t.Errorf("failed to marshall configuration struct: %v", err)
	}
	if err := ioutil.WriteFile(constants.ConfigurationFile, marshalledCfg, os.ModePerm); err != nil {
		t.Errorf("failed to write to file: %v", err)
	}
	defer func() {
		if err := os.Remove(constants.ConfigurationFile); err != nil {
			t.Errorf("failed to remove config file: %v", err)
		}
	}()
	fileCfg, err := LoadConfiguration()
	if err != nil {
		t.Errorf("failed to load configuration file: %v", err)
	}
	if !reflect.DeepEqual(cfg, *fileCfg) {
		t.Errorf("structs to not match: got %+v want %+v", *fileCfg, cfg)
	}
}
