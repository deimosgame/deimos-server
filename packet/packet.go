package packet

import (
	"errors"
)

type Packet struct {
	Id   byte
	Data []byte
}

func ReadPacket(packetBuffer []byte) (Packet, error) {
	if len(packetBuffer) < 3 {
		return Packet{}, errors.New("Invalid packet!")
	}

	// Checksum
	packetChecksum := packetBuffer[0]
	if packetChecksum != byte(len(packetBuffer)%256) {
		return Packet{}, errors.New("Invalid checksum (corrupted packet?)")
	}

	packet := Packet{
		Id:   packetBuffer[1],
		Data: packetBuffer[2:],
	}
	return packet, nil
}

// TODO: Create a new packet, add fields, send it...
