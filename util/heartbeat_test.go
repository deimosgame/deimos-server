package util

import (
	"testing"
)

func TestHeartbeat(t *testing.T) {
	cfg := HeartbeatConfig{
		Ip:         "0.0.0.0",
		Port:       1518,
		Name:       "Test server - DO NOT CONNECT",
		PlayedMap:  "testmap1",
		Players:    "Artemis, Vomusseind",
		MaxPlayers: 42,
	}
	err := Heartbeat("https://akadok.deimos-ga.me", &cfg)
	if err != nil {
		t.Log("Heartbeat error - MAY BE RELATED TO NETWORK OR MASTER SERVER")
		t.Fail()
	}
}
