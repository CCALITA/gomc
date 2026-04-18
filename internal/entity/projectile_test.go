package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// SpawnArrow factory
// ---------------------------------------------------------------------------

func TestSpawnArrow_HasCorrectComponents(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	vel := mcmath.Vec3{X: 0, Y: 1, Z: 30}
	e := SpawnArrow(w, pos, vel, 2.5)

	assert.True(t, w.Alive(e))

	// Transform
	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	// PhysicsBody
	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)
	assert.Equal(t, vel, pb.Body.Velocity)
	assert.InDelta(t, 0.1, float64(pb.Body.BoundingBox.Size().X), 0.001)
	assert.InDelta(t, 0.1, float64(pb.Body.BoundingBox.Size().Y), 0.001)
	assert.InDelta(t, 0.1, float64(pb.Body.BoundingBox.Size().Z), 0.001)

	// EntityTypeComp
	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeArrow, et.Type)

	// Projectile
	proj, ok := ecs.GetStore[Projectile](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(2.5), proj.Power)
	assert.False(t, proj.InGround)

	// Lifetime
	lt, ok := ecs.GetStore[Lifetime](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 60.0, lt.Remaining)
}

func TestSpawnArrow_PowerStoredCorrectly(t *testing.T) {
	tests := []struct {
		name  string
		power float32
	}{
		{"zero power", 0.0},
		{"half power", 0.5},
		{"full power", 1.0},
		{"max power", 3.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := ecs.NewWorld()
			e := SpawnArrow(w, mcmath.Vec3{}, mcmath.Vec3{}, tc.power)

			proj, ok := ecs.GetStore[Projectile](w).Get(e)
			assert.True(t, ok)
			assert.Equal(t, tc.power, proj.Power)
		})
	}
}

func TestSpawnArrow_LifetimeSet(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnArrow(w, mcmath.Vec3{}, mcmath.Vec3{}, 1.0)

	lt, ok := ecs.GetStore[Lifetime](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 60.0, lt.Remaining)
}

func TestSpawnArrow_GravitySetToArrowValue(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnArrow(w, mcmath.Vec3{}, mcmath.Vec3{}, 1.0)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(20.0), pb.Body.Gravity)
}

// ---------------------------------------------------------------------------
// ProjectileSystem - gravity
// ---------------------------------------------------------------------------

func TestProjectileSystem_GravityAppliesEachTick(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 0, Y: 100, Z: 0}
	vel := mcmath.Vec3{X: 10, Y: 0, Z: 0}
	e := SpawnArrow(w, pos, vel, 1.0)

	// Use a no-collision block getter so gravity is applied but no
	// collisions occur.
	physSys := &PhysicsSystem{
		GetBlockAABBs: func(_ mcmath.AABB) []mcmath.AABB { return nil },
	}
	projSys := &ProjectileSystem{}

	// Run one tick at dt=1.0s. The PhysicsSystem applies gravity via
	// ResolveCollision, and after that the position should have changed.
	physSys.Update(w, 1.0)
	projSys.Update(w, 1.0)

	tr, _ := ecs.GetStore[Transform](w).Get(e)
	// The arrow should have moved in X and fallen in Y due to gravity.
	assert.Greater(t, tr.Position.X, float32(0))
	assert.Less(t, tr.Position.Y, float32(100))
}

func TestProjectileSystem_InGroundStopsMovement(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnArrow(w, mcmath.Vec3{X: 0, Y: 10, Z: 0}, mcmath.Vec3{X: 5, Y: 0, Z: 5}, 1.0)

	// Manually set InGround.
	proj, _ := ecs.GetStore[Projectile](w).Get(e)
	proj.InGround = true

	projSys := &ProjectileSystem{}
	projSys.Update(w, 1.0)

	pb, _ := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.Equal(t, float32(0), pb.Body.Velocity.X)
	assert.Equal(t, float32(0), pb.Body.Velocity.Y)
	assert.Equal(t, float32(0), pb.Body.Velocity.Z)
}

func TestProjectileSystem_BlockCollisionSetsInGround(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 5, Z: 5}
	e := SpawnArrow(w, pos, mcmath.Vec3{}, 1.0)

	// Provide a block that overlaps the arrow position.
	block := mcmath.AABB{
		Min: mcmath.Vec3{X: 4, Y: 4, Z: 4},
		Max: mcmath.Vec3{X: 6, Y: 6, Z: 6},
	}

	projSys := &ProjectileSystem{
		GetBlockAABBs: func(_ mcmath.AABB) []mcmath.AABB {
			return []mcmath.AABB{block}
		},
	}
	projSys.Update(w, 0.05)

	proj, _ := ecs.GetStore[Projectile](w).Get(e)
	assert.True(t, proj.InGround)
}

func TestProjectileSystem_PhysicsOnGroundSetsInGround(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnArrow(w, mcmath.Vec3{X: 0, Y: 0, Z: 0}, mcmath.Vec3{}, 1.0)

	// Simulate the physics body reporting OnGround.
	pb, _ := ecs.GetStore[PhysicsBody](w).Get(e)
	pb.Body.OnGround = true

	projSys := &ProjectileSystem{}
	projSys.Update(w, 0.05)

	proj, _ := ecs.GetStore[Projectile](w).Get(e)
	assert.True(t, proj.InGround)

	// Velocity should be zeroed.
	assert.Equal(t, float32(0), pb.Body.Velocity.X)
	assert.Equal(t, float32(0), pb.Body.Velocity.Y)
	assert.Equal(t, float32(0), pb.Body.Velocity.Z)
}

// ---------------------------------------------------------------------------
// Integration: Lifetime system cleans up arrow
// ---------------------------------------------------------------------------

func TestArrow_LifetimeExpiry(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnArrow(w, mcmath.Vec3{}, mcmath.Vec3{}, 1.0)

	// Override lifetime to expire quickly.
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: 0.5})

	sys := &LifetimeSystem{}

	// First tick: still alive.
	sys.Update(w, 0.3)
	assert.True(t, w.Alive(e))

	// Second tick: should be destroyed.
	sys.Update(w, 0.3)
	assert.False(t, w.Alive(e))
}
