package packet

import (
	"bytes"
	"errors"
)

type Packet struct {
	Id   byte
	Data []byte
}

// NewPacket creates an empty packet with its id
func NewPacket(id byte) *Packet {
	return &Packet{
		Id: id,
	}
}

// ReadPacket reads a byte array contents and tries to parse it as a packet
func ReadPacket(packetBuffer []byte) (*Packet, error) {
	if len(packetBuffer) < 3 {
		return &Packet{}, errors.New("Invalid packet!")
	}

	// Checksum
	packetChecksum := packetBuffer[0]
	if packetChecksum != byte(len(packetBuffer)%256) {
		return &Packet{}, errors.New("Invalid checksum (corrupted packet?)")
	}

	packet := &Packet{
		Id:   packetBuffer[1],
		Data: packetBuffer[2:],
	}
	return packet, nil
}

// AddField adds a new field at the end of the current packet
func (p *Packet) AddField(fieldData []byte) error {
	packetBuffer := bytes.NewBuffer(p.Data)
	if p.Data != nil {
		if err := packetBuffer.WriteByte(0); err != nil {
			return err
		}
	}
	if _, err := packetBuffer.Write(fieldData); err != nil {
		return err
	}
	p.Data = packetBuffer.Bytes()
	return nil
}

// AddFieldBytes is an alias for AddField, except it doesn't needs an array
func (p *Packet) AddFieldBytes(b ...byte) {
	p.AddField(b)
}

// GetField returns the value of a field using its index
func (p *Packet) GetField(n int) ([]byte, error) {
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
	return p.Data[start:j], nil
}

// Encode encodes the packets to a byte array in order to send it on the network
func (p *Packet) Encode() *[]byte {
	// Checksum
	checksum := byte((len(p.Data) + 2) % 256)
	// Buffer
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(checksum)
	buf.WriteByte(p.Id)
	buf.Write(p.Data)
	result := buf.Bytes()
	return &result
}
