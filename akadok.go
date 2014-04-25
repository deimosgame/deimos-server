package main

import (
	"bitbucket.org/deimosgame/go-akadok/entity"
	"bitbucket.org/deimosgame/go-akadok/util"
	"net"
	"time"
)

const (
	MasterServer      = "https://akadok.deimos-ga.me"
	HeartbeatInterval = 15 * time.Second
	BroadcastInterval = 20 * time.Millisecond
)

var (
	config           *AkadokConfig
	log              *util.Logger
	configFile       = "server.cfg"
	masterServerLost = false

	networkInput chan *World

	// Game-related variables
	currentMap string
	players    map[*net.UDPAddr]entity.Player
)

func main() {
	// temp inititalization
	currentMap = "coolmap"

	/* Config loading */

	LoadConfig()

	/* Logging Engine */

	InitLogging()

	log.Notice("Akadok is loading...")

	/* Server IP Resolving */

	ResolveIP()

	/* Heartbeat scheduling */

	Heartbeat()

	/* Network routine start */

	StartServer()

	log.Notice("Akadok has started")

	// Keeps the server idle ATM
	for {
		select {}

		time.Sleep(time.Millisecond * 50)
	}
}
