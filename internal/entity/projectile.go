package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// arrowGravity is the gravitational acceleration for arrows (blocks/s^2).
const arrowGravity float32 = 20.0

// arrowLifetime is the default lifetime for an arrow entity in seconds.
const arrowLifetime float64 = 60.0

// arrowBBox is the bounding box for an arrow entity (0.1 cube).
var arrowBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.05, Y: -0.05, Z: -0.05},
	Max: mcmath.Vec3{X: 0.05, Y: 0.05, Z: 0.05},
}

// Projectile is an ECS component for projectile entities such as arrows.
type Projectile struct {
	Power    float32
	InGround bool
}

// ProjectileSystem detects ground and block collisions for projectile
// entities, setting InGround and zeroing velocity when a hit occurs.
// It operates on entities that have Transform, Projectile, and PhysicsBody
// components. Gravity is handled by PhysicsSystem; this system only manages
// the projectile-specific collision state.
type ProjectileSystem struct {
	// GetBlockAABBs returns solid block AABBs overlapping the given region.
	// When nil, only PhysicsBody.OnGround is used for ground detection.
	GetBlockAABBs func(mcmath.AABB) []mcmath.AABB
}

// Update checks each projectile for ground or block collisions.
func (s *ProjectileSystem) Update(w *ecs.World, dt float64) {
	ecs.Query3[Transform, Projectile, PhysicsBody](w, func(e ecs.Entity, t *Transform, proj *Projectile, pb *PhysicsBody) {
		if proj.InGround {
			// Projectile is stuck; zero out velocity.
			pb.Body.Velocity = mcmath.Vec3{}
			return
		}

		// Check whether the physics body has landed on ground (set by PhysicsSystem).
		if pb.Body.OnGround {
			proj.InGround = true
			pb.Body.Velocity = mcmath.Vec3{}
			return
		}

		// Also check block collision at the current position when a
		// block-query function is provided.
		if s.GetBlockAABBs != nil {
			worldAABB := pb.Body.WorldAABB()
			blocks := s.GetBlockAABBs(worldAABB)
			for _, b := range blocks {
				if worldAABB.Intersects(b) {
					proj.InGround = true
					pb.Body.Velocity = mcmath.Vec3{}
					return
				}
			}
		}
	})
}

// SpawnArrow creates an arrow projectile entity with the given position,
// velocity, and power. The arrow has a 0.1-cube AABB, a PhysicsBody with
// arrow-specific gravity, a 60-second Lifetime, and a Projectile component.
func SpawnArrow(w *ecs.World, pos, velocity mcmath.Vec3, power float32) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(arrowBBox)
	body.Position = pos
	body.Velocity = velocity
	body.Gravity = arrowGravity
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeArrow})
	ecs.GetStore[Projectile](w).Set(e, Projectile{Power: power})
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: arrowLifetime})

	return e
}
