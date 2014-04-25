package main

import (
	"os"
	"testing"
)

func TestInitLogging(t *testing.T) {
	InitLogging()
	log.Debug("Testing logging initialization")
	if _, err := os.Stat("server.log"); os.IsNotExist(err) {
		t.Log("Error creating the log file")
		t.Fail()
	}
	os.Remove("server.lof")
}
