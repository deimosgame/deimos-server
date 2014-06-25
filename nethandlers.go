package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/deimosgame/go-akadok/packet"
)

var (
	Handlers = make(map[byte]interface{})
)

// SetupHandlers contains the handlers for each packet ID
func SetupPacketHandlers() {
	RegisterPacketHandler(0x00, HandleHandshakePacket)
	RegisterPacketHandler(0x01, HandleClientConnectionPacket)
	RegisterPacketHandler(0x02, HandleDisconnectionPacket)
	RegisterPacketHandler(0x03, HandleChatPacket)
	RegisterPacketHandler(0x04, HandleAcknowledgementPacket)
	RegisterPacketHandler(0x05, HandleMovementPacket)
	RegisterPacketHandler(0x07, HandleInformationChangePacket)
	RegisterPacketHandler(0x0C, HandleDamagePacket)

	// Bouncing packets
	RegisterPacketHandler(0x08, HandleBounce(packet.PacketTypeUDP))
	RegisterPacketHandler(0x0B, HandleBounce(packet.PacketTypeTCP))
}

// HandlePacket bootstraps the handler goroutine with useful information such
// as the PacketHandler object and the packet in itself
func HandlePacket(handler interface{}, addr *Address, p *packet.Packet,
	player *Player) {
	h := &PacketHandler{Address: addr}
	if player == nil {
		if addr.TCPAddr != nil {
			player, _ = MatchByTCPAddress(addr.TCPAddr)
		} else if addr.UDPAddr != nil {
			player, _ = MatchByUDPAddress(addr.UDPAddr)
		}
	}
	h.Player = player
	if player == nil {
		h.Error()
		return
	}
	// Magic happens
	handler.(func(*PacketHandler, *packet.Packet))(h, p)
}

// RegisterPacketHandler adds/edits a handler for a given packet type
// Handlers must have the following prototype:
// (h *PacketHandler, packet *packet.Packet)
func RegisterPacketHandler(packetId byte, handler interface{}) {
	Handlers[packetId] = handler
}

// UnregisterPacketHandler deletes a handler from the handler table
func UnregisterPacketHandler(packetId byte) bool {
	if _, ok := Handlers[packetId]; !ok {
		return false
	}
	delete(Handlers, packetId)
	return true
}

// CheckHandler tries to use a handler for packets
func UsePacketHandler(origin *Address, p *packet.Packet, pl *Player) {
	if handler, ok := Handlers[p.Id]; ok {
		// Starts a new goroutine for the handler
		go HandlePacket(handler, origin, p, pl)
	} else {
		log.Warn("An unknown packet has been received!")
	}
}

/**
 *  PacketHandler type for easy network communications in handlers
 */

type PacketHandler struct {
	Address *Address
	Player  *Player
}

// Answer adds a packet to send to the network queue
func (h *PacketHandler) Answer(p *packet.Packet) {
	h.Address.Send(p, h.Player)
}

// Error returns an error packet to the client
func (h *PacketHandler) Error() {
	errorPacket := packet.New(packet.PacketTypeTCP, 0)
	errorPacket.AddFieldBytes(0)
	h.Answer(errorPacket)
}

// GetPlayer allows a handler to easily get a player from its address
func (h *PacketHandler) GetPlayer() (*Player, error) {
	for _, player := range players {
		if (h.Address.TCPAddr != nil && player.Address.TCPAddr != nil &&
			(*h.Address.TCPAddr).String() == (*player.Address.TCPAddr).String()) ||
			(h.Address.TCPAddr != nil && player.Address.TCPAddr != nil &&
				h.Address.TCPAddr.String() == player.Address.TCPAddr.String()) {
			return player, nil
		}
	}
	return nil, errors.New("Unknown player")
}

// HandleBounce manages bouncing packets to all players, except the sender
func HandleBounce(packetType byte) func(h *PacketHandler, p *packet.Packet) {
	return func(h *PacketHandler, p *packet.Packet) {
		p.Type = packetType
		for _, currentPlayer := range players {
			if currentPlayer.Equals(h.Player) {
				continue
			}
			currentPlayer.Send(p)
		}
	}
}

/**
 *  Various packet handlers (convention: HandleXPacket)
 */

// HandleHandshakePacket (0x00) is not represented anymore here since it is now
// hard-coded into net_tcp.go for connection initialization purposes.

// HandleHandshakePacket (0x00) just triggers an error if it is called, since
// the handler shouldn't deal with handshake packets
func HandleHandshakePacket(h *PacketHandler, p *packet.Packet) {
	h.Error()
	h.Player.Remove()
}

// HandleClientConnectionPacket (0x01). Allows a player to connect if
// everything is alright
func HandleClientConnectionPacket(h *PacketHandler, p *packet.Packet) {
	// Retrive fields for the connection
	userId, err := p.GetFieldString(0)
	if err != nil {
		h.Error()
		return
	}
	// Check if the account is not already used
	for _, player := range players {
		if player.Account == userId {
			h.Error()
			return
		}
	}
	token, err := p.GetFieldString(len(userId) + 1)
	if err != nil {
		h.Error()
		return
	}

	// Check the credentials of the user
	validToken, err := CheckToken(userId, token)
	if err != nil {
		h.Error()
		return
	}
	outPacket := packet.New(packet.PacketTypeTCP, 0x01)
	if !validToken {
		// 0 for denied connection
		outPacket.AddFieldBytes(0)
		h.Answer(outPacket)
		return
	}

	// Modify the player previously created during the handsake
	player := h.Player
	player.Account = userId
	player.LastAcknowledged = &World{}
	player.Initialized = true
	CheckUnlockedAchivements(player)
	player.RefreshName()

	// Achievement: log into a server
	UnlockAchievement(player, 1)

	// Send authorization packet
	outPacket.AddFieldBytes(1)
	outPacket.AddFieldString(currentMap)
	h.Answer(outPacket)

	UpdatePlayerList()

	log.Info(player.Name + " (" + player.Account + " - " +
		(*h.Address.TCPAddr).String() + ") has joined the game!")
	SendMessage(player.Name + " has joined the game!")
}

// HandleDisconnectionPacket (0x02) handles player disconnections
func HandleDisconnectionPacket(h *PacketHandler, p *packet.Packet) {
	// Just remove the player, the GC will do the rest
	player, err := h.GetPlayer()
	if err != nil {
		h.Error()
		return
	}
	player.Remove()
	SendMessage(player.Name + " has left the server.")
}

// HandleChatPacket (0x03) handles the chat packets
func HandleChatPacket(h *PacketHandler, p *packet.Packet) {
	player, err := h.GetPlayer()
	if err != nil {
		h.Error()
		return
	}
	message, err := p.GetFieldString(0)
	if err != nil {
		h.Error()
		return
	}

	if message[0] == '/' {
		// Handle commands
		if !h.Player.IsOperator() {
			h.Player.SendMessage("You are not allowed to run commands on the " +
				"server.")
			return
		}
		HandleCommand(message[1:], h.Player)
		return
	} else if message[0] == '@' {
		// Handle private messages
		messageSplit := strings.Split(message[1:], " ")
		if messageSplit[0] == "*" && !h.Player.IsOperator() {
			h.Player.SendMessage("You can't send a private message to " +
				"everybody.")
			return
		}
		matchedPlayers := MatchPlayers(messageSplit[0])
		if len(matchedPlayers) == 0 {
			h.Player.SendMessage("No player with this name has been found.")
			return
		}
		if len(matchedPlayers) > 1 && !h.Player.IsOperator() {
			// Only operators can PM multiple players at once
			h.Player.SendMessage("You can't send private messages to " +
				"multiple persons simultaneously.")
			return
		}
		messageText := message[len(messageSplit[0])+2:]
		for _, currentPlayer := range matchedPlayers {
			currentPlayer.SendMessage("<PM " + player.Name + "> " + messageText)
		}
		return
	}

	SendMessage("<" + player.Name + "> " + message)
}

// HandleAcknowledgementPacket (0x04) handles world acknowledgement packets from
// the client
func HandleAcknowledgementPacket(h *PacketHandler, p *packet.Packet) {
	idBytes, err := p.GetField(0, 4)
	if err != nil || len(idBytes) != 4 {
		fmt.Println("ack2")
		h.Error()
		return
	}
	id := binary.LittleEndian.Uint32(idBytes)
	if snapshot, ok := worldSnapshots[id]; ok {
		h.Player.LastAcknowledged = snapshot
	} else {
		h.Error()
		return
	}
}

// HandleMovementPacket (0x05) changes the position of the player
func HandleMovementPacket(h *PacketHandler, p *packet.Packet) {
	player := h.Player
	if len(p.Data) != 40 {
		h.Error()
		return
	}
	// Getting variables out of the packet
	for i := 0; i < 40; i += 4 {
		field, err := p.GetField(i, 4)
		if err != nil {
			h.Error()
			return
		}
		var element float32
		buf := bytes.NewReader(field)
		err = binary.Read(buf, binary.LittleEndian, &element)
		if err != nil {
			h.Error()
			return
		}
		switch i {
		case 0:
			player.X = element
		case 4:
			player.Y = element
		case 8:
			player.Z = element
		case 12:
			player.XRotation = element
		case 16:
			player.YRotation = element
		case 20:
			player.XVelocity = element
		case 24:
			player.YVelocity = element
		case 28:
			player.ZVelocity = element
		case 32:
			player.AngularVelocityX = element
		case 36:
			player.AngularVelocityY = element
		}
	}
	player.LastUpdate = time.Now()
}

// HandleInformationChangePacket (0x07) is a packet for small information
// changes that does not desserve to be present in a move event
func HandleInformationChangePacket(h *PacketHandler, p *packet.Packet) {
	player, err := h.GetPlayer()
	if err != nil || len(p.Data) != 4 {
		h.Error()
		return
	}
	// Read packet data
	weapon, err := p.GetField(0, 1)
	if err != nil {
		h.Error()
		return
	}
	model, err := p.GetField(1, 1)
	if err != nil {
		h.Error()
		return
	}
	refreshByte, err := p.GetField(2, 1)
	if err != nil {
		h.Error()
		return
	}
	lifeState, err := p.GetField(3, 1)
	if err != nil {
		h.Error()
		return
	}
	// Update player
	player.CurrentWeapon = weapon[0]
	player.ModelId = model[0]

	if player.CurrentWeapon == 5 {
		// Achievement: unlock the mystery weapon
		UnlockAchievement(player, 3)
	}

	if refreshByte[0] != 0 {
		player.RefreshName()
	}

	if player.LifeState == lifeState[0] {
		return
	}

	// Last damage decoding, mainly for achievements
	var damage int32
	var damageAuthorBytes []byte
	var damageAuthor *Player
	if player.LastDamagePacket != nil {
		damageBytes, err := player.LastDamagePacket.GetField(1, 4)
		if err != nil {
			h.Error()
			return
		}
		buf := bytes.NewReader(damageBytes)
		binary.Read(buf, binary.LittleEndian, &damage)
		damageAuthorBytes, err = player.LastDamagePacket.GetField(0, 1)
		if err != nil {
			h.Error()
			return
		}
		damageAuthor = players[damageAuthorBytes[0]]
	}

	// Happens when the player dies
	player.LifeState = lifeState[0]
	if lifeState[0] != 0 {
		return
	}

	if player.LastDamagePacket == nil {
		log.Infof("%s died.", player.Name)
		return
	}
	log.Infof("%s was killed by %s.", player.Name, damageAuthor.Name)

	if !damageAuthor.Equals(h.Player) && player.CurrentStreak > 5 {
		// Achievement: Instant cooling
		UnlockAchievement(damageAuthor, 12)
	}

	// Kill packet, for Manu
	killPacket := packet.New(packet.PacketTypeTCP, 0x0D)
	victimId := byte(0)
	for i, currentPlayer := range players {
		if currentPlayer.Equals(h.Player) {
			victimId = i
			break
		}
	}
	killPacket.AddFieldBytes(victimId)
	killPacket.AddFieldBytes(damageAuthorBytes[0])
	if damageAuthor.Equals(h.Player) {
		killPacket.AddFieldBytes(0xFF)
	} else {
		killPacket.AddFieldBytes(damageAuthor.CurrentWeapon)
	}
	for _, currentPlayer := range players {
		currentPlayer.Send(killPacket)
	}

	player.CurrentStreak = 0
	OnPlayerKill(h.Player, damageAuthor)
}

// HandleDamagePacket handles player damage, may it be from the player himself
// or from another player
func HandleDamagePacket(h *PacketHandler, p *packet.Packet) {
	hitPlayerBytes, err := p.GetField(0, 1)
	if err != nil {
		h.Error()
		return
	}
	hitPlayer, ok := players[hitPlayerBytes[0]]
	if !ok {
		h.Error()
		return
	}

	if hitPlayer.Equals(h.Player) {
		// Achievement: Self-Harm
		UnlockAchievement(h.Player, 7)
	}

	// Broadcast the damage packet to all players
	damagePacket := packet.New(packet.PacketTypeTCP, 0x0C)
	damageFieldBytes, err := p.GetField(1, 4)
	if err != nil {
		h.Error()
		return
	}
	damagePacket.AddField(damageFieldBytes)
	hitPlayer.LastDamagePacket = p
	hitPlayer.Send(damagePacket)
}
