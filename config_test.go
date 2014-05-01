package main

import (
	"os"
	"testing"
)

// Simple test to ensure everything does not blows up
func TestConfigLoading(t *testing.T) {
	// Executes almost everything in a clean test environement: file creation,
	// reading and parsing.
	LoadConfig()
	// Test for a few values
	if config.Name != defaultConfig.Name ||
		config.Port != defaultConfig.Port ||
		config.MaxPlayers != defaultConfig.MaxPlayers ||
		config.Verbose != defaultConfig.Verbose ||
		config.LogFile != defaultConfig.LogFile ||
		config.RegisterServer != defaultConfig.RegisterServer {
		t.Fail()
	}
	// Cleaning
	os.Remove("server.cfg")
}

func TestNormalizeName(t *testing.T) {
	if NormalizeName("ThisIsATest") != "this_is_a_test" {
		t.Fail()
	}
}

func TestUnNormalizeName(t *testing.T) {
	if UnNormalizeName("this_is_a_test") != "ThisIsATest" {
		t.Fail()
	}
}
