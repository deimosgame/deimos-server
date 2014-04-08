package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"net"
	"strconv"
	"time"
)

func startServer() {
	go server()
}

func server() {
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
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {

	var buf [512]byte

	n, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		log.Warn("Had trouble receiving an UDP packet!")
	}

	packet, err := packet.ReadPacket(buf[:n])
	if err != nil {
		log.Warn("Corrupted packet received!")
	}
	log.Debug(strconv.Itoa(int(packet.Id)), string(packet.Data))

	daytime := time.Now().String()

	conn.WriteToUDP([]byte(daytime), addr)
}
