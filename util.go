package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"strings"
	"time"
)

// resolveIP uses the util package to resolve server external IP address
func resolveIP() {
	if config.Host == nil {
		log.Debug("Resolving external IP address...")
		ip := util.ResolveIP(masterServer)
		if ip != nil {
			config.Host = ip
		} else {
			log.Warn("Couldn't resolve external IP address!")
			config.Host = defaultConfig.Host
		}
		log.Debug("Resolved IP address is ", config.Host.String())
	}
}

// heartbeat inits the heartbeat to the master server
func heartbeat() {
	if !config.RegisterServer {
		return
	}
	go func() {
		tickChan := time.Tick(heartbeatInterval)
		for {
			select {
			case <-tickChan:
				log.Debug("Sending a heartbeat to the master server")

				heartbeatConfig := util.HeartbeatConfig{
					Ip:         config.Host.String(),
					Port:       config.Port,
					Name:       config.Name,
					PlayedMap:  currentMap,
					Players:    strings.Join(players, ", "),
					MaxPlayers: config.MaxPlayers,
				}
				err := util.Heartbeat(masterServer, &heartbeatConfig)

				if !masterServerLost && err != nil {
					log.Warn("Error while sending data to master server!")
					masterServerLost = true
				} else if masterServerLost {
					log.Notice("Regained connection with the master server")
				}
			}
		}
	}()

}

// initLogging creates the log object and sets its params
func initLogging() {
	log = util.InitLogging(config.LogFile)
	// Log everything to file
	log.ToFile = true
	// Change debug mode if needed
	log.DebugMode = config.Verbose
}
