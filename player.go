package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Player struct {
	Name    string `prefix:"N"`
	Account string
	Address *net.UDPAddr

	// Gameplay values
	Score    byte `prefix:"L"`
	Instance string

	// Position
	X float32 `prefix:"X"`
	Y float32 `prefix:"Y"`
	Z float32 `prefix:"Z"`
	// Rotation
	XRotation float32 `prefix:"P"`
	YRotation float32 `prefix:"Q"`
	// Velocity
	XVelocity        float32 `prefix:"T"`
	YVelocity        float32 `prefix:"S"`
	ZVelocity        float32 `prefix:"T"`
	AngularVelocityX float32 `prefix:"U"`
	AngularVelocityY float32 `prefix:"V"`

	// Misc values
	ModelId       string `prefix:"M"`
	CurrentWeapon byte   `prefix:"W"`

	LastUpdate       time.Time
	LastAcknowledged *World
	Initialized      bool
}

// MatchName checks if a player name begins with a specific expression
func (p *Player) Match(name string) bool {
	return strings.ToLower(p.Name[:len(name)]) == strings.ToLower(name) ||
		name == "*"
}

// Send send multiple packets to a player
func (p *Player) Send(packets ...*packet.Packet) {
	for _, pkt := range packets {
		message := OutboundMessage{
			Address: p.Address,
			Packet:  pkt,
		}
		networkInput <- &message
	}
}

// NextTick updates a player for the next tick (for prediction purposes)
func (p *Player) NextTick() {
	if time.Since(p.LastUpdate) < time.Millisecond*15 {
		return
	}

	p.X = p.X + p.XVelocity*tickRateSecs
	p.Y = p.Y + p.YVelocity*tickRateSecs
	p.Z = p.Z + p.ZVelocity*tickRateSecs

	p.XRotation = p.XRotation + p.AngularVelocityX*tickRateSecs
	p.YRotation = p.YRotation + p.AngularVelocityY*tickRateSecs

	p.LastUpdate = time.Now()
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
	for i, player := range players {
		if player.Address.String() == p.Address.String() {
			delete(players, i)
			return
		}
	}
}
