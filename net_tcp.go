package main

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"

	"github.com/deimosgame/deimos-server/packet"
)

// TCPServer is designed to be ran as a goroutine to listen for incoming new TCP
// connections
func TCPServer() {
	service := ":" + strconv.Itoa(config.Port)
	l, err := net.Listen("tcp", service)
	if err != nil {
		log.Panic("Error when starting TCP server:", err.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error while accepting a TCP connection:", err.Error())
			continue
		}
		go TCPHandleClient(&conn)
	}
}

// TCPHandleClient is a goroutine individually launched for every new client
// by the TCP listener
func TCPHandleClient(conn *net.Conn) {
	defer (*conn).Close()

	// Making of a new initialized player object
	player := &Player{
		Address: &Address{
			TCPAddr: (*conn).RemoteAddr().(*net.TCPAddr),
		},
		Initialized:     false,
		TCPNetworkInput: make(chan *packet.Packet, NetworkChannelSize),
	}

	err := TCPPreHandling(conn, player)
	if err != nil {
		log.Warn("Wrong handshake packet received!", err.Error())
		return
	}

	// Player addition
	i := byte(0)
	for ; i <= 255; i++ {
		if _, ok := players[i]; !ok {
			break
		}
	}
	players[i] = player
	go TCPHandleClientSend(conn, player)

	for {
		// Receive packets
		p, stop, err := TCPReadPacket(conn)
		if stop {
			log.Debug(err.Error())
			return
		}
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		addr := (*conn).RemoteAddr().(*net.TCPAddr)
		UsePacketHandler(&Address{
			TCPAddr: addr,
		}, p, player)
	}
}

// TCPHandleClientSend reads the player's TCPNetworkInput channel and sends
// packets over the network
func TCPHandleClientSend(conn *net.Conn, player *Player) {
	for {
		m, ok := <-player.TCPNetworkInput
		if !ok {
			// Closed channel
			return
		}

		for _, currentPacket := range m.Encode() {
			(*conn).Write(currentPacket)
		}
	}
}

// TCPPreHandling manages player connections through the handshake
func TCPPreHandling(conn *net.Conn, player *Player) error {
	p, _, err := TCPReadPacket(conn)
	if err != nil {
		return err
	}
	if p.Id != 0x00 {
		return errors.New("Unexpected packet received")
	}

	// Response packet forging
	outPacket, err := packet.New(packet.PacketTypeTCP, 0), nil
	if version, err := p.GetField(0, 1); err != nil ||
		version[0] != ProtocolVersion {
		outPacket.AddFieldBytes(0)
		player.Send(outPacket)
		return errors.New("Incorrect client version")
	}

	// UDP port parsing
	udpPortBytes, err := p.GetField(1, 4)
	if err != nil {
		outPacket.AddFieldBytes(0)
		player.Send(outPacket)
		return errors.New("Error while parsing the client UDP port")
	}

	udpPort := binary.LittleEndian.Uint32(udpPortBytes)
	player.Address.UDPAddr = &net.UDPAddr{
		IP:   player.Address.TCPAddr.IP,
		Port: int(udpPort),
		Zone: player.Address.TCPAddr.Zone,
	}

	outPacket.AddFieldBytes(ProtocolVersion)
	player.Send(outPacket)
	return nil
}

// TCPReadPacket tries to read a packet from a TCP connection
func TCPReadPacket(conn *net.Conn) (*packet.Packet, bool, error) {
	buf := make([]byte, packet.PacketSize)
	n, err := (*conn).Read(buf)
	if err != nil {
		return nil, true, errors.New("Had trouble when receiving a TCP packet!")
	}

	packetData := buf[:n]
	p, err := packet.ReadPacket(packetData)

	if err != nil {
		return nil, false, errors.New("Corrupted TCP packet received")
	}
	p.Type = packet.PacketTypeTCP
	log.Debug(strconv.Itoa(int(p.Id)), string(p.Data))

	return p, false, nil
}
