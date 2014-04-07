package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
)

const (
	configFile = "server.cfg"
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

	/* Heartbeat scheduling */

	/* Network routine start */

	log.Notice("Akadok has started")
}
