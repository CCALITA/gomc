//go:build !ci

package player

import (
	"math/rand"
	"testing"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRNG() *rand.Rand {
	return rand.New(rand.NewSource(42))
}

func TestCast_CreatesBobber(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	origin := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	dir := mcmath.Vec3{X: 0, Y: 0, Z: -1}

	fs.Cast(w, rng, origin, dir)

	assert.True(t, fs.Active)
	assert.True(t, w.Alive(fs.BobberEntity))
	assert.False(t, fs.CatchReady)
	assert.Greater(t, fs.WaitTicks, 0)
	assert.GreaterOrEqual(t, fs.WaitTicks, minWaitTicks)
	assert.LessOrEqual(t, fs.WaitTicks, maxWaitTicks)
}

func TestReel_WithoutCast_ReturnsFalse(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	stack, ok := fs.Reel(w, rng)
	assert.False(t, ok)
	assert.True(t, stack.IsEmpty())
}

func TestReel_BeforeCatchReady_ReturnsFalse(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	fs.Cast(w, rng, mcmath.Vec3{}, mcmath.Vec3{X: 0, Y: 0, Z: -1})
	require.True(t, fs.Active)
	require.False(t, fs.CatchReady)

	stack, ok := fs.Reel(w, rng)
	assert.False(t, ok)
	assert.True(t, stack.IsEmpty())
	assert.False(t, fs.Active, "fishing state should be reset after reel")
}

func TestWaitTimer_SetsCatchReady(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	fs.Cast(w, rng, mcmath.Vec3{}, mcmath.Vec3{X: 0, Y: 0, Z: -1})
	ticks := fs.WaitTicks

	// Tick down all but one.
	for i := 0; i < ticks-1; i++ {
		fs.UpdateBobber(w, 0.05)
	}
	assert.False(t, fs.CatchReady)

	// Final tick.
	fs.UpdateBobber(w, 0.05)
	assert.True(t, fs.CatchReady)
}

func TestReel_WhenCatchReady_ReturnsItem(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	fs.Cast(w, rng, mcmath.Vec3{}, mcmath.Vec3{X: 0, Y: 0, Z: -1})
	bobber := fs.BobberEntity

	// Force catch ready.
	fs.WaitTicks = 0
	fs.UpdateBobber(w, 0.05)
	require.True(t, fs.CatchReady)

	stack, ok := fs.Reel(w, rng)
	assert.True(t, ok)
	assert.Equal(t, 1, stack.Count)
	assert.False(t, fs.Active)
	assert.False(t, w.Alive(bobber), "bobber should be destroyed on reel")
}

func TestBobber_DespawnOnReel(t *testing.T) {
	w := ecs.NewWorld()
	rng := newTestRNG()
	fs := &FishingState{}

	fs.Cast(w, rng, mcmath.Vec3{}, mcmath.Vec3{X: 0, Y: 0, Z: -1})
	bobber := fs.BobberEntity
	require.True(t, w.Alive(bobber))

	// Reel without catch.
	fs.Reel(w, rng)
	assert.False(t, w.Alive(bobber))
}

func TestLootDistribution(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	const trials = 10000

	fishCount := 0
	junkCount := 0
	treasureCount := 0

	fishSet := map[item.ItemID]bool{item.Cod: true, item.Salmon: true}
	junkSet := map[item.ItemID]bool{item.Stick: true, item.StringItem: true, item.Bowl: true}
	treasureSet := map[item.ItemID]bool{item.Bow: true, item.Book: true, item.Saddle: true}

	for i := 0; i < trials; i++ {
		id := rollLoot(rng)
		switch {
		case fishSet[id]:
			fishCount++
		case junkSet[id]:
			junkCount++
		case treasureSet[id]:
			treasureCount++
		default:
			t.Fatalf("unexpected loot item: %d", id)
		}
	}

	fishPct := float64(fishCount) / float64(trials) * 100
	junkPct := float64(junkCount) / float64(trials) * 100
	treasurePct := float64(treasureCount) / float64(trials) * 100

	// Allow +/- 3% tolerance.
	assert.InDelta(t, 85.0, fishPct, 3.0, "fish should be ~85%%, got %.1f%%", fishPct)
	assert.InDelta(t, 10.0, junkPct, 3.0, "junk should be ~10%%, got %.1f%%", junkPct)
	assert.InDelta(t, 5.0, treasurePct, 3.0, "treasure should be ~5%%, got %.1f%%", treasurePct)
}
