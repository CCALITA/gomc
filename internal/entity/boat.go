package entity

import (
	"math"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

const (
	// boatWaterSpeed is the boat speed on water (blocks/s).
	boatWaterSpeed float32 = 8.0
	// boatLandSpeed is the boat speed on land (blocks/s).
	boatLandSpeed float32 = 2.0
	// boatWaterSurfaceOffset is how far above the water surface the boat floats.
	boatWaterSurfaceOffset float32 = 0.1
	// boatHealth is the number of hits a boat can take before breaking.
	boatHealth float32 = 6.0
)

// boatBBox is the bounding box for a boat (1.375 wide, 0.5625 tall).
var boatBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.6875, Y: 0, Z: -0.6875},
	Max: mcmath.Vec3{X: 0.6875, Y: 0.5625, Z: 0.6875},
}

// BoatData holds boat-specific state for an entity.
type BoatData struct {
	Rider      ecs.Entity
	HasRider   bool
	WaterSpeed float32
	LandSpeed  float32
}

// BoatSystem processes boat entities each tick, handling water floating,
// speed changes, and rider movement.
type BoatSystem struct {
	// IsWaterAt reports whether the block at the given position is water.
	// When nil, no block is ever considered water.
	IsWaterAt func(pos mcmath.Vec3) bool

	// WaterSurfaceY returns the Y coordinate of the water surface at (X, Z).
	// When nil, defaults to the block Y position.
	WaterSurfaceY func(x, z float32) float32
}

// Update processes all boat entities.
func (s *BoatSystem) Update(w *ecs.World, _ float64) {
	ecs.Query3[BoatData, Transform, PhysicsBody](w, func(_ ecs.Entity, bd *BoatData, t *Transform, pb *PhysicsBody) {
		onWater := s.isWater(t.Position)

		if onWater {
			surfaceY := s.waterSurface(t.Position.X, t.Position.Z)
			t.Position.Y = surfaceY + boatWaterSurfaceOffset
			pb.Body.Position.Y = t.Position.Y
			pb.Body.Velocity.Y = 0
			pb.Body.Gravity = 0
		} else {
			pb.Body.Gravity = physics.DefaultGravity
		}

		if bd.HasRider {
			speed := bd.LandSpeed
			if onWater {
				speed = bd.WaterSpeed
			}

			vel := pb.Body.Velocity
			horizSq := vel.X*vel.X + vel.Z*vel.Z
			if horizSq > speed*speed {
				scale := speed / float32(math.Sqrt(float64(horizSq)))
				pb.Body.Velocity.X = vel.X * scale
				pb.Body.Velocity.Z = vel.Z * scale
			}

			riderTransform, ok := ecs.GetStore[Transform](w).Get(bd.Rider)
			if ok {
				riderTransform.Position = mcmath.Vec3{
					X: t.Position.X,
					Y: t.Position.Y + boatBBox.Max.Y,
					Z: t.Position.Z,
				}
			}
			riderPB, ok := ecs.GetStore[PhysicsBody](w).Get(bd.Rider)
			if ok {
				riderPB.Body.Position = riderTransform.Position
				riderPB.Body.Velocity = pb.Body.Velocity
			}
		}
	})
}

// MountBoat places a rider entity into the boat.
func MountBoat(w *ecs.World, boat, rider ecs.Entity) {
	bd, ok := ecs.GetStore[BoatData](w).Get(boat)
	if !ok || bd.HasRider {
		return
	}
	bd.Rider = rider
	bd.HasRider = true
}

// DismountBoat removes the rider from the boat.
func DismountBoat(w *ecs.World, boat ecs.Entity) {
	bd, ok := ecs.GetStore[BoatData](w).Get(boat)
	if !ok || !bd.HasRider {
		return
	}
	bd.HasRider = false
	bd.Rider = 0
}

// BreakBoat destroys the boat entity and drops a boat item.
func BreakBoat(w *ecs.World, boat ecs.Entity) {
	bd, ok := ecs.GetStore[BoatData](w).Get(boat)
	if ok && bd.HasRider {
		DismountBoat(w, boat)
	}

	t, ok := ecs.GetStore[Transform](w).Get(boat)
	if !ok {
		w.DestroyEntity(boat)
		return
	}

	pos := t.Position
	w.DestroyEntity(boat)

	// Drop a boat item with a small upward velocity.
	SpawnItem(w, item.Boat, pos, mcmath.Vec3{Y: 4})
}

// isWater checks whether the block at the given position is water.
func (s *BoatSystem) isWater(pos mcmath.Vec3) bool {
	if s.IsWaterAt == nil {
		return false
	}
	return s.IsWaterAt(pos) || s.IsWaterAt(mcmath.Vec3{X: pos.X, Y: pos.Y - 0.1, Z: pos.Z})
}

// waterSurface returns the Y coordinate of the water surface.
func (s *BoatSystem) waterSurface(x, z float32) float32 {
	if s.WaterSurfaceY != nil {
		return s.WaterSurfaceY(x, z)
	}
	return 0
}
