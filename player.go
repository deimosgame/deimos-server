package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"net"
)

type Player struct {
	Name          string
	AkadokAccount string
	Address       *net.UDPAddr

	// Position
	X float32
	Y float32
	Z float32
	// Rotation
	BodyRotation float32
	HeadRotation float32
	// Velocity
	XVelocity           float32
	YVelocity           float32
	ZVelocity           float32
	BodyAngularVelocity float32
	HeadAngularVelocity float32

	// Misc values
	ModelId        string
	SelectedWeapon byte
}

// MatchName checks if a player name begins with a specific expression
func (p *Player) Match(name string) bool {
	return p.Name[:len(name)] == name || name == "*"
}

func (p *Player) Kick(reason string) {
	if reason == "" {
		reason = "Kicked!"
	}
	kickPacket := packet.New(0x02)
	reasonBytes := []byte(reason)
	kickPacket.AddField(&reasonBytes)
	kickMessage := OutboundMessage{
		Address: p.Address,
		Packet:  kickPacket,
	}
	networkInput <- &kickMessage
	delete(players, p.Address)
}
