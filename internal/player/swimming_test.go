//go:build !ci

package player

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
	"github.com/stretchr/testify/assert"
)

// fakeBlockLookup returns a BlockAtPositionFn backed by a map of positions
// to block IDs, defaulting to block.Air.
func fakeBlockLookup(blocks map[mcmath.BlockPos]uint16) BlockAtPositionFn {
	return func(pos mcmath.BlockPos) uint16 {
		if id, ok := blocks[pos]; ok {
			return id
		}
		return block.Air
	}
}

func TestNewSwimState(t *testing.T) {
	s := NewSwimState()
	assert.False(t, s.InWater)
	assert.False(t, s.Submerged)
	assert.Equal(t, MaxBreathTicks, s.BreathTicks)
	assert.Equal(t, DefaultSwimSpeedFactor, s.SwimSpeed)
}

func TestWaterDetection_FeetInWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, s.InWater)
	assert.False(t, s.Submerged)
}

func TestWaterDetection_FullySubmerged(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
		eyeBlock:  block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, s.InWater)
	assert.True(t, s.Submerged)
}

func TestWaterDetection_FlowingWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.FlowingWater,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, s.InWater)
}

func TestWaterDetection_NotInWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})

	s.Update(body, fakeBlockLookup(nil), playerPos, eyePos, SwimInput{}, 0.05)

	assert.False(t, s.InWater)
	assert.False(t, s.Submerged)
}

func TestSwimSpeedReduction(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.Velocity.X = 10.0
	body.Velocity.Z = 10.0

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.InDelta(t, 10.0*DefaultSwimSpeedFactor, body.Velocity.X, 0.01)
	assert.InDelta(t, 10.0*DefaultSwimSpeedFactor, body.Velocity.Z, 0.01)
}

func TestSwimDisablesGravity(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.Gravity = physics.DefaultGravity

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.Equal(t, float32(0), body.Gravity)
}

func TestBreathCountdown_Submerged(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
		eyeBlock:  block.Water,
	}

	// One second = 20 ticks.
	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 1.0)

	assert.Equal(t, MaxBreathTicks-20, s.BreathTicks)
}

func TestBreathCountdown_ReachesZero(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
		eyeBlock:  block.Water,
	}

	// Submerge for 16 seconds (320 ticks > MaxBreathTicks=300).
	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 16.0)

	assert.Equal(t, 0, s.BreathTicks)
}

func TestBreathReset_OnLeavingWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	submergedBlocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
		eyeBlock:  block.Water,
	}

	// Deplete some breath.
	s.Update(body, fakeBlockLookup(submergedBlocks), playerPos, eyePos, SwimInput{}, 5.0)
	assert.Less(t, s.BreathTicks, MaxBreathTicks)

	// Surface (head out of water).
	s.Update(body, fakeBlockLookup(nil), playerPos, eyePos, SwimInput{}, 0.05)

	assert.Equal(t, MaxBreathTicks, s.BreathTicks)
}

func TestBreathReset_HeadAboveWaterFeetStillIn(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()
	eyeBlock := eyePos.Floor()

	submergedBlocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
		eyeBlock:  block.Water,
	}

	// Deplete some breath while fully submerged.
	s.Update(body, fakeBlockLookup(submergedBlocks), playerPos, eyePos, SwimInput{}, 5.0)
	assert.Less(t, s.BreathTicks, MaxBreathTicks)

	// Head surfaces but feet still in water.
	feetOnlyBlocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}
	s.Update(body, fakeBlockLookup(feetOnlyBlocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, s.InWater)
	assert.False(t, s.Submerged)
	assert.Equal(t, MaxBreathTicks, s.BreathTicks)
}

func TestJumpInWater_SwimsUp(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{JumpPressed: true}, 0.05)

	assert.Equal(t, SwimUpVelocity, body.Velocity.Y)
}

func TestSneakInWater_Sinks(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{SneakPressed: true}, 0.05)

	assert.Equal(t, SwimSinkVelocity, body.Velocity.Y)
}

func TestNoFallDamage_IntoWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	// First call: player enters water (wasInWater = false).
	entered := s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, entered, "should report water entry")
}

func TestNoFallDamage_AlreadyInWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	// Enter water.
	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	// Already in water: should not report a new entry.
	entered := s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)
	assert.False(t, entered, "should not report entry when already in water")
}

func TestSubmergedVsFeetOnly(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	// Feet only: InWater true, Submerged false.
	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}
	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.True(t, s.InWater)
	assert.False(t, s.Submerged)

	// Breath should remain full since head is not submerged.
	assert.Equal(t, MaxBreathTicks, s.BreathTicks)
}

func TestDefaultVerticalVelocity_InWater(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	blocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	// No jump/sneak: gentle sinking.
	s.Update(body, fakeBlockLookup(blocks), playerPos, eyePos, SwimInput{}, 0.05)

	assert.Equal(t, SwimBuoyancy, body.Velocity.Y)
}

func TestGravityRestoredOnExit(t *testing.T) {
	s := NewSwimState()
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.Gravity = physics.DefaultGravity

	playerPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	eyePos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	feetBlock := playerPos.Floor()

	waterBlocks := map[mcmath.BlockPos]uint16{
		feetBlock: block.Water,
	}

	// Enter water: gravity set to 0.
	s.Update(body, fakeBlockLookup(waterBlocks), playerPos, eyePos, SwimInput{}, 0.05)
	assert.Equal(t, float32(0), body.Gravity)

	// Leave water: gravity is not restored by SwimState (that is the
	// controller's responsibility), but InWater should be false.
	s.Update(body, fakeBlockLookup(nil), playerPos, eyePos, SwimInput{}, 0.05)
	assert.False(t, s.InWater)
}
