package packet

import (
	"bytes"
	"errors"
)

// ReadPacket arranges and reads multiple packets of the same type at once,
// resulting in one large packet
func ReadPacket(receivedPackets ...[]byte) (*Packet, error) {
	packetCount := len(receivedPackets)

	if packetCount > 255 {
		return nil, errors.New("Too many packets!")
	}

	decodedPackets := make([]*Packet, packetCount)
	// Reorder received packets
	for _, currentPacket := range receivedPackets {
		decodedPacket, err := ReadSinglePacket(currentPacket)
		if err != nil {
			return nil, err
		}
		if decodedPacket.Index >= byte(packetCount) {
			return nil, errors.New("Invalid packet index")
		}
		decodedPackets[decodedPacket.Index] = decodedPacket
	}

	// Put all packets back together
	dataBuffer := bytes.NewBuffer(nil)
	for _, currentPacket := range decodedPackets {
		dataBuffer.Write(currentPacket.Data)
	}
	decodedPackets[0].Data, decodedPackets[0].Total = dataBuffer.Bytes(), 1
	return decodedPackets[0], nil
}

// IsSplitted checks if a packet will need to be splitted
func (p *Packet) IsSplitted() bool {
	return p.Total > 1 || len(p.Data) > PacketSize-4
}

// IsSplitted checks if a raw packet is splitted
func IsSplitted(p []byte) (byte, error) {
	if len(p) < 4 {
		return 0, errors.New("Invalid packet")
	}
	return (p)[3], nil
}
