package main

import "testing"

func TestCheckConnectivity(t *testing.T) {
	// Test will fail in airplane mode, or just remove wireless connection.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Internet connection check failed.")
		}
	}()
	CheckConnectivity()
}
