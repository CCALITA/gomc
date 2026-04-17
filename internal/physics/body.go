package physics

import "github.com/fanxiyao/gomc/internal/mcmath"

const (
	// DefaultGravity is the default gravitational acceleration (blocks/s^2).
	DefaultGravity float32 = 28.0
	// DefaultAirDrag is the drag coefficient applied while airborne.
	DefaultAirDrag float32 = 0.02
	// DefaultGroundDrag is the drag coefficient applied while on the ground.
	DefaultGroundDrag float32 = 0.6
	// DefaultMaxFallSpeed is the terminal velocity (blocks/s).
	DefaultMaxFallSpeed float32 = 78.4
)

// Body represents a physics body with position, velocity, and collision state.
type Body struct {
	Position     mcmath.Vec3
	Velocity     mcmath.Vec3
	BoundingBox  mcmath.AABB // relative to Position
	OnGround     bool
	NoClip       bool
	Gravity      float32
	Drag         float32
	MaxFallSpeed float32
}

// NewBody creates a Body with sensible defaults and the given bounding box
// (specified relative to the body's position origin).
func NewBody(boundingBox mcmath.AABB) Body {
	return Body{
		BoundingBox:  boundingBox,
		Gravity:      DefaultGravity,
		Drag:         DefaultAirDrag,
		MaxFallSpeed: DefaultMaxFallSpeed,
	}
}

// WorldAABB returns the body's AABB in world space by offsetting BoundingBox
// by Position.
func (b *Body) WorldAABB() mcmath.AABB {
	return b.BoundingBox.Offset(b.Position)
}
