package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"errors"
	"net"
)

// SetupHandlers contains the handlers for each packet ID
func SetupHandlers() {
	Handle(0x00, PacketHandshakeHandler)
	Handle(0x01, PacketClientConnectionHandler)
	Handle(0x03, PacketChatHandler)
}

func HandlePacket(handler interface{}, addr *net.UDPAddr, p *packet.Packet) {
	user := nil
	if player, ok := players[addr]; ok {
		user = player
	}
	h := PacketHandler{addr, user}
	// Magic happens
	(handler.(func(*PacketHandler, *packet.Packet)))(&h, p)
}

/**
 *  PacketHandler type for easy network communications in handlers
 */

type PacketHandler struct {
	Address *net.UDPAddr
	Player  *Player
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
	player, ok := players[h.Address]
	if !ok {
		return nil, errors.New("Unknown player")
	}
	return player, nil
}

/**
 *  Various packet handlers
 */

// Handshake (0x00)
func PacketHandshakeHandler(h *PacketHandler, p *packet.Packet) {
	outPacket := packet.New(0)
	if version, err := p.GetField(0); err != nil ||
		(*version)[0] != ProtocolVersion {
		outPacket.AddFieldBytes(0)
	} else {
		outPacket.AddFieldBytes(ProtocolVersion)
	}
	h.Answer(outPacket)
}

// ClientConnection (0x01). Allows a player to connect if everything is alright
func PacketClientConnectionHandler(h *PacketHandler, p *packet.Packet) {
	// Retrive fields for the connection
	userId, err := p.GetField(0)
	if err != nil {
		h.Error()
		return
	}
	token, err := p.GetField(1)
	if err != nil {
		h.Error()
		return
	}
	// Check the credentials of the user
	validToken, err := CheckToken(string(*userId), string(*token))
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
		Account: string(*userId),
		Address: h.Address,
	}
	newPlayer.RefreshName()
	players[h.Address] = &newPlayer
	// Send authorization packet
	outPacket.AddFieldBytes(1)
	currentMapBytes := []byte(currentMap)
	outPacket.AddField(&currentMapBytes)
	h.Answer(outPacket)
	log.Info(newPlayer.Name + "(" + newPlayer.Account + h.Address.IP.String() +
		") has joined the game!")
	SendMessage(newPlayer.Name + " has joined the game!")
}

func PacketChatHandler(h *PacketHandler, p *packet.Packet) {
	message, err := p.GetField(0)
	if err != nil {
		h.Error()
		return
	}
	SendMessage("<" + h.Player.Name + "> " + string(*message))
}
