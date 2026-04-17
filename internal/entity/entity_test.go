package entity

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// ---------------------------------------------------------------------------
// Component storage
// ---------------------------------------------------------------------------

func TestComponentStorage(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	// Store and retrieve Transform.
	ecs.GetStore[Transform](w).Set(e, Transform{
		Position: mcmath.Vec3{X: 1, Y: 2, Z: 3},
		Yaw:      90,
		Pitch:    -45,
	})
	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(1), tr.Position.X)
	assert.Equal(t, float32(90), tr.Yaw)

	// Store and retrieve Health.
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(20), h.Current)

	// Store and retrieve EntityTypeComp.
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeZombie})
	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeZombie, et.Type)

	// Store and retrieve AI.
	ecs.GetStore[AI](w).Set(e, AI{State: AIChase, Target: 42, Timer: 5.0})
	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, AIChase, ai.State)
	assert.Equal(t, ecs.Entity(42), ai.Target)

	// Store and retrieve Inventory.
	inv := Inventory{HotbarIndex: 3}
	inv.Slots[0] = ItemSlot{ItemID: 1, Count: 64}
	ecs.GetStore[Inventory](w).Set(e, inv)
	gotInv, ok := ecs.GetStore[Inventory](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 3, gotInv.HotbarIndex)
	assert.Equal(t, uint16(1), gotInv.Slots[0].ItemID)

	// Store and retrieve Lifetime.
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: 60.0})
	lt, ok := ecs.GetStore[Lifetime](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 60.0, lt.Remaining)

	// Store and retrieve Damage.
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 5, Knockback: mcmath.Vec3{X: 1}})
	d, ok := ecs.GetStore[Damage](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(5), d.Amount)

	// Store and retrieve Name.
	ecs.GetStore[Name](w).Set(e, Name{Value: "Steve"})
	n, ok := ecs.GetStore[Name](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, "Steve", n.Value)

	// Store and retrieve PhysicsBody.
	body := physics.NewBody(mcmath.AABB{})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})
	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.NotNil(t, pb.Body)
}

// ---------------------------------------------------------------------------
// PhysicsSystem
// ---------------------------------------------------------------------------

func TestPhysicsSystemUpdatesPosition(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	body := physics.NewBody(mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	})
	body.Position = mcmath.Vec3{X: 0, Y: 100, Z: 0}
	body.Velocity = mcmath.Vec3{X: 10, Y: 0, Z: 0}
	body.NoClip = true // no collision for a clean test

	ecs.GetStore[Transform](w).Set(e, Transform{Position: body.Position})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	sys := &PhysicsSystem{}
	sys.Update(w, 1.0)

	tr, _ := ecs.GetStore[Transform](w).Get(e)
	// With NoClip the body should have moved by velocity * dt.
	assert.InDelta(t, 10.0, float64(tr.Position.X), 1.0)
}

func TestPhysicsSystemWithCollision(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	body := physics.NewBody(mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	})
	body.Position = mcmath.Vec3{X: 0, Y: 10, Z: 0}
	body.Velocity = mcmath.Vec3{}

	ecs.GetStore[Transform](w).Set(e, Transform{Position: body.Position})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	// Floor at y=0.
	floor := mcmath.AABB{
		Min: mcmath.Vec3{X: -100, Y: -1, Z: -100},
		Max: mcmath.Vec3{X: 100, Y: 0, Z: 100},
	}

	sys := &PhysicsSystem{
		GetBlockAABBs: func(_ mcmath.AABB) []mcmath.AABB {
			return []mcmath.AABB{floor}
		},
	}

	// Simulate several ticks of falling.
	for range 100 {
		sys.Update(w, 0.05)
	}

	tr, _ := ecs.GetStore[Transform](w).Get(e)
	pb, _ := ecs.GetStore[PhysicsBody](w).Get(e)
	// Entity should have landed on the ground.
	assert.InDelta(t, 0.0, float64(tr.Position.Y), 0.1)
	assert.True(t, pb.Body.OnGround)
}

// ---------------------------------------------------------------------------
// AISystem
// ---------------------------------------------------------------------------

func TestAIIdleToChase(t *testing.T) {
	w := ecs.NewWorld()

	// Create a player within chase range.
	playerPos := mcmath.Vec3{X: 5, Y: 0, Z: 0}
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: playerPos})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Create a mob.
	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIIdle, Timer: 10})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIChase, ai.State)
	assert.Equal(t, player, ai.Target)
}

func TestAIChaseToAttack(t *testing.T) {
	w := ecs.NewWorld()

	// Player within attack range.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 1, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIChase, Target: player})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIAttack, ai.State)
}

func TestAIChaseToIdle_PlayerFar(t *testing.T) {
	w := ecs.NewWorld()

	// Player far away.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 100, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIChase, Target: player})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIIdle, ai.State)
}

func TestAIIdleToWander(t *testing.T) {
	w := ecs.NewWorld()

	// No player at all; timer expired.
	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIIdle, Timer: 0})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIWander, ai.State)
	assert.Greater(t, ai.Timer, 0.0)
}

func TestAIWanderToIdle(t *testing.T) {
	w := ecs.NewWorld()

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIWander, Timer: 0})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIIdle, ai.State)
}

func TestAIAttackToChase_PlayerMoved(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 5, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIAttack, Target: player})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIChase, ai.State)
}

func TestAIFleeToIdle(t *testing.T) {
	w := ecs.NewWorld()

	// Player far away.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 100, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIFlee})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(mob)
	assert.Equal(t, AIIdle, ai.State)
}

// ---------------------------------------------------------------------------
// LifetimeSystem
// ---------------------------------------------------------------------------

func TestLifetimeExpiry(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: 1.0})

	sys := &LifetimeSystem{}

	// First tick: still alive.
	sys.Update(w, 0.5)
	assert.True(t, w.Alive(e))

	lt, _ := ecs.GetStore[Lifetime](w).Get(e)
	assert.InDelta(t, 0.5, lt.Remaining, 0.001)

	// Second tick: should be destroyed.
	sys.Update(w, 0.6)
	assert.False(t, w.Alive(e))
}

func TestLifetimeMultipleEntities(t *testing.T) {
	w := ecs.NewWorld()

	e1 := w.NewEntity()
	ecs.GetStore[Lifetime](w).Set(e1, Lifetime{Remaining: 0.5})

	e2 := w.NewEntity()
	ecs.GetStore[Lifetime](w).Set(e2, Lifetime{Remaining: 2.0})

	sys := &LifetimeSystem{}
	sys.Update(w, 1.0)

	assert.False(t, w.Alive(e1))
	assert.True(t, w.Alive(e2))
}

// ---------------------------------------------------------------------------
// DamageSystem
// ---------------------------------------------------------------------------

func TestDamageApplication(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	body := physics.NewBody(mcmath.AABB{})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Damage](w).Set(e, Damage{
		Amount:    5,
		Knockback: mcmath.Vec3{X: 0, Y: 5, Z: 0},
	})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	h, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(15), h.Current)

	pb, _ := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.Equal(t, float32(5), pb.Body.Velocity.Y)

	// Damage component should be removed after processing.
	assert.False(t, ecs.GetStore[Damage](w).Has(e))
}

func TestDamageWithoutPhysicsBody(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 10, Max: 10})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 3})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	h, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(7), h.Current)
	assert.False(t, ecs.GetStore[Damage](w).Has(e))
}

// ---------------------------------------------------------------------------
// HealthSystem
// ---------------------------------------------------------------------------

func TestHealthDeath(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 0, Max: 20})

	sys := &HealthSystem{}
	sys.Update(w, 0.0)

	assert.False(t, w.Alive(e))
}

func TestHealthAlive(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 10, Max: 20})

	sys := &HealthSystem{}
	sys.Update(w, 0.0)

	assert.True(t, w.Alive(e))
}

func TestHealthNegativeDeath(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: -5, Max: 20})

	sys := &HealthSystem{}
	sys.Update(w, 0.0)

	assert.False(t, w.Alive(e))
}

// ---------------------------------------------------------------------------
// Factory functions
// ---------------------------------------------------------------------------

func TestSpawnPlayer(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := SpawnPlayer(w, "Steve", pos)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(20), h.Current)
	assert.Equal(t, float32(20), h.Max)

	n, ok := ecs.GetStore[Name](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, "Steve", n.Value)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypePlayer, et.Type)

	_, ok = ecs.GetStore[Inventory](w).Get(e)
	assert.True(t, ok)
}

func TestSpawnZombie(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 60, Z: 5}
	e := SpawnZombie(w, pos)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(20), h.Current)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeZombie, et.Type)

	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, AIIdle, ai.State)

	_, ok = ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
}

func TestSpawnSkeleton(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: -3, Y: 50, Z: 7}
	e := SpawnSkeleton(w, pos)

	assert.True(t, w.Alive(e))

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeSkeleton, et.Type)

	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, AIIdle, ai.State)
}

func TestSpawnItem(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 0, Y: 65, Z: 0}
	vel := mcmath.Vec3{X: 1, Y: 3, Z: -1}
	e := SpawnItem(w, 256, pos, vel)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, vel, pb.Body.Velocity)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeItem, et.Type)

	lt, ok := ecs.GetStore[Lifetime](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 300.0, lt.Remaining)

	inv, ok := ecs.GetStore[Inventory](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, uint16(256), inv.Slots[0].ItemID)
	assert.Equal(t, 1, inv.Slots[0].Count)
}

// ---------------------------------------------------------------------------
// Integration: scheduler runs all systems together
// ---------------------------------------------------------------------------

func TestSchedulerIntegration(t *testing.T) {
	w := ecs.NewWorld()

	// Spawn a player and a zombie near the player.
	player := SpawnPlayer(w, "TestPlayer", mcmath.Vec3{X: 0, Y: 100, Z: 0})
	zombie := SpawnZombie(w, mcmath.Vec3{X: 5, Y: 100, Z: 0})

	// Spawn an item with short lifetime.
	item := SpawnItem(w, 1, mcmath.Vec3{X: 0, Y: 100, Z: 0}, mcmath.Vec3{})

	// Override lifetime for quick expiry.
	ecs.GetStore[Lifetime](w).Set(item, Lifetime{Remaining: 0.5})

	sched := ecs.NewScheduler()
	sched.Add(&PhysicsSystem{})
	sched.Add(&AISystem{Rand: rand.New(rand.NewSource(0))})
	sched.Add(&LifetimeSystem{})
	sched.Add(&DamageSystem{})
	sched.Add(&HealthSystem{})

	// Run a few ticks.
	for range 20 {
		sched.Update(w, 0.05)
	}

	// Player and zombie should still be alive.
	assert.True(t, w.Alive(player))
	assert.True(t, w.Alive(zombie))

	// Item should have expired (20 * 0.05 = 1.0s > 0.5s).
	assert.False(t, w.Alive(item))
}

// ---------------------------------------------------------------------------
// DamageSystem + HealthSystem pipeline
// ---------------------------------------------------------------------------

func TestDamageThenDeath(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 5, Max: 20})
	body := physics.NewBody(mcmath.AABB{})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 10})

	dmgSys := &DamageSystem{}
	hpSys := &HealthSystem{}

	dmgSys.Update(w, 0.0)
	hpSys.Update(w, 0.0)

	assert.False(t, w.Alive(e))
}

// ---------------------------------------------------------------------------
// AI chase movement
// ---------------------------------------------------------------------------

func TestAIChaseMoveTowardsPlayer(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 10, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	mob := w.NewEntity()
	ecs.GetStore[Transform](w).Set(mob, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[AI](w).Set(mob, AI{State: AIChase, Target: player})

	sys := &AISystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 1.0)

	tr, _ := ecs.GetStore[Transform](w).Get(mob)
	// Mob should have moved towards the player (positive X direction).
	assert.Greater(t, tr.Position.X, float32(0))
}
