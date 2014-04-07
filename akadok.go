package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
)

const (
	configFile   = "server.cfg"
	masterServer = "https://akadok.deimos-ga.me"
)

var (
	config *akadokConfig
	log    *util.Logger
)

func main() {
	/* Config loading */

	loadConfig()

	/* Logging Engine */

	log = util.InitLogging(config.LogFile)
	// Log everything to file
	log.ToFile = true
	// Change debug mode if needed
	log.DebugMode = config.Verbose
	log.Notice("Akadok is loading...")

	// TODO

	/* Server IP Resolving */

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

	/* Heartbeat scheduling */

	/* Network routine start */

	log.Notice("Akadok has started")
}
