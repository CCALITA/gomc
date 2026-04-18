// Package player defines the PlayerCamera interface used by the Controller
// to decouple player logic from the concrete render.Camera implementation.
package player

import "github.com/fanxiyao/gomc/internal/mcmath"

// PlayerCamera is the minimal camera interface the player controller needs
// for movement, aiming, and interaction raycasting. Concrete implementations
// (e.g. render.Camera) satisfy this interface.
type PlayerCamera interface {
	Forward() mcmath.Vec3
	Right() mcmath.Vec3
	Rotate(deltaYaw, deltaPitch float32)
	GetPosition() mcmath.Vec3
	SetPosition(pos mcmath.Vec3)
}
