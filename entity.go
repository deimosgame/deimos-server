package main

import (
	"time"
)

type Entity struct {
	UUID string

	// Position
	X float32
	Y float32
	Z float32
	// Rotation
	XRotation float32
	YRotation float32
	ZRotation float32
	// Velocity
	XVelocity        float32
	YVelocity        float32
	ZVelocity        float32
	XAngularVelocity float32
	YAngularVelocity float32
	ZAngularVelocity float32

	ModelId string

	LastUpdate time.Time
}

// NextTick computes the state of the entity at the world's next tick
func (e *Entity) NextTick() {
	if time.Since(e.LastUpdate) < time.Millisecond*15 {
		return
	}

	e.X = e.X + e.XVelocity*tickRateSecs
	e.Y = e.Y + e.YVelocity*tickRateSecs
	e.Z = e.Z + e.ZVelocity*tickRateSecs

	e.XRotation = e.XRotation + e.XAngularVelocity*tickRateSecs
	e.YRotation = e.YRotation + e.YAngularVelocity*tickRateSecs
	e.ZRotation = e.ZRotation + e.ZAngularVelocity*tickRateSecs
}
