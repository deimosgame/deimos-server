package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"fmt"
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

// Heartbeat inits the heartbeat to the master server
func Heartbeat() {
	if !config.RegisterServer {
		return
	}
	go func() {
		// heartbeatCaller wraps util.Heartbeat with config generation
		HeartbeatCaller := func() {
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
		}

		// Produces a heartbeat every so often
		tickChan := time.Tick(HeartbeatInterval)

		HeartbeatCaller()

		for {
			select {
			case <-tickChan:
				HeartbeatCaller()
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

// InitLogging creates the log object and sets its params
func InitLogging() {
	log = util.InitLogging(config.LogFile)
	// Log everything to file
	log.ToFile = true
	// Change debug mode if needed
	log.DebugMode = config.Verbose
}
