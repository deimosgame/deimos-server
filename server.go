package main

import (
	"os"
	"time"

	"github.com/deimosgame/deimos-server/util"
)

const (
	ProtocolVersion    = byte(1)
	APIServer          = "https://deimos-ga.me/api"
	MasterServer       = "https://akadok.deimos-ga.me"
	HeartbeatInterval  = 15 * time.Second
	BroadcastInterval  = 20 * time.Millisecond
	NetworkChannelSize = 10
)

var (
	config     *DeimosConfig
	log        *util.Logger
	configFile = "server.cfg"

	apiServerLost     = false
	masterServerLost  = false
	serverKeepupAlert = false
	insecureAlert     = false

	UdpNetworkInput = make(chan *UDPOutboundMessage, NetworkChannelSize)
	APIInput        = make(chan *APIRequest, NetworkChannelSize)
	tickRateSecs    float32

	// Game-related variables
	currentMap      string
	worldSnapshotId uint32 = 0
	worldSnapshots         = make(map[uint32]*World)
	players                = make(map[byte]*Player)
	entities               = make(map[*Entity]bool)
)

func main() {
	/* Config loading */

	LoadConfig()
	currentMap = config.Maps[0]

	/* Logging engine */

	InitLogging()

	log.Notice("Deimos server is loading...")
	//log.DebugMode = true

	/* Server IP resolving */

	ResolveIP()

	/* Setup handlers for incoming packets and commands */

	SetupPacketHandlers()
	SetupCommandHandlers()

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

	go UDPServer()
	go TCPServer()
	go APIProcess()

	/* Tadaaaa */

	log.Notice("Deimos server has started")

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
