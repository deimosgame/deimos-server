package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"fmt"
	"net"
	"strconv"
)

var (
	Handlers = make(map[byte]interface{})
)

type UDPOutboundMessage struct {
	Address *net.UDPAddr
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
	go HandleClient(conn)
	for {
		select {
		case m := <-networkInput:
			encodedPackets := m.Packet.Encode()
			for _, currentPacket := range *encodedPackets {
				conn.WriteToUDP(*currentPacket, m.Address)
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
			fmt.Println(err)
			log.Warn("Had trouble receiving an UDP packet!")
			continue
		}

		packetData := buf[:n]
		p, err := packet.ReadPacket(&packetData)

		if err != nil {
			log.Warn("Corrupted packet received!")
			continue
		}
		log.Debug(strconv.Itoa(int(p.Id)), string(p.Data))

		if p.IsSplitted() {
			// TODO: stack splitted packets (maybe one day)
		}

		UsePacketHandler(addr, p)
	}
}
