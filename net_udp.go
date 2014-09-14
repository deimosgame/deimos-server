package main

import (
	"net"
	"strconv"

	"github.com/deimosgame/deimos-server/packet"
)

type UDPOutboundMessage struct {
	Address *Address
	Packet  *packet.Packet
}

// Server is the main function of the server, which mainly handles outbound data
func UDPServer() {
	service := ":" + strconv.Itoa(config.Port)
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		log.Panic("Error while resolving UDP address")
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panic("Couldn't bind server to UDP")
	}

	// Starts the handler for inbound packets
	go UDPHandleClient(conn)
	for {
		select {
		case m := <-UdpNetworkInput:
			for _, currentPacket := range m.Packet.Encode() {
				conn.WriteToUDP(currentPacket, m.Address.UDPAddr)
			}
		}
	}
}

// HandleClient manages incoming packets and dispatches them to their respective
// handlers
func UDPHandleClient(conn *net.UDPConn) {
	for {
		var buf [packet.PacketSize]byte

		n, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Debug("Had trouble receiving an UDP packet!", err.Error())
			continue
		}

		packetData := buf[:n]
		p, err := packet.ReadPacket(packetData)

		if err != nil {
			log.Warn("Corrupted UDP packet received!")
			continue
		}
		p.Type = packet.PacketTypeUDP
		log.Debug(strconv.Itoa(int(p.Id)), string(p.Data))

		if p.IsSplitted() {
			// TODO: stack splitted packets (maybe someday)
		}

		UsePacketHandler(&Address{
			UDPAddr: addr,
		}, p, nil)
	}
}
