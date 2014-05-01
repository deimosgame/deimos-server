package packet

import (
	"bytes"
	"errors"
)

const (
	PacketSize = 576
)

type Packet struct {
	Id, Index, Total byte
	Data             []byte
}

// New creates an empty packet with its id
func New(id byte) *Packet {
	return &Packet{
		Id: id,
	}
}

// ReadPacket reads a byte array contents and tries to parse it as a packet
func ReadSinglePacket(packetBuffer *[]byte) (*Packet, error) {
	if len(*packetBuffer) < 3 {
		return &Packet{}, errors.New("Invalid packet!")
	}

	// Checksum
	packetChecksum := (*packetBuffer)[0]
	if packetChecksum != byte(len(*packetBuffer)%256) {
		return &Packet{}, errors.New("Invalid checksum (corrupted packet?)")
	}

	packet := &Packet{
		Id:    (*packetBuffer)[1],
		Index: (*packetBuffer)[2],
		Total: (*packetBuffer)[3],
		Data:  (*packetBuffer)[4:],
	}
	return packet, nil
}

// AddField adds a new field at the end of the current packet
func (p *Packet) AddField(fieldData *[]byte) error {
	packetBuffer := bytes.NewBuffer(p.Data)
	if _, err := packetBuffer.Write(*fieldData); err != nil {
		return err
	}
	p.Data = packetBuffer.Bytes()
	return nil
}

// AddFieldBytes is an alias for AddField, except it doesn't needs an array
func (p *Packet) AddFieldBytes(b ...byte) {
	p.AddField(&b)
}

// AddFieldString adds a string to a packet
func (p *Packet) AddFieldString(s *string) {
	data := append([]byte(*s), 0)
	p.AddField(&data)
}

// GetField returns the value of a field using its index
func (p *Packet) GetField(n int) (*[]byte, error) {
	// Begining of the field
	i := 0
	for j := 0; j < n; i++ {
		if i >= len(p.Data) {
			return nil, errors.New("Field not found")
		}
		if p.Data[i] == 0 {
			j++
		}
	}
	start := i
	// End of the field
	j := i
	for ; j < len(p.Data); j++ {
		if p.Data[j] == 0 {
			break
		}
	}
	field := p.Data[start:j]
	return &field, nil
}

// Encode encodes the packets to a byte array in order to send it on the network
func (p *Packet) Encode() *[]*[]byte {
	// Remove the last \00 if necessary
	if p.Data[len(p.Data)-1] == 0 {
		p.Data = p.Data[:len(p.Data)-1]
	}
	if len(p.Data) <= PacketSize-4 {
		// Checksum
		checksum := byte((len(p.Data) + 4) % 256)
		// Buffer
		buf := bytes.NewBuffer(nil)
		buf.WriteByte(checksum)
		buf.WriteByte(p.Id)
		buf.WriteByte(p.Index)
		buf.WriteByte(p.Total)
		buf.Write(p.Data)
		result := buf.Bytes()
		return &[]*[]byte{&result}
	}
	// Splitted packet
	packetCount := len(p.Data)/(PacketSize-4) + 1
	packets := make([]*[]byte, packetCount)
	for i := 0; i < packetCount; i++ {
		currentPacketLen := PacketSize - 4
		if len(p.Data)-i*(PacketSize-4) < PacketSize-4 {
			currentPacketLen = len(p.Data)
		}
		currentPacket := Packet{
			Id:    p.Id,
			Index: byte(i),
			Total: byte(packetCount),
			Data:  p.Data[i*(PacketSize-4) : currentPacketLen],
		}
		packets[i] = (*currentPacket.Encode())[0]
	}
	return &packets
}
