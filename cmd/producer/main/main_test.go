package main

import (
	"bytes"
	"database/sql"
	"os"
	"os/exec"
	"testing"

	"github.com/SIGBlockchain/project_aurum/internal/constants"
	"github.com/SIGBlockchain/project_aurum/internal/sqlstatements"
)

var removeFiles = true

func TestSuite(t *testing.T) {
	defer func() { // Function is dangerous, consider only running with flag
		if removeFiles {
			if err := os.Remove("blockchain.dat"); err != nil {
				t.Errorf("failed to remove blockchain.dat:\n%s", err.Error())
			}
			if err := os.Remove(constants.MetadataTable); err != nil {
				t.Errorf("failed to remove metadatata.tab:\n%s", err.Error())
			}
			if err := os.Remove(constants.AccountsTable); err != nil {
				t.Errorf("failed to remove accounts.db:\n%s", err.Error())
			}
		}
	}()
	type testArg struct {
		name       string
		runCommand []string
	}
	testArgs := []testArg{
		{
			name:       "Genesis",
			runCommand: []string{"go", "run", "main.go", "-d", "-g", "--supply=100"},
		},
		{
			name:       "Loop",
			runCommand: []string{"go", "run", "main.go", "-d", "--interval=1000ms", "--blocks=3"},
		},
	}
	var stdout bytes.Buffer
	for _, arg := range testArgs {
		t.Run(arg.name, func(t *testing.T) {
			cmd := exec.Command(arg.runCommand[0], arg.runCommand[1:]...)
			cmd.Stdout = &stdout
			if err := cmd.Start(); err != nil {
				t.Errorf("failed to run main command because: %s", err.Error())
			}
			if err := cmd.Wait(); err != nil {
				t.Errorf("main call returned with: %s.", err.Error())
				t.Logf("Stdout: %s", string(stdout.Bytes()))
			}
		})
	}

	dbc, err := sql.Open("sqlite3", constants.MetadataTable)
	if err != nil {
		t.Errorf("failed to open sqlite database")
	}
	defer func() {
		if err := dbc.Close(); err != nil {
			t.Errorf("failed to close database connection")
		}
	}()
	var count int
	var expectedCount = 4
	if err := dbc.QueryRow(sqlstatements.GET_COUNT_EVERYTHING_FROM_METADATA).Scan(&count); err != nil {
		t.Errorf("failed to query rows")
	}
	if count != expectedCount {
		t.Errorf("invalid number of blocks: %d != %d", count, expectedCount)
	}
}
