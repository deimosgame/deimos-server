package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"net"
)

func SetupHandlers() {
	Handle(0, Handshake)
}

func Handshake(addr *net.UDPAddr, p *packet.Packet) {
	outPacket := packet.NewPacket(0)
	outPacket.AddFieldBytes(1)
	networkInput <- &OutboundMessage{
		Address: addr,
		Packet:  outPacket,
	}
}
