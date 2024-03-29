package packet

import (
	"bytes"
	"testing"
)

func TestNewPacket(t *testing.T) {
	packet := New(PacketTypeTCP, 0)
	if packet.Id != 0 || !bytes.Equal(packet.Data, []byte{}) {
		t.Fail()
	}
}

func TestReadPacket(t *testing.T) {
	// Valid packet
	if _, err := ReadPacket([]byte{12, 0, 0, 0, 1, 2, 3, 0, 4, 5, 6}); err != nil {
		t.Fail()
	}

	// Valid packet #2
	if _, err := ReadPacket([]byte{2, 0, 0, 1, 1}); err != nil {
		t.Fail()
	}

	// Corrupted checksum
	if _, err := ReadPacket([]byte{5, 0, 0, 0, 1, 2, 3, 0, 4, 5, 6}); err == nil {
		t.Log("ReadPacket missed checksum check")
		t.Fail()
	}

	// Wrong size
	if _, err := ReadPacket([]byte{1, 9}); err == nil {
		t.Log("ReadPacket missed size check")
		t.Fail()
	}
}

func TestAddField(t *testing.T) {
	packet := New(PacketTypeTCP, 0)
	// Add a few bytes
	b1, b2 := []byte{1, 2, 3}, []byte{4, 5, 6}
	packet.AddField(b1)
	packet.AddField(b2)

	// Check read bytes
	if d1, err := packet.GetField(0, 3); err != nil || !bytes.Equal(d1, b1) {
		t.Fail()
	}
	if d2, err := packet.GetField(3, 3); err != nil || !bytes.Equal(d2, b2) {
		t.Fail()
	}
}

func TestAddFieldBytes(t *testing.T) {
	packet := New(PacketTypeTCP, 0)
	// Add a few bytes
	b1, b2 := []byte{1, 2, 3}, []byte{4, 5, 6}
	packet.AddFieldBytes(1, 2, 3)
	packet.AddFieldBytes(4, 5, 6)

	// Check read bytes
	if d1, err := packet.GetField(0, 3); err != nil || !bytes.Equal(d1, b1) {
		t.Fail()
	}
	if d2, err := packet.GetField(3, 3); err != nil || !bytes.Equal(d2, b2) {
		t.Fail()
	}
}

func TestGetField(t *testing.T) {
	packet := New(PacketTypeTCP, 0)
	d := []byte{1, 2, 3}
	packet.AddField(d)

	// Check read bytes
	if r, err := packet.GetField(0, 3); err != nil || !bytes.Equal(r, d) {
		t.Fail()
	}

	// Unexisting field
	if _, err := packet.GetField(3, 3); err == nil {
		t.Log("Packet found an unexisting field in a packet")
		t.Fail()
	}
}

func TestEncode(t *testing.T) {
	expectedPacketContents := []byte{12, 0, 0, 0, 1, 2, 3, 4, 5, 6}

	packet := New(PacketTypeTCP, 0)
	packet.AddFieldBytes(1, 2, 3)
	packet.AddFieldBytes(4, 5, 6)
	result := packet.Encode()

	if !bytes.Equal(result[0], expectedPacketContents) {
		t.Fail()
	}
}
