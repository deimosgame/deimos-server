package main

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"time"

	"github.com/deimosgame/deimos-server/packet"
)

type World struct {
	Players     map[byte]*Player
	Entities    []*Entity
	Time        time.Time
	Initialized bool
}

// WorldSimulation does all the world simulation work
func WorldSimulation() {
	tickRate := time.Millisecond * time.Duration(config.Tickrate)
	for {
		start := time.Now()

		// Execute world simulation
		for _, player := range players {
			player.NextTick()
		}
		for entity, _ := range entities {
			entity.NextTick()
		}

		// Remove world snapshots older than 10 seconds
		for id, snapshot := range worldSnapshots {
			if time.Since(snapshot.Time) > time.Second*10 {
				delete(worldSnapshots, id)
			}
		}

		// Save the current world state as a snapshot
		save := &World{Initialized: true}
		save.Players = make(map[byte]*Player)
		for i, p := range players {
			x := *p
			save.Players[i] = &x
		}
		save.Entities = make([]*Entity, len(entities))
		i := byte(0)
		for e, _ := range entities {
			x := *e
			save.Entities[i] = &x
			i++
		}
		save.Time = time.Now()
		worldSnapshots[worldSnapshotId] = save
		worldSnapshotId++

		// Broadcast the snapshot to players
		for _, player := range players {
			p := save.Packet(worldSnapshotId-1, player)
			player.Send(p...)
		}

		// Check if the calculation took more than the tick rate value
		diff := time.Since(start)
		if diff < tickRate {
			if serverKeepupAlert {
				serverKeepupAlert = false
				// log.Notice("Server is synchronized again")
			}
			time.Sleep(tickRate - diff)
		} else if !serverKeepupAlert {
			serverKeepupAlert = true
			// log.Warn("Server can't keep up! Lower the tick rate!")
		}
	}
}

// Packet generates packets used to broadcast the world state to specific
// players
func (w *World) Packet(uuid uint32, receiver *Player) []*packet.Packet {
	packets, i := make([]*packet.Packet, 1), 0
	packets[i] = packet.New(packet.PacketTypeUDP, 0x04)

	idBuf := bytes.NewBuffer(nil)
	binary.Write(idBuf, binary.LittleEndian, uuid)
	bufBytes := idBuf.Bytes()
	packets[i].AddField(bufBytes)

	if len(w.Players) == 0 {
		return packets
	}

	for j, p1 := range w.Players {
		// Do not send the player if he is the receiver
		if p1.Equals(receiver) {
			continue
		}

		var newBytes []byte

		if receiver.LastAcknowledged == nil ||
			!receiver.LastAcknowledged.Initialized {
			newBytes = makePlayerPacket(j, p1, &Player{})
		} else {
			// Search for player's previous state in the other world
			playerExists := false
			for k, p2 := range receiver.LastAcknowledged.Players {
				if k == j && p2.Equals(p1) {
					playerExists = true
					newBytes = makePlayerPacket(j, p1, p2)
					break
				}
			}

			if !playerExists {
				newBytes = makePlayerPacket(j, p1, &Player{})
			}
		}

		// Smooth splitting
		if len(packets[i].Data)+len(newBytes)+2 > packet.PacketSize {
			packets = append(packets, packet.New(packet.PacketTypeUDP,
				0x04))
			i++
		}

		packets[i].AddField(newBytes)
	}
	return packets
}

// makePlayerPacket creates a player element in the world packet based on
// another player (p2 can be empty player)
func makePlayerPacket(id byte, p1, p2 *Player) []byte {
	buf := bytes.NewBuffer(nil)

	val := reflect.ValueOf(p1).Elem()
	var val2 reflect.Value
	if p2.Initialized {
		val2 = reflect.ValueOf(p2).Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		fieldValue1, fieldType := val.Field(i).Interface(),
			val.Type().Field(i)
		fieldTag := fieldType.Tag

		// Check if data has to be sent
		prefix := []byte(fieldTag.Get("prefix"))
		if len(prefix) == 0 {
			continue
		}

		// Compare p1 to p2
		if p2.Initialized {
			fieldValue2 := val2.Field(i).Interface()
			if fieldValue1 == fieldValue2 {
				continue
			}
		}

		buf.Write([]byte{byte('A'), id})

		// Write new data to packet
		buf.Write(prefix)
		switch fieldType.Type.String() {
		case "string":
			buf.Write([]byte(fieldValue1.(string)))
			buf.WriteByte(0x00)
		case "float32":
			binary.Write(buf, binary.LittleEndian, fieldValue1.(float32))
		case "byte", "uint8":
			buf.WriteByte(fieldValue1.(byte))
		default:
			log.Info(fieldType.Type.String(), fieldType.Name)
			log.Panic("Unknown data type encountered when encoding broadcast " +
				"packet!")
		}
	}
	return buf.Bytes()
}

// SendMessage messages all players on the server
func SendMessage(message string) {
	messagePacket := packet.New(packet.PacketTypeUDP, 0x03)
	messagePacket.AddFieldString(message)
	for _, currentPlayer := range players {
		currentPlayer.Send(messagePacket)
	}
	log.Info(message)
}
