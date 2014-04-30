package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"net"
	"strconv"
)

var (
	Handlers = make(map[byte]interface{})
)

type OutboundMessage struct {
	Address *net.UDPAddr
	Packet  *packet.Packet
}

// StartServer starts the UDP server in a different goroutine
func StartServer() {
	networkInput = make(chan *OutboundMessage)
	go Server(networkInput)
}

// Server is the main function of the server, which mainly handles outbound data
func Server(input chan *OutboundMessage) {
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
		case m := <-input:
			encodedPackets := m.Packet.Encode()
			for _, currentPacket := range *encodedPackets {
				conn.WriteToUDP(*currentPacket, m.Address)
			}
		}
	}
}

// HandleClient manages incoming packets and dispatches them to their respective
// handlers
func HandleClient(conn *net.UDPConn) {
	for {
		var buf [packet.PacketSize]byte

		n, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Warn("Had trouble receiving an UDP packet!")
		}

		packetData := buf[:n]
		p, err := packet.ReadPacket(&packetData)

		if err != nil {
			log.Warn("Corrupted packet received!")
			continue
		}
		log.Debug(strconv.Itoa(int(p.Id)), string(p.Data))

		if p.IsSplitted() {
			// TODO: stack splitted packets
		}

		CheckHandler(addr, p)
	}
}

// Handler adds/edits a handler for a given packet type
// Handlers must have the following prototype:
// (h *PacketHandler, packet *packet.Packet)
func Handle(packetId byte, handler interface{}) {
	Handlers[packetId] = handler
}

// RemoveHandler deletes a handler from the handler table
func RemoveHandler(packetId byte) bool {
	if _, ok := Handlers[packetId]; !ok {
		return false
	}
	delete(Handlers, packetId)
	return true
}

// CheckHandler tries to use a handler for packets
func CheckHandler(origin *net.UDPAddr, p *packet.Packet) {
	if handler, ok := Handlers[p.Id]; ok {
		// Starts a new goroutine for the handler
		go HandlePacket(handler, origin, p)
	} else {
		log.Warn("An unknown packet has been received!")
	}
}

// SendMessage messages all players on the server
func SendMessage(message string) {
	messagePacket := packet.New(0x03)
	messageBytes := []byte(message)
	messagePacket.AddField(&messageBytes)
	for _, currentPlayer := range players {
		currentPlayer.Send(messagePacket)
	}
}
