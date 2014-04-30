package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

type Player struct {
	Name    string
	Account string
	Address *net.UDPAddr

	// Position
	X float32
	Y float32
	Z float32
	// Rotation
	BodyRotation float32
	HeadRotation float32
	// Velocity
	XVelocity           float32
	YVelocity           float32
	ZVelocity           float32
	BodyAngularVelocity float32
	HeadAngularVelocity float32

	// Misc values
	ModelId        string
	SelectedWeapon byte
}

// MatchName checks if a player name begins with a specific expression
func (p *Player) Match(name string) bool {
	return p.Name[:len(name)] == name || name == "*"
}

func (p *Player) Send(pkt *packet.Packet) {
	message := OutboundMessage{
		Address: p.Address,
		Packet:  pkt,
	}
	networkInput <- &message
}

// Kick kicks a player out of the server
func (p *Player) Kick(reason string) {
	if reason == "" {
		reason = "Kicked!"
	}
	kickPacket := packet.New(0x02)
	reasonBytes := []byte(reason)
	kickPacket.AddField(&reasonBytes)
	p.Send(kickPacket)
	SendMessage(p.Name + " has been kicked!")
	p.Remove()
}

// RefreshName gets the player name from the web
func (p *Player) RefreshName() error {
	apiUrl := "https://deimos-ga.me/api/get-name/"
	resp, err := http.Get(apiUrl + p.Account)
	if err != nil {
		return err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	type ApiResponse struct {
		Success bool
		Name    string
	}
	var responseStruct ApiResponse
	err = json.Unmarshal(responseData, &responseStruct)
	if err != nil {
		return err
	}
	p.Name = responseStruct.Name
	return nil
}

// Remove remove a player form the server
func (p *Player) Remove() {
	delete(players, p.Address.String())
}
