package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
	log.Close()
	os.Remove("server.log")
}

func TestCheckToken(t *testing.T) {
	// Get a real token for our dummy account
	apiUrl, user, password := "https://deimos-ga.me/api/get-token/",
		"test@deimos-ga.me", "testacc"
	resp, err := http.Get(apiUrl + user + "/" + password)
	if err != nil {
		t.Log("Possible network communication issues with the master server")
		t.Fail()
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
		return
	}
	type TokenQueryResponse struct {
		Success bool
		Token   string
	}
	tokenResponse := TokenQueryResponse{}
	json.Unmarshal(body, &tokenResponse)

	// Check this token against the API
	checkResult, err := CheckToken(user, tokenResponse.Token)
	if err != nil || checkResult == false {
		t.Fail()
		return
	}
}
