package main

import (
	"bitbucket.org/deimosgame/go-akadok/util"
	"net"
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
	config           *AkadokConfig
	log              *util.Logger
	configFile       = "server.cfg"
	masterServerLost = false

	networkInput chan *OutboundMessage

	// Game-related variables
	currentMap string
	players    map[*net.UDPAddr]*Player
)

func main() {
	// temp inititalization
	currentMap = "coolmap"

	/* Config loading */

	LoadConfig()

	/* Logging engine */

	InitLogging()

	log.Notice("Akadok is loading...")

	/* Commands parsing */

	ParseCommands()

	/* Server IP resolving */

	ResolveIP()

	/* Heartbeat scheduling */

	Heartbeat()

	/* Network routine start */

	StartServer()

	/* Setup handlers for incoming packets */

	SetupHandlers()

	/* Tadaaaa */

	log.Notice("Akadok has started")

	// Keeps the server idle
	for {
		select {}

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
