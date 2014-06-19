package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
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
	RegisterPacketHandler(0x08, HandleSoundPacket)
}

func HandlePacket(handler interface{}, addr *net.UDPAddr, p *packet.Packet) {
	h := PacketHandler{addr}
	// Magic happens
	(handler.(func(*PacketHandler, *packet.Packet)))(&h, p)
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
func UsePacketHandler(origin *net.UDPAddr, p *packet.Packet) {
	if handler, ok := Handlers[p.Id]; ok {
		// Starts a new goroutine for the handler
		go HandlePacket(handler, origin, p)
	} else {
		log.Warn("An unknown packet has been received!")
	}
}

/**
 *  PacketHandler type for easy network communications in handlers
 */

type PacketHandler struct {
	Address *net.UDPAddr
}

// Answer adds a packet to send to the network queue
func (h *PacketHandler) Answer(p *packet.Packet) {
	networkInput <- &OutboundMessage{
		Address: h.Address,
		Packet:  p,
	}
}

// Error returns an error packet to the client
func (h *PacketHandler) Error() {
	errorPacket := packet.New(0)
	errorPacket.AddFieldBytes(0)
	h.Answer(errorPacket)
}

// GetPlayer allows a handler to easily get a player from its address
func (h *PacketHandler) GetPlayer() (*Player, error) {
	for _, player := range players {
		if player.Address.String() == h.Address.String() {
			return player, nil
		}
	}
	return nil, errors.New("Unknown player")
}

/**
 *  Various packet handlers (convention: HandleXPacket)
 */

// HandleHandshakePacket (0x00)
func HandleHandshakePacket(h *PacketHandler, p *packet.Packet) {
	outPacket := packet.New(0)
	if version, err := p.GetField(0, 1); err != nil ||
		(*version)[0] != ProtocolVersion {
		outPacket.AddFieldBytes(0)
	} else {
		outPacket.AddFieldBytes(ProtocolVersion)
	}
	h.Answer(outPacket)
}

// HandleClientConnectionPacket (0x01). Allows a player to connect if
// everything is alright
func HandleClientConnectionPacket(h *PacketHandler, p *packet.Packet) {
	if _, err := h.GetPlayer(); err == nil {
		// Player is already connected
		h.Error()
		return
	}

	// Retrive fields for the connection
	userId, err := p.GetFieldString(0)
	if err != nil {
		h.Error()
		return
	}
	// Check if the account is not already used
	for _, player := range players {
		if player.Account == *userId {
			h.Error()
			return
		}
	}
	token, err := p.GetFieldString(len(*userId) + 1)
	if err != nil {
		h.Error()
		return
	}

	// Check the credentials of the user
	validToken, err := CheckToken(*userId, *token)
	if err != nil {
		h.Error()
		return
	}
	outPacket := packet.New(0x01)
	if !validToken {
		// 0 for denied connection
		outPacket.AddFieldBytes(0)
		h.Answer(outPacket)
		return
	}

	// Create a player
	newPlayer := Player{
		Account:          string(*userId),
		Address:          h.Address,
		LastAcknowledged: &World{},
		Initialized:      true,
	}
	newPlayer.RefreshName()

	i := byte(0)
	for ; i <= 255; i++ {
		if _, ok := players[i]; !ok {
			break
		}
	}
	players[i] = &newPlayer

	// Send authorization packet
	outPacket.AddFieldBytes(1)
	outPacket.AddFieldString(&currentMap)
	h.Answer(outPacket)

	UpdatePlayerList()

	log.Info(newPlayer.Name + " (" + newPlayer.Account + " - " +
		h.Address.IP.String() + ") has joined the game!")
	SendMessage(newPlayer.Name + " has joined the game!")
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
	log.Info(player.Name + " has left the server.")
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
	log.Info("<" + player.Name + "> " + *message)
	SendMessage("<" + player.Name + "> " + *message)
}

// HandleAcknowledgementPacket (0x04) handles world acknowledgement packets from
// the client
func HandleAcknowledgementPacket(h *PacketHandler, p *packet.Packet) {
	player, err := h.GetPlayer()
	if err != nil {
		h.Error()
		return
	}
	idBytes, err := p.GetField(0, 4)
	if err != nil || len(*idBytes) != 4 {
		h.Error()
		return
	}
	id := binary.LittleEndian.Uint32(*idBytes)
	if snapshot, ok := worldSnapshots[id]; ok {
		player.LastAcknowledged = snapshot
	} else {
		h.Error()
		return
	}
}

// HandleMovementPacket (0x05) changes the position of the player
func HandleMovementPacket(h *PacketHandler, p *packet.Packet) {
	player, err := h.GetPlayer()
	if err != nil || len(p.Data) != 40 {
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
		buf := bytes.NewReader(*field)
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
	player.CurrentWeapon = (*weapon)[0]
	player.ModelId = (*model)[0]

	if player.LifeState != (*lifeState)[0] {
		player.LifeState = (*lifeState)[0]
		if (*lifeState)[0] == 0 {
			log.Infof("%s died", player.Name)
		}
	}

	if (*refreshByte)[0] != 0 {
		player.RefreshName()
	}
}

// HandleSoundPacket handles the sound packets emitted from players
func HandleSoundPacket(h *PacketHandler, p *packet.Packet) {

}
