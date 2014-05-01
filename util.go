package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ResolveIP uses the util package to resolve server external IP address
func ResolveIP() {
	if config.Host != nil {
		return
	}
	log.Debug("Resolving external IP address...")
	ip := util.ResolveIP(MasterServer)
	if ip != nil {
		config.Host = ip
	} else {
		log.Warn("Couldn't resolve external IP address!")
		config.Host = defaultConfig.Host
	}
	log.Info(fmt.Sprintf("Server IP address is %s:%d", config.Host.String(),
		config.Port))
}

// Heartbeat is responsible of the heartbeat to the master server
func Heartbeat() {
	if !config.RegisterServer {
		return
	}
	for {
		log.Debug("Sending a heartbeat to the master server")

		// Generating player list
		playerList, i := make([]string, len(players)), 0
		for _, v := range players {
			playerList[i] = v.Name
			i++
		}

		heartbeatConfig := util.HeartbeatConfig{
			Ip:         config.Host.String(),
			Port:       config.Port,
			Name:       config.Name,
			PlayedMap:  currentMap,
			Players:    strings.Join(playerList, ", "),
			MaxPlayers: config.MaxPlayers,
		}
		err := util.Heartbeat(MasterServer, &heartbeatConfig)

		if !masterServerLost && err != nil {
			log.Warn("Error while sending data to master server!")
			masterServerLost = true
		} else if masterServerLost {
			log.Notice("Regained connection with the master server")
			masterServerLost = false
		}
		time.Sleep(HeartbeatInterval)
	}
}

// InitLogging creates the log object and sets its params
func InitLogging() {
	log = util.InitLogging(config.LogFile)
	// Log everything to file
	log.ToFile = true
	// Change debug mode if needed
	log.DebugMode = config.Verbose
}

// CheckToken verifies a token with a user id
func CheckToken(deimosId, token string) (bool, error) {
	apiUrl := "https://deimos-ga.me/api/validate-token/"
	resp, err := http.Get(apiUrl + deimosId + "/" + token)
	if err != nil {
		return CheckInsecure(), err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CheckInsecure(), err
	}
	type Answer struct {
		Success bool
	}
	answer := Answer{}
	err = json.Unmarshal(body, &answer)
	if err != nil {
		return CheckInsecure(), err
	}
	if insecureAlert {
		insecureAlert = false
		log.Notice("Authentication server has been reached. Server is now in secure mode")
	}
	return answer.Success || config.Insecure, nil
}

// CheckInsecure emits an alert for insecure mode if needed
func CheckInsecure() bool {
	if !config.AutoInsecure {
		return false
	}
	if !insecureAlert {
		insecureAlert = true
		log.Notice("Authentication server is unreachable. Server is now in insecure mode")
	}
	return true
}
