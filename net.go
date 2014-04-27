package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"net"
	"strconv"
)

func StartServer() {
	networkInput := make(chan *World)
	go Server(networkInput)
}

func Server(input chan *World) {
	service := ":" + strconv.Itoa(config.Port)
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		log.Panic("Error while resolving UDP address")
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panic("Couldn't bind server to UDP")
	}

	for {
		select {
		case <-input:

		default:
			HandleClient(conn)
		}
	}
}

func HandleClient(conn *net.UDPConn) {

	var buf [packet.PacketSize]byte

	n, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		log.Warn("Had trouble receiving an UDP packet!")
	}

	packetData := buf[:n]
	packet, err := packet.ReadPacket(&packetData)

	if err != nil {
		log.Warn("Corrupted packet received!")
	}
	log.Debug(strconv.Itoa(int(packet.Id)), string(packet.Data))

	conn.WriteToUDP([]byte("Hello world"), addr)
}
