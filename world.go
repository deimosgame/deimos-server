package main

import (
	"time"
)

type World struct {
	players  []*Player
	entities []*Entity
}

var (
	tickRate = time.Millisecond * time.Duration(config.TickRate)
)

// WorldSimulation does all the world simulation work
func WorldSimulation() {
	for {
		start := time.Now()

		// TODO: Execute world simulation

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
