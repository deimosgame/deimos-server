package packet

import (
	"bytes"
)

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

func (p *Packet) IsSplitted() bool {
	return p.Total > 1 || len(p.Data) > 572
}
