package packet

import (
	"bytes"
	"errors"
)

// ReadPacket arranges and reads multiple packets of the same type at once,
// resulting in one large packet
func ReadPacket(receivedPackets ...*[]byte) (*Packet, error) {
	decodedPackets := make([]*Packet, len(receivedPackets))
	// Reorder received packets
	for _, currentPacket := range receivedPackets {
		decodedPacket, err := ReadSinglePacket(currentPacket)
		if err != nil {
			return nil, err
		}
		decodedPackets[decodedPacket.Index] = decodedPacket
	}
	// Put all packets together
	dataBuffer := bytes.NewBuffer(nil)
	for _, currentPacket := range decodedPackets {
		dataBuffer.Write(currentPacket.Data)
	}
	decodedPackets[0].Data, decodedPackets[0].Total = dataBuffer.Bytes(), 1
	return decodedPackets[0], nil
}

// IsSplitted checks if a packet will need to be splitted
func (p *Packet) IsSplitted() bool {
	return p.Total > 1 || len(p.Data) > 572
}

// IsSplitted checks if a raw packet is splitted
func IsSplitted(p *[]byte) (byte, err) {
	if len(*p) < 4 {
		return 0, errors.New("Invalid packet")
	}
	return (*p)[3], nil
}
