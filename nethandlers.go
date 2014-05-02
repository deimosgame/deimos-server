package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"encoding/binary"
	"errors"
	"net"
)

// SetupHandlers contains the handlers for each packet ID
func SetupHandlers() {
	Handle(0x00, HandleHandshakePacket)
	Handle(0x01, HandleClientConnectionPacket)
	Handle(0x02, HandleDisconnectionPacket)
	Handle(0x03, HandleChatPacket)
	Handle(0x04, HandleAcknowledgementPacket)
}

func HandlePacket(handler interface{}, addr *net.UDPAddr, p *packet.Packet) {
	h := PacketHandler{addr}
	// Magic happens
	(handler.(func(*PacketHandler, *packet.Packet)))(&h, p)
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
	player, ok := players[h.Address.String()]
	if !ok {
		return nil, errors.New("Unknown player")
	}
	return player, nil
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
	if _, ok := players[h.Address.String()]; ok {
		// Player is already connected (what a dumbass!)
		h.Error()
		return
	}
	// Retrive fields for the connection
	userId, err := p.GetFieldString(0)
	if err != nil {
		h.Error()
		return
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
	players[h.Address.String()] = &newPlayer
	// Send authorization packet
	outPacket.AddFieldBytes(1)
	outPacket.AddFieldString(&currentMap)
	h.Answer(outPacket)
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
