package main

import (
	"errors"
	"net"

	"bitbucket.org/deimosgame/go-akadok/packet"
)

// Generic type for TCP and UDP packets
type Address struct {
	TCPAddr *net.TCPAddr
	UDPAddr *net.UDPAddr
}

// Compare enables easy comparison of two Address structures
func (a1 *Address) Compare(a2 *Address) bool {
	return (a2.TCPAddr != nil && a1.TCPAddr != nil &&
		(*a2.TCPAddr).String() == (*a1.TCPAddr).String()) ||
		(a1.UDPAddr != nil && a2.UDPAddr != nil &&
			a1.UDPAddr.String() == a2.UDPAddr.String())
}

// Send tries its best to send a packet to somebody
// don't hurt it if it fails :((
func (a *Address) Send(p *packet.Packet, player *Player) error {
	if player == nil && a.TCPAddr == nil && a.UDPAddr == nil {
		return errors.New("Cannot match anything with no addresses")
	}

	defer func() {
		recover()
	}()

	var err error
	if p.Type == packet.PacketTypeUDP {
		if player == nil && a.UDPAddr == nil {
			// Retreive player by its address
			player, err = MatchByTCPAddress(a.TCPAddr)
			if err != nil {
				return err
			}
		}
		UdpNetworkInput <- &UDPOutboundMessage{
			Address: player.Address,
			Packet:  p,
		}
		return nil
	} else if p.Type == packet.PacketTypeTCP {
		if player == nil {
			if a.TCPAddr != nil {
				// Match player by its TCP address, simply
				player, err = MatchByTCPAddress(a.TCPAddr)

			} else {
				// Match player by its UDP address
				player, err = MatchByUDPAddress(a.UDPAddr)
			}
			if err != nil {
				return err
			}
		}
		player.TCPNetworkInput <- p
	}
	return errors.New("Unknown packet type")
}
