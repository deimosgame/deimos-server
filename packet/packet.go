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
	computedChecksum := byte(0)
	for i := 1; i < len(*packetBuffer); i++ {
		currentByte := (*packetBuffer)[i]
		for j := 0; currentByte > 0; j++ {
			computedChecksum += byte(j%2+1) * (currentByte % 2)
			currentByte = currentByte >> 1
		}
	}
	if (*packetBuffer)[0] != computedChecksum {
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

func (p *Packet) GetField(index, length int) (*[]byte, error) {
	if index+length > len(p.Data) {
		return &[]byte{}, errors.New("Data too long")
	}
	data := p.Data[index : index+length]
	return &data, nil
}

// GetField returns the value of a field using its index
func (p *Packet) GetFieldString(index int) (*string, error) {
	if index > len(p.Data) {
		s := ""
		return &s, errors.New("Unknown index")
	}
	j := index
	for ; j < len(p.Data); j++ {
		if p.Data[j] == 0 {
			break
		}
	}
	field := string(p.Data[index:j])
	return &field, nil
}

// Encode encodes the packets to a byte array in order to send it on the network
func (p *Packet) Encode() *[]*[]byte {
	// Remove the last \00 if necessary
	i := 1
	for ; i <= len(p.Data) && p.Data[len(p.Data)-i] == 0; i++ {
	}
	p.Data = p.Data[:len(p.Data)-i+1]

	if len(p.Data) <= PacketSize-4 {
		// First buffer
		buf := bytes.NewBuffer(nil)
		buf.WriteByte(p.Id)
		buf.WriteByte(p.Index)
		buf.WriteByte(p.Total)
		buf.Write(p.Data)

		// Checksum
		checksum := byte(0)
		for _, currentByte := range buf.Bytes() {
			for j := 0; currentByte > 0; j++ {
				checksum += byte(j%2+1) * (currentByte % 2)
				currentByte = currentByte >> 1
			}
		}

		// Final buffer
		finalBuf := bytes.NewBuffer(nil)
		finalBuf.WriteByte(checksum)
		finalBuf.Write(buf.Bytes())
		result := finalBuf.Bytes()
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
