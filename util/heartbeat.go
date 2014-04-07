package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type HeartbeatConfig struct {
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Name       string `json:"name"`
	PlayedMap  string `json:"map"`
	Players    string `json:"players"`
	MaxPlayers int    `json:"maxplayers"`
}

func Heartbeat(masterServer string, cfg *HeartbeatConfig) (err error) {
	client := &http.Client{}

	encodedJson, _ := json.Marshal(cfg)
	r, _ := http.NewRequest("POST", masterServer, bytes.NewBuffer(encodedJson))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", strconv.Itoa(len(encodedJson)))

	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	stringResult, err := ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(stringResult), "true") {
		return errors.New("Master server gave an unexpected response")
	}
	return nil
}
