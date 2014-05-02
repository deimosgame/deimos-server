package main

type Entity struct {
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

func (e *Entity) NextTick() {
	if time.Since(player.LastUpdate) < time.Millisecond*15 {
		continue
	}

	e.X = e.X + e.XVelocity*tickRateSecs
	e.Y = e.Y + e.YVelocity*tickRateSecs
	e.Z = e.Z + e.ZVelocity*tickRateSecs

	p.XRotation = p.XRotation + p.AngularVelocityX*tickRateSecs
	p.YRotation = p.YRotation + p.AngularVelocityY*tickRateSecs
	p.ZRotation = p.ZRotation + p.AngularVelocityZ*tickRateSecs
}
