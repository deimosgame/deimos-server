package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"bytes"
	"encoding/binary"
	"reflect"
	"time"
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
		i := byte(0)
		for _, p := range players {
			x := *p
			save.Players[i] = &x
			i++
		}
		save.Entities = make([]*Entity, len(entities))
		i = 0
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
			player.Send(save.Packet(worldSnapshotId-1,
				player.LastAcknowledged)...)
		}

		// Check if the calculation took more than the tick rate value
		diff := time.Since(start)
		if diff < tickRate {
			if serverKeepupAlert {
				serverKeepupAlert = false
				log.Notice("Server is synchronized again")
			}
			time.Sleep(tickRate - diff)
		} else if !serverKeepupAlert {
			serverKeepupAlert = true
			log.Warn("Server can't keep up! Lower the tick rate!")
		}
	}
}

func (w *World) Packet(uuid uint32, compareTo *World) []*packet.Packet {
	packets, i := make([]*packet.Packet, 1), 0
	packets[i] = packet.New(0x04)

	idBuf := bytes.NewBuffer(nil)
	binary.Write(idBuf, binary.LittleEndian, uuid)
	bufBytes := idBuf.Bytes()
	packets[i].AddField(&bufBytes)

	if len(w.Players) > 0 {
		addedField := false

		for j, p1 := range w.Players {
			var newBytes []byte

			if !compareTo.Initialized {
				newBytes = makePlayerPacket(p1, &Player{})
			} else {
				// Search for player's previous state in the other world
				playerExists := false
				for _, p2 := range compareTo.Players {
					if p1.Address.String() == p2.Address.String() {
						playerExists = true
						newBytes = makePlayerPacket(p1, p2)
						break
					}
				}

				if !playerExists {
					newBytes = makePlayerPacket(p1, &Player{})
				}
			}

			// Smooth splitting
			if len(packets[i].Data)+len(newBytes)+2 > packet.PacketSize {
				packets = append(packets, packet.New(0x04))
				i++
			}

			// Player prefix + player ID
			if !addedField && len(newBytes) > 0 {
				addedField = true
				packets[i].AddFieldBytes(byte('A'), j)
			}

			packets[i].AddField(&newBytes)
		}
	}
	return packets
}

// makePlayerPacket creates a player element in the world packet based on
// another player (p2 can be empty player)
func makePlayerPacket(p1, p2 *Player) []byte {
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

		// Write new data to packet
		buf.Write(prefix)
		switch fieldType.Type.String() {
		case "string":
			buf.Write([]byte(fieldValue1.(string)))
		case "float32":
			binary.Write(buf, binary.LittleEndian, fieldValue1.(float32))
		}
	}
	return buf.Bytes()
}
