package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/deimosgame/go-akadok/packet"
)

type Player struct {
	Name    string
	Account string
	Address *Address

	// Gameplay values
	LifeState byte `prefix:"A"`
	Score     byte `prefix:"L"`
	Instance  string

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
	ModelId       byte `prefix:"M"`
	CurrentWeapon byte `prefix:"W"`

	Victims          int
	CurrentStreak    int
	Achievements     []int
	LastDamagePacket *packet.Packet
	LastUpdate       time.Time
	LastAcknowledged *World
	TCPNetworkInput  chan *packet.Packet
	Initialized      bool
}

// MatchByUDPAddress tries to match an UDP address with the player using it
func MatchByUDPAddress(addr *net.UDPAddr) (*Player, error) {
	for _, player := range players {
		if player.Address.UDPAddr.String() == addr.String() {
			return player, nil
		}
	}
	return nil, errors.New("Player not found")
}

// MatchByTCPAddress tries to match a TCP address with the player using it
func MatchByTCPAddress(addr *net.TCPAddr) (*Player, error) {
	for _, player := range players {
		if (*player.Address.TCPAddr).String() == (*addr).String() {
			return player, nil
		}
	}
	return nil, errors.New("Player not found")
}

// MatchPlayer tries to find a player using his name
func MatchPlayers(name string) []*Player {
	playerList := make([]*Player, 0)
	for _, currentPlayer := range players {
		if currentPlayer.Match(name) {
			playerList = append(playerList, currentPlayer)
		}
	}
	return playerList
}

// Match checks if a player name begins with a specific expression
func (p *Player) Match(name string) bool {
	fmt.Println(p.Name, len(name))
	return strings.ToLower(p.Name[:len(name)]) == strings.ToLower(name) ||
		name == "*"
}

// Equals checks whether or not a player is another player
func (p *Player) Equals(p2 *Player) bool {
	return p.Account == p2.Account
}

// Send send multiple packets to a player
func (p *Player) Send(packets ...*packet.Packet) {
	for _, pkt := range packets {
		p.Address.Send(pkt, p)
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

// SendMessage sends a message to a single player
func (p *Player) SendMessage(message string) {
	messagePacket := packet.New(packet.PacketTypeUDP, 0x03)
	messagePacket.AddFieldString(message)
	p.Send(messagePacket)
}

// Kick kicks a player out of the server
func (p *Player) Kick(reason string) {
	if reason == "" {
		reason = "Kicked!"
	}
	kickPacket := packet.New(packet.PacketTypeUDP, 0x02)
	reasonBytes := []byte(reason)
	kickPacket.AddField(reasonBytes)
	p.Send(kickPacket)
	SendMessage(p.Name + " has been kicked!")
	p.Remove()
}

// IsOperator checks if a player is allowed to run commands on the server
func (p *Player) IsOperator() bool {
	for _, currentOperator := range config.Operators {
		if currentOperator == p.Account {
			return true
		}
	}
	return false
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
		if p.Address.Compare(player.Address) {
			// Network channel closing
			close(player.TCPNetworkInput)
			// Player deletion
			delete(players, i)
			break
		}
	}
	UpdatePlayerList()
}

// UpdatePlayerList sends the packet 0x06 to make clients update the player list
func UpdatePlayerList() {
	buf := bytes.NewBuffer(nil)
	for i, player := range players {
		buf.WriteByte(i)
		buf.Write([]byte(player.Name))
		buf.WriteByte(0x00)
	}
	p := packet.New(packet.PacketTypeUDP, 0x06)
	bufferBytes := buf.Bytes()
	p.AddField(bufferBytes)
	for _, player := range players {
		player.Send(p)
	}
}
