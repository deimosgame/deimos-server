package main

import (
	"bitbucket.org/deimosgame/go-akadok/entity"
	"bitbucket.org/deimosgame/go-akadok/packet"
	"errors"
	"net"
)

// SetupHandlers contains the handlers for each packet ID
func SetupHandlers() {
	Handle(0x00, Handshake)
}

func HandlePacket(handler interface{}, addr *net.UDPAddr, p *packet.Packet) {
	h := PacketHandler{addr}
	// Magic happens
	(handler.(func(*PacketHandler, *packet.Packet)))(&h, p)
}

// PacketHandler type for easy network communications in handlers
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

// GetPlayer allows a handler to easily get a player from its address
func (h *PacketHandler) GetPlayer() (*entity.Player, error) {
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
func Handshake(h *PacketHandler, p *packet.Packet) {
	outPacket := packet.NewPacket(0)
	message := []byte("Hello world")
	outPacket.AddField(&message)
	h.Answer(outPacket)
}
