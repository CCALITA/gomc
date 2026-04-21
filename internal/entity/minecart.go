package entity

import (
	"math"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Minecart speed constants.
const (
	// minecartPoweredSpeed is the speed on powered rails (blocks per second).
	minecartPoweredSpeed float32 = 8.0
	// minecartDeceleration is the speed lost per second on unpowered rails.
	minecartDeceleration float32 = 0.5
	// minecartSlopeAccel is the speed gained per second on downward slopes.
	minecartSlopeAccel float32 = 4.0
	// minecartMaxSpeed caps the minecart speed.
	minecartMaxSpeed float32 = 8.0
)

// MinecartSystem moves minecarts along rails each tick.
type MinecartSystem struct {
	// GetBlock returns the block ID at the given block position.
	// Must be set by the game layer.
	GetBlock func(x, y, z int32) uint16
}

// Update processes all entities with MinecartData and Transform components.
func (s *MinecartSystem) Update(w *ecs.World, dt float64) {
	if s.GetBlock == nil {
		return
	}

	dt32 := float32(dt)
	transformStore := ecs.GetStore[Transform](w)

	ecs.Query2[MinecartData, Transform](w, func(e ecs.Entity, mc *MinecartData, t *Transform) {
		bx := int32(math.Floor(float64(t.Position.X)))
		by := int32(math.Floor(float64(t.Position.Y)))
		bz := int32(math.Floor(float64(t.Position.Z)))

		currentBlock := s.GetBlock(bx, by, bz)

		if !block.IsRail(currentBlock) {
			mc.OnRail = false
			mc.Speed = 0
			return
		}

		mc.OnRail = true

		// Determine speed based on rail type.
		if block.IsPoweredRail(currentBlock) {
			mc.Speed = minecartPoweredSpeed
		} else {
			// Check for slope: rail with air above and rail below.
			blockBelow := s.GetBlock(bx, by-1, bz)
			if block.IsRail(blockBelow) {
				// Going downhill — accelerate.
				mc.Speed += minecartSlopeAccel * dt32
			} else {
				// Flat unpowered rail — decelerate.
				mc.Speed -= minecartDeceleration * dt32
			}
		}

		// Clamp speed.
		if mc.Speed < 0 {
			mc.Speed = 0
		}
		if mc.Speed > minecartMaxSpeed {
			mc.Speed = minecartMaxSpeed
		}

		if mc.Speed == 0 {
			return
		}

		// Move along direction.
		dir := mc.Direction.Normal()
		displacement := dir.Scale(mc.Speed * dt32)
		t.Position = t.Position.Add(displacement)

		// Check for rail at new position and adjust direction if needed.
		newBX := int32(math.Floor(float64(t.Position.X)))
		newBZ := int32(math.Floor(float64(t.Position.Z)))
		if newBX != bx || newBZ != bz {
			nextBlock := s.GetBlock(newBX, by, newBZ)
			if !block.IsRail(nextBlock) {
				// Try to find a rail in an adjacent horizontal direction (turn).
				turned := s.tryTurn(mc, newBX, by, newBZ)
				if !turned {
					// No rail ahead — stop.
					mc.Speed = 0
				}
			}
		}

		// Move rider with minecart.
		if mc.Rider != 0 {
			if riderT, ok := transformStore.Get(mc.Rider); ok {
				riderT.Position = t.Position.Add(mcmath.Vec3{Y: 0.7})
			}
		}
	})
}

// tryTurn attempts to redirect the minecart when it reaches a position
// without a rail ahead. It checks perpendicular directions for rails.
func (s *MinecartSystem) tryTurn(mc *MinecartData, bx, by, bz int32) bool {
	// Try turning left and right relative to current direction.
	candidates := perpendicularDirections(mc.Direction)
	for _, d := range candidates {
		n := d.Normal()
		checkX := bx + int32(n.X)
		checkZ := bz + int32(n.Z)
		if block.IsRail(s.GetBlock(checkX, by, checkZ)) {
			mc.Direction = d
			return true
		}
	}
	return false
}

// perpendicularDirections returns the two horizontal directions perpendicular
// to the given direction.
func perpendicularDirections(d mcmath.Direction) [2]mcmath.Direction {
	switch d {
	case mcmath.North, mcmath.South:
		return [2]mcmath.Direction{mcmath.East, mcmath.West}
	case mcmath.East, mcmath.West:
		return [2]mcmath.Direction{mcmath.North, mcmath.South}
	default:
		return [2]mcmath.Direction{mcmath.North, mcmath.East}
	}
}

// MountMinecart sets the rider of the minecart to the given entity.
func MountMinecart(w *ecs.World, minecart, rider ecs.Entity) {
	mc, ok := ecs.GetStore[MinecartData](w).Get(minecart)
	if !ok {
		return
	}
	mc.Rider = rider
}

// DismountMinecart clears the rider from the minecart.
func DismountMinecart(w *ecs.World, minecart ecs.Entity) {
	mc, ok := ecs.GetStore[MinecartData](w).Get(minecart)
	if !ok {
		return
	}
	mc.Rider = 0
}
