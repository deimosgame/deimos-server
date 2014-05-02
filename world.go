package main

import (
	"bitbucket.org/deimosgame/go-akadok/packet"
	"bytes"
	"encoding/binary"
	"reflect"
	"time"
)

type World struct {
	players  []*Player
	entities []*Entity
}

func (w *World) Packet() {
	p := packet.New(0x04)
	if len(w.players) > 0 {
		playerPrefix := byte('A')
		p.AddFieldBytes(playerPrefix)
		for _, player := range players {
			val := reflect.ValueOf(*player).Elem()
			for i := 0; i < val.NumField(); i++ {
				fieldValue, fieldType := val.Field(i).Interface(),
					val.Type().Field(i)
				fieldTag := fieldType.Tag
				prefix := []byte(fieldTag.Get("prefix"))
				if len(prefix) == 0 {
					continue
				}
				p.AddField(&prefix)
				switch fieldType.Type.String() {
				case "string":
					valueBytes := []byte(fieldValue.(string))
					p.AddField(&valueBytes)
				case "float32":
					buf := bytes.NewBuffer(nil)
					binary.Write(buf, binary.LittleEndian, fieldValue.(float32))
					bufBytes := buf.Bytes()
					p.AddField(&bufBytes)
				}
			}
		}
	}
}

// WorldSimulation does all the world simulation work
func WorldSimulation() {
	tickRate := time.Millisecond * time.Duration(config.TickRate)
	for {
		start := time.Now()

		// Execute world simulation
		for _, player := range players {
			player.NextTick()
		}
		for entity, _ := range entities {
			entity.NextTick()
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
