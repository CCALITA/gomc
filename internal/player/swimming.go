package player

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

const (
	// MaxBreathTicks is the number of ticks a player can remain submerged
	// before drowning begins. At 20 ticks per second this equals 15 seconds.
	MaxBreathTicks int = 300

	// DefaultSwimSpeedFactor is the horizontal velocity multiplier while
	// the player is in water (0.6x normal speed).
	DefaultSwimSpeedFactor float32 = 0.6

	// SwimUpVelocity is the vertical velocity applied when the player
	// presses the jump key while in water.
	SwimUpVelocity float32 = 3.0

	// SwimSinkVelocity is the vertical velocity applied when the player
	// presses the sneak key while in water.
	SwimSinkVelocity float32 = -2.0

	// SwimBuoyancy is the passive downward velocity applied while in water
	// when neither jump nor sneak is pressed, simulating gentle sinking.
	SwimBuoyancy float32 = -1.0
)

// SwimState tracks whether the player is in water, whether the player is
// fully submerged, the remaining breath ticks, and the swim speed factor.
type SwimState struct {
	InWater    bool
	Submerged  bool
	BreathTicks int
	SwimSpeed  float32
}

// NewSwimState returns a SwimState initialised with full breath and the
// default swim speed factor.
func NewSwimState() *SwimState {
	return &SwimState{
		BreathTicks: MaxBreathTicks,
		SwimSpeed:   DefaultSwimSpeedFactor,
	}
}

// BlockAtPositionFn abstracts world block lookups for testability.
type BlockAtPositionFn func(pos mcmath.BlockPos) uint16

// SwimInput bundles the input flags relevant to swimming so that
// Update does not depend directly on the input package.
type SwimInput struct {
	JumpPressed  bool
	SneakPressed bool
}

// Update checks the blocks at the player's feet and eyes, then adjusts
// physics state accordingly. It returns whether the player entered water
// this tick (for fall damage cancellation by the caller).
//
// Parameters:
//   - body: the player's physics body (velocities and gravity are modified)
//   - getBlock: block lookup function
//   - playerPos: world position at the player's feet
//   - eyePos: world position at the player's eye height
//   - inp: jump/sneak input state
//   - dt: seconds elapsed since last tick
func (s *SwimState) Update(
	body *physics.Body,
	getBlock BlockAtPositionFn,
	playerPos mcmath.Vec3,
	eyePos mcmath.Vec3,
	inp SwimInput,
	dt float32,
) bool {
	wasInWater := s.InWater

	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	s.InWater = block.IsWater(getBlock(feetBlock))
	s.Submerged = block.IsWater(getBlock(eyeBlock))

	if s.InWater {
		s.applySwimPhysics(body, inp)
	}

	s.updateBreath(dt)

	// Report water entry for fall damage cancellation.
	return s.InWater && !wasInWater
}

// applySwimPhysics modifies the body velocity for swimming: reduces
// horizontal speed, disables gravity, and maps jump/sneak to vertical
// movement.
func (s *SwimState) applySwimPhysics(body *physics.Body, inp SwimInput) {
	// Reduce horizontal velocity.
	body.Velocity.X *= s.SwimSpeed
	body.Velocity.Z *= s.SwimSpeed

	// Disable gravity while in water.
	body.Gravity = 0

	// Vertical control.
	switch {
	case inp.JumpPressed:
		body.Velocity.Y = SwimUpVelocity
	case inp.SneakPressed:
		body.Velocity.Y = SwimSinkVelocity
	default:
		body.Velocity.Y = SwimBuoyancy
	}
}

// updateBreath decrements or resets BreathTicks based on submersion state.
func (s *SwimState) updateBreath(dt float32) {
	if s.Submerged {
		// Decrement one tick per 1/20 second. dt is in seconds, so
		// convert to ticks (20 ticks/s).
		ticksToRemove := int(dt * 20)
		if ticksToRemove < 1 {
			ticksToRemove = 1
		}
		s.BreathTicks -= ticksToRemove
		if s.BreathTicks < 0 {
			s.BreathTicks = 0
		}
	} else {
		s.BreathTicks = MaxBreathTicks
	}
}
