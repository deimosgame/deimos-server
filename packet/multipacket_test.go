package packet

import (
	"bytes"
	"testing"
)

func TestReadPackets(t *testing.T) {
	// 578 bytes packet
	pkt, pktdata := New(5), []byte("Hello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHe")
	pkt.AddField(&pktdata)
	result, err := ReadPacket(*pkt.Encode()...)
	if err != nil || bytes.Compare(result.Data, pktdata) != 0 {
		t.Fail()
	}
}

func TestSplitted(t *testing.T) {
	pkt, pktdata := New(5), []byte("Hello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHello worldHe")
	pkt.AddField(&pktdata)
	if !pkt.IsSplitted() {
		t.Fail()
	}
}
