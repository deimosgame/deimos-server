package main

import (
	"time"
)

const (
	configFile   = "server.cfg"
	masterServer = "https://akadok.deimos-ga.me"
)

var (
	config *akadokConfig
	log    *util.Logger

	// Game-related variables
	currentMap string
	players    []string
)

func main() {
	/* Config loading */

	loadConfig()

	/* Logging Engine */

	initLogging()

	log.Notice("Akadok is loading...")

	/* Server IP Resolving */

	resolveIP()

	/* Heartbeat scheduling */

	heartbeat()

	// TODO

	/* Network routine start */

	log.Notice("Akadok has started")

	// Keeps the server idle ATM
	for {
		time.Sleep(time.Millisecond * 50)
	}
}
