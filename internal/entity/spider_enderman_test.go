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
// Factory tests
// ---------------------------------------------------------------------------

func TestSpawnSpider(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 60, Z: 5}
	e := SpawnSpider(w, pos)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(16), h.Current)
	assert.Equal(t, float32(16), h.Max)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeSpider, et.Type)

	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, AIIdle, ai.State)
	assert.True(t, ai.SpiderClimb)
	assert.False(t, ai.EndermanTeleport)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)
	// Spider AABB: 1.4 wide, 0.9 tall
	assert.InDelta(t, 1.4, float64(pb.Body.BoundingBox.Max.X-pb.Body.BoundingBox.Min.X), 0.01)
	assert.InDelta(t, 0.9, float64(pb.Body.BoundingBox.Max.Y-pb.Body.BoundingBox.Min.Y), 0.01)
}

func TestSpawnEnderman(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: -3, Y: 50, Z: 7}
	e := SpawnEnderman(w, pos)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(40), h.Current)
	assert.Equal(t, float32(40), h.Max)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeEnderman, et.Type)

	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, AIIdle, ai.State)
	assert.True(t, ai.EndermanTeleport)
	assert.False(t, ai.SpiderClimb)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)
	// Enderman AABB: 0.6 wide, 2.9 tall
	assert.InDelta(t, 0.6, float64(pb.Body.BoundingBox.Max.X-pb.Body.BoundingBox.Min.X), 0.01)
	assert.InDelta(t, 2.9, float64(pb.Body.BoundingBox.Max.Y-pb.Body.BoundingBox.Min.Y), 0.01)
}

// ---------------------------------------------------------------------------
// Spider AI: hostile at night, passive during day
// ---------------------------------------------------------------------------

func TestSpiderHostileAtNight(t *testing.T) {
	w := ecs.NewWorld()

	// Player within chase range.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 5, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Spider.
	spider := w.NewEntity()
	ecs.GetStore[Transform](w).Set(spider, Transform{Position: mcmath.Vec3{}})
	body := physics.NewBody(spiderBBox)
	ecs.GetStore[PhysicsBody](w).Set(spider, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(spider, AI{State: AIIdle, Timer: 10, SpiderClimb: true})

	// Night tick (15000 is in the 13000-23000 range).
	sys := &AISystem{Rand: rand.New(rand.NewSource(42)), GameTick: 15000}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(spider)
	assert.Equal(t, AIChase, ai.State, "spider should chase player at night")
}

func TestSpiderPassiveDuringDay(t *testing.T) {
	w := ecs.NewWorld()

	// Player within chase range.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 5, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Spider.
	spider := w.NewEntity()
	ecs.GetStore[Transform](w).Set(spider, Transform{Position: mcmath.Vec3{}})
	body := physics.NewBody(spiderBBox)
	ecs.GetStore[PhysicsBody](w).Set(spider, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(spider, AI{State: AIIdle, Timer: 0, SpiderClimb: true})

	// Day tick (6000 is daytime).
	sys := &AISystem{Rand: rand.New(rand.NewSource(42)), GameTick: 6000}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(spider)
	// During the day, spider should behave passively (idle/wander), not chase.
	assert.NotEqual(t, AIChase, ai.State, "spider should not chase during the day")
	assert.NotEqual(t, AIAttack, ai.State, "spider should not attack during the day")
}

func TestSpiderClimbsDuringChase(t *testing.T) {
	w := ecs.NewWorld()

	// Player above the spider.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 5, Y: 10, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Spider already chasing.
	spider := w.NewEntity()
	ecs.GetStore[Transform](w).Set(spider, Transform{Position: mcmath.Vec3{}})
	body := physics.NewBody(spiderBBox)
	ecs.GetStore[PhysicsBody](w).Set(spider, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(spider, AI{State: AIChase, Target: player, SpiderClimb: true})

	// Night tick so spider stays hostile.
	sys := &AISystem{Rand: rand.New(rand.NewSource(42)), GameTick: 15000}
	sys.Update(w, 0.1)

	pb, _ := ecs.GetStore[PhysicsBody](w).Get(spider)
	assert.Equal(t, spiderClimbSpeed, pb.Body.Velocity.Y, "spider should have upward velocity when climbing")
}

// ---------------------------------------------------------------------------
// Enderman AI: teleport on damage
// ---------------------------------------------------------------------------

func TestEndermanTeleportOnDamage(t *testing.T) {
	w := ecs.NewWorld()

	enderman := SpawnEnderman(w, mcmath.Vec3{X: 10, Y: 64, Z: 10})

	originalPos := mcmath.Vec3{X: 10, Y: 64, Z: 10}

	// Apply damage.
	ecs.GetStore[Damage](w).Set(enderman, Damage{Amount: 5, Knockback: mcmath.Vec3{}})

	sys := &DamageSystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.0)

	tr, ok := ecs.GetStore[Transform](w).Get(enderman)
	assert.True(t, ok)

	// Enderman should have teleported away.
	dist := tr.Position.Distance(originalPos)
	assert.GreaterOrEqual(t, dist, endermanTeleportMin,
		"enderman should teleport at least %f blocks away", endermanTeleportMin)
	assert.LessOrEqual(t, dist, endermanTeleportMax+1,
		"enderman should teleport at most ~%f blocks away", endermanTeleportMax)

	// Health should be reduced.
	h, _ := ecs.GetStore[Health](w).Get(enderman)
	assert.Equal(t, float32(35), h.Current)

	// Damage component should be removed.
	assert.False(t, ecs.GetStore[Damage](w).Has(enderman))
}

func TestNonEndermanDoesNotTeleport(t *testing.T) {
	w := ecs.NewWorld()

	zombie := SpawnZombie(w, mcmath.Vec3{X: 10, Y: 64, Z: 10})
	originalPos := mcmath.Vec3{X: 10, Y: 64, Z: 10}

	ecs.GetStore[Damage](w).Set(zombie, Damage{Amount: 5, Knockback: mcmath.Vec3{}})

	sys := &DamageSystem{Rand: rand.New(rand.NewSource(42))}
	sys.Update(w, 0.0)

	tr, ok := ecs.GetStore[Transform](w).Get(zombie)
	assert.True(t, ok)
	assert.Equal(t, originalPos, tr.Position, "non-enderman should not teleport")
}

// ---------------------------------------------------------------------------
// Enderman AI: aggro when player looks at them
// ---------------------------------------------------------------------------

func TestEndermanAggroOnLook(t *testing.T) {
	w := ecs.NewWorld()

	// Player at origin, looking along +X.
	player := w.NewEntity()
	playerPos := mcmath.Vec3{X: 0, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(player, Transform{Position: playerPos})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Enderman at X=30, directly in the player's look direction.
	enderman := w.NewEntity()
	endermanPos := mcmath.Vec3{X: 30, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(enderman, Transform{Position: endermanPos})
	body := physics.NewBody(endermanBBox)
	body.Position = endermanPos
	ecs.GetStore[PhysicsBody](w).Set(enderman, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(enderman, AI{State: AIIdle, Timer: 10, EndermanTeleport: true})

	sys := &AISystem{
		Rand:          rand.New(rand.NewSource(42)),
		GameTick:      15000, // night
		PlayerEyePos:  playerPos,
		PlayerLookDir: mcmath.Vec3{X: 1, Y: 0, Z: 0}, // looking at enderman
	}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(enderman)
	assert.Equal(t, AIChase, ai.State, "enderman should aggro when looked at")
}

func TestEndermanNoAggroWhenNotLookedAt(t *testing.T) {
	w := ecs.NewWorld()

	// Player at origin, looking along +Z (away from enderman).
	player := w.NewEntity()
	playerPos := mcmath.Vec3{X: 0, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(player, Transform{Position: playerPos})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Enderman at X=30, not in look direction.
	enderman := w.NewEntity()
	endermanPos := mcmath.Vec3{X: 30, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(enderman, Transform{Position: endermanPos})
	body := physics.NewBody(endermanBBox)
	body.Position = endermanPos
	ecs.GetStore[PhysicsBody](w).Set(enderman, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(enderman, AI{State: AIIdle, Timer: 10, EndermanTeleport: true})

	sys := &AISystem{
		Rand:          rand.New(rand.NewSource(42)),
		GameTick:      15000,
		PlayerEyePos:  playerPos,
		PlayerLookDir: mcmath.Vec3{X: 0, Y: 0, Z: 1}, // looking away
	}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(enderman)
	// Enderman is beyond normal chase range (30 > 16) and not looked at.
	assert.NotEqual(t, AIChase, ai.State, "enderman should not aggro when not looked at")
}

func TestEndermanAggroTooFarAway(t *testing.T) {
	w := ecs.NewWorld()

	// Player at origin, looking along +X.
	player := w.NewEntity()
	playerPos := mcmath.Vec3{X: 0, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(player, Transform{Position: playerPos})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Enderman at X=100, too far for look aggro.
	enderman := w.NewEntity()
	endermanPos := mcmath.Vec3{X: 100, Y: 0, Z: 0}
	ecs.GetStore[Transform](w).Set(enderman, Transform{Position: endermanPos})
	body := physics.NewBody(endermanBBox)
	body.Position = endermanPos
	ecs.GetStore[PhysicsBody](w).Set(enderman, PhysicsBody{Body: &body})
	ecs.GetStore[AI](w).Set(enderman, AI{State: AIIdle, Timer: 10, EndermanTeleport: true})

	sys := &AISystem{
		Rand:          rand.New(rand.NewSource(42)),
		GameTick:      15000,
		PlayerEyePos:  playerPos,
		PlayerLookDir: mcmath.Vec3{X: 1, Y: 0, Z: 0},
	}
	sys.Update(w, 0.1)

	ai, _ := ecs.GetStore[AI](w).Get(enderman)
	assert.NotEqual(t, AIChase, ai.State, "enderman should not aggro beyond look range")
}

// ---------------------------------------------------------------------------
// Spider night/day boundary tests
// ---------------------------------------------------------------------------

func TestSpiderNightBoundaryStart(t *testing.T) {
	// Tick 13000 is the start of night.
	sys := &AISystem{GameTick: 13000}
	assert.True(t, sys.isNight(), "tick 13000 should be night")
}

func TestSpiderNightBoundaryEnd(t *testing.T) {
	// Tick 22999 is the last night tick.
	sys := &AISystem{GameTick: 22999}
	assert.True(t, sys.isNight(), "tick 22999 should be night")
}

func TestSpiderDayBoundary(t *testing.T) {
	// Tick 23000 is the start of daytime (wraps to dawn).
	sys := &AISystem{GameTick: 23000}
	assert.False(t, sys.isNight(), "tick 23000 should be day")
}

func TestSpiderDayTick12999(t *testing.T) {
	// Tick 12999 is the last day tick before night.
	sys := &AISystem{GameTick: 12999}
	assert.False(t, sys.isNight(), "tick 12999 should be day")
}

// ---------------------------------------------------------------------------
// Damage type coverage
// ---------------------------------------------------------------------------

func TestSpiderDamageAmount(t *testing.T) {
	w := ecs.NewWorld()
	spider := SpawnSpider(w, mcmath.Vec3{})
	dmg := mobDamageForEntity(w, spider)
	assert.Equal(t, spiderDamage, dmg)
}

func TestEndermanDamageAmount(t *testing.T) {
	w := ecs.NewWorld()
	enderman := SpawnEnderman(w, mcmath.Vec3{})
	dmg := mobDamageForEntity(w, enderman)
	assert.Equal(t, endermanDamage, dmg)
}

// ---------------------------------------------------------------------------
// Spawner inclusion
// ---------------------------------------------------------------------------

func TestSpawnerIncludesSpiderAndEnderman(t *testing.T) {
	// Verify hostile mob types include spider and enderman.
	found := map[string]bool{}
	w := ecs.NewWorld()
	for _, entry := range hostileMobTypes {
		e := entry.Spawn(w, mcmath.Vec3{})
		et, _ := ecs.GetStore[EntityTypeComp](w).Get(e)
		switch et.Type {
		case TypeSpider:
			found["spider"] = true
		case TypeEnderman:
			found["enderman"] = true
		}
	}
	assert.True(t, found["spider"], "spider should be in hostile mob types")
	assert.True(t, found["enderman"], "enderman should be in hostile mob types")
}
