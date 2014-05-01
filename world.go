package main

import (
	"time"
)

type World struct {
	players  []*Player
	entities []*Entity
}

// WorldSimulation does all the world simulation work
func WorldSimulation() {
	for {
		time.Sleep(50 * time.Millisecond)
	}
}
