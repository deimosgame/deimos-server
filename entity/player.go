package entity

import "net"

type Player struct {
	Name          string
	AkadokAccount string
	Addr          *net.UDPAddr

	// Position
	X float32
	Y float32
	Z float32
	// Rotation
	BodyRotation float32
	HeadRotation float32
	// Velocity
	XVelocity           float32
	YVelocity           float32
	ZVelocity           float32
	BodyAngularVelocity float32
	HeadAngularVelocity float32

	// Misc values
	ModelId        string
	SelectedWeapon byte
}
