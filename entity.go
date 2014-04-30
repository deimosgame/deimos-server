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
}
