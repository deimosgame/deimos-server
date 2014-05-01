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
	tickRate := time.Millisecond * time.Duration(config.TickRate)
	for {
		start := time.Now()

		// Execute world simulation
		for _, player := range players {
			player.NextTick()
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
