package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"time"
)

const (
	masterServer      = "https://akadok.deimos-ga.me"
	heartbeatInterval = 15 * time.Second
)

var (
	config           *akadokConfig
	log              *util.Logger
	configFile       = "server.cfg"
	masterServerLost = false

	// Game-related variables
	currentMap string
	players    []string
)

func main() {
	// temp inititalization
	currentMap = "coolmap"
	players = []string{"Artemis", "Vomusseind"}

	/* Config loading */

	loadConfig()

	/* Logging Engine */

	initLogging()

	log.Notice("Akadok is loading...")

	/* Server IP Resolving */

	resolveIP()

	/* Heartbeat scheduling */

	heartbeat()

	/* Network routine start */

	StartServer()

	log.Notice("Akadok has started")

	// Keeps the server idle ATM
	for {
		time.Sleep(time.Millisecond * 50)
	}
}
