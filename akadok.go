package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"os"
	"time"
)

const (
	ProtocolVersion   = byte(1)
	MasterServer      = "https://akadok.deimos-ga.me"
	HeartbeatInterval = 15 * time.Second
	BroadcastInterval = 20 * time.Millisecond
)

var (
	config     *AkadokConfig
	log        *util.Logger
	configFile = "server.cfg"

	masterServerLost  = false
	serverKeepupAlert = false
	insecureAlert     = false

	networkInput = make(chan *OutboundMessage)

	// Game-related variables
	currentMap string
	players    = make(map[string]*Player)
)

func main() {
	// temp inititalization
	currentMap = "coolmap"

	/* Config loading */

	LoadConfig()

	/* Logging engine */

	InitLogging()

	log.Notice("Akadok is loading...")

	/* Server IP resolving */

	ResolveIP()

	/* Setup handlers for incoming packets */

	SetupHandlers()

	/**
	 *  Sub-routines permanently executed (in goroutines)
	 */

	/* Commands parsing routine */

	go CommandParser()

	/* Heartbeat scheduling */

	go Heartbeat()

	/* Start world simulation routine */

	go WorldSimulation()

	/* Network routine start */

	go Server()

	/* Tadaaaa */

	log.Notice("Akadok has started")

	// Keeps the main process idle
	for {
		time.Sleep(time.Millisecond * 50)
	}
}

// Stop stops the server gracefully
func Stop(reason string) {
	if reason == "" {
		log.Info("Stopping the server!")
		reason = "Server is stopping!"
	} else {
		log.Info("Stopping the server: " + reason)
	}
	for _, currentPlayer := range players {
		currentPlayer.Kick(reason)
	}
	os.Exit(0)
}
