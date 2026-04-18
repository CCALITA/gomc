//go:build !ci

package player

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
	"github.com/fanxiyao/gomc/internal/world"
	"github.com/stretchr/testify/assert"
)

// newTestWorld creates a minimal world with the given seed for testing.
func newTestWorld(seed int64) *world.World {
	return world.NewWorld(seed)
}

// newTestBody creates a physics body at the given position.
func newTestBody(pos mcmath.Vec3) *physics.Body {
	bbox := mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	}
	b := physics.NewBody(bbox)
	b.Position = pos
	return &b
}

// sumDamageByType returns the total damage for events matching the given type.
func sumDamageByType(events []DamageEvent, damageType string) float32 {
	var total float32
	for _, e := range events {
		if e.Type == damageType {
			total += e.Amount
		}
	}
	return total
}

func TestNewPlayerDamageTracker(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	assert.Equal(t, float32(0), tracker.FallDistance)
	assert.Equal(t, float32(0), tracker.LavaContactTime)
	assert.Equal(t, float32(0), tracker.UnderwaterTime)
	assert.Equal(t, float32(0), tracker.FireTime)
}

func TestFallDamage_NoDamageBelowThreshold(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	// Simulate a 2-block fall (below the 3-block threshold).
	body.OnGround = false
	body.Velocity.Y = -10 // 10 blocks/s downward
	dt := float32(0.2)    // 0.2 seconds = 2 blocks

	events := tracker.Update(body, w, pos, dt)
	assert.Empty(t, events)
	assert.InDelta(t, 2.0, tracker.FallDistance, 0.01)

	// Player lands.
	body.OnGround = true
	body.Velocity.Y = 0
	events = tracker.Update(body, w, pos, dt)
	assert.Empty(t, events, "no damage for a 2-block fall")
}

func TestFallDamage_SevenDamageFromTenBlockFall(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	// Simulate a 10-block fall.
	body.OnGround = false
	body.Velocity.Y = -20 // 20 blocks/s
	dt := float32(0.5)    // 0.5 seconds = 10 blocks

	events := tracker.Update(body, w, pos, dt)
	assert.Empty(t, events, "no damage while still falling")
	assert.InDelta(t, 10.0, tracker.FallDistance, 0.01)

	// Land.
	body.OnGround = true
	body.Velocity.Y = 0
	events = tracker.Update(body, w, pos, dt)

	assert.Len(t, events, 1)
	assert.Equal(t, DamageTypeFall, events[0].Type)
	assert.InDelta(t, 7.0, events[0].Amount, 0.01)
}

func TestFallDamage_FallDistanceResetsOnLanding(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	// Accumulate fall distance.
	body.OnGround = false
	body.Velocity.Y = -20
	tracker.Update(body, w, pos, 0.5) // 10 blocks

	// Land.
	body.OnGround = true
	body.Velocity.Y = 0
	tracker.Update(body, w, pos, 0.05)

	assert.Equal(t, float32(0), tracker.FallDistance)
}

func TestFallDamage_NoDamageExactlyAtThreshold(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	// Exactly 3 blocks = threshold, no damage.
	body.OnGround = false
	body.Velocity.Y = -15
	tracker.Update(body, w, pos, 0.2) // 3 blocks

	body.OnGround = true
	body.Velocity.Y = 0
	events := tracker.Update(body, w, pos, 0.05)
	assert.Empty(t, events)
}

func TestFallDamage_NoAccumulationWhileRising(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	// Player moving upward should not accumulate fall distance.
	body.OnGround = false
	body.Velocity.Y = 10
	tracker.Update(body, w, pos, 0.5)

	assert.Equal(t, float32(0), tracker.FallDistance)
}

func TestLavaDamage(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Lava)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	dt := float32(1.0)

	events := tracker.Update(body, w, pos, dt)

	lavaDmg := sumDamageByType(events, DamageTypeLava)
	assert.InDelta(t, 4.0, lavaDmg, 0.01)
}

func TestLavaDamage_ResetWhenNotInLava(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Lava)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	tracker.Update(body, w, pos, 0.5)
	assert.Greater(t, tracker.LavaContactTime, float32(0))

	w.SetBlock(feetPos, block.Air)
	tracker.Update(body, w, pos, 0.5)
	assert.Equal(t, float32(0), tracker.LavaContactTime)
}

func TestDrowning_NoDamageBeforeBreathHold(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	headPos := mcmath.Vec3{X: 0, Y: 64 + EyeOffset, Z: 0}
	headBlock := headPos.Floor()
	cp := headBlock.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(headBlock, block.Water)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	events := tracker.Update(body, w, pos, 10.0)

	drownDmg := sumDamageByType(events, DamageTypeDrown)
	assert.Equal(t, float32(0), drownDmg)
	assert.InDelta(t, 10.0, tracker.UnderwaterTime, 0.01)
}

func TestDrowning_DamageAfterBreathHold(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	headPos := mcmath.Vec3{X: 0, Y: 64 + EyeOffset, Z: 0}
	headBlock := headPos.Floor()
	cp := headBlock.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(headBlock, block.Water)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	tracker.Update(body, w, pos, 15.0)
	events := tracker.Update(body, w, pos, 1.0)

	drownDmg := sumDamageByType(events, DamageTypeDrown)
	assert.InDelta(t, 2.0, drownDmg, 0.01)
}

func TestDrowning_NoWhenHeadAboveWater(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Water)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	events := tracker.Update(body, w, pos, 30.0)

	drownDmg := sumDamageByType(events, DamageTypeDrown)
	assert.Equal(t, float32(0), drownDmg)
}

func TestDrowning_ResetOnSurface(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	headPos := mcmath.Vec3{X: 0, Y: 64 + EyeOffset, Z: 0}
	headBlock := headPos.Floor()
	cp := headBlock.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(headBlock, block.Water)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	tracker.Update(body, w, pos, 10.0)
	assert.InDelta(t, 10.0, tracker.UnderwaterTime, 0.01)

	w.SetBlock(headBlock, block.Air)
	tracker.Update(body, w, pos, 0.05)

	assert.Equal(t, float32(0), tracker.UnderwaterTime)
}

func TestFireDamage(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Fire)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	dt := float32(1.0)

	events := tracker.Update(body, w, pos, dt)

	fireDmg := sumDamageByType(events, DamageTypeFire)
	assert.InDelta(t, 1.0, fireDmg, 0.01)
}

func TestFireDamage_ResetOffFire(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Fire)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	tracker.Update(body, w, pos, 1.0)
	assert.Greater(t, tracker.FireTime, float32(0))

	w.SetBlock(feetPos, block.Air)
	tracker.Update(body, w, pos, 0.05)
	assert.Equal(t, float32(0), tracker.FireTime)
}

func TestNoDamageOnSafeGround(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	events := tracker.Update(body, w, pos, 1.0)
	assert.Empty(t, events)
}

func TestMultipleDamageSources(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	body.OnGround = true

	feetPos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	cp := feetPos.ToChunkPos()
	w.LoadChunk(cp)
	w.SetBlock(feetPos, block.Lava)

	headPos := mcmath.Vec3{X: 0, Y: 64 + EyeOffset, Z: 0}
	headBlock := headPos.Floor()
	hcp := headBlock.ToChunkPos()
	w.LoadChunk(hcp)
	w.SetBlock(headBlock, block.Water)

	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	tracker.Update(body, w, pos, 15.0)
	events := tracker.Update(body, w, pos, 1.0)

	assert.InDelta(t, 4.0, sumDamageByType(events, DamageTypeLava), 0.01)
	assert.InDelta(t, 2.0, sumDamageByType(events, DamageTypeDrown), 0.01)
}

func TestFallDamage_AccumulatesAcrossMultipleTicks(t *testing.T) {
	tracker := NewPlayerDamageTracker()
	w := newTestWorld(42)
	body := newTestBody(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pos := mcmath.Vec3{X: 0, Y: 64, Z: 0}

	body.OnGround = false
	body.Velocity.Y = -10

	// Fall across 3 ticks: 2 + 2 + 2 = 6 blocks.
	tracker.Update(body, w, pos, 0.2)
	tracker.Update(body, w, pos, 0.2)
	tracker.Update(body, w, pos, 0.2)

	assert.InDelta(t, 6.0, tracker.FallDistance, 0.01)

	// Land.
	body.OnGround = true
	body.Velocity.Y = 0
	events := tracker.Update(body, w, pos, 0.05)

	assert.Len(t, events, 1)
	assert.Equal(t, DamageTypeFall, events[0].Type)
	assert.InDelta(t, 3.0, events[0].Amount, 0.01) // 6 - 3 = 3
}
