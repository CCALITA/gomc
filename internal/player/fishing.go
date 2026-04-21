package player

import (
	"math/rand"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Fishing wait-tick bounds (100-600 ticks, ~5-30 seconds at 20 tps).
const (
	minWaitTicks = 100
	maxWaitTicks = 600
)

// Loot probability thresholds (out of 1000 for integer math).
const (
	fishThreshold = 850 // 0-849 = fish (85%)
	junkThreshold = 950 // 850-949 = junk (10%)
	// 950-999 = treasure (5%)
)

// Fish loot table.
var fishLoot = []item.ItemID{item.Cod, item.Salmon}

// Junk loot table.
var junkLoot = []item.ItemID{item.Stick, item.StringItem, item.Bowl}

// Treasure loot table.
var treasureLoot = []item.ItemID{item.Bow, item.Book, item.Saddle}

// FishingState holds the state of an active fishing session.
type FishingState struct {
	Active       bool
	BobberEntity ecs.Entity
	WaitTicks    int
	CatchReady   bool
}

// Cast spawns a bobber entity and begins a fishing session. The bobber is
// launched with a projectile arc from origin along direction.
func (fs *FishingState) Cast(w *ecs.World, rng *rand.Rand, origin, direction mcmath.Vec3) {
	velocity := direction.Scale(12.0)
	velocity.Y = 8.0

	bobber := entity.SpawnFishingBobber(w, origin, velocity)

	fs.Active = true
	fs.BobberEntity = bobber
	fs.CatchReady = false
	fs.WaitTicks = minWaitTicks + rng.Intn(maxWaitTicks-minWaitTicks+1)
}

// Reel attempts to reel in the fishing line. If CatchReady is true, a random
// loot item is returned and the bobber is destroyed. Returns the caught item
// stack and whether a catch was made.
func (fs *FishingState) Reel(w *ecs.World, rng *rand.Rand) (item.ItemStack, bool) {
	if !fs.Active {
		return item.ItemStack{}, false
	}

	defer fs.reset(w)

	if !fs.CatchReady {
		return item.ItemStack{}, false
	}

	lootID := rollLoot(rng)
	return item.NewItemStack(lootID, 1), true
}

// UpdateBobber decrements the wait timer each tick and sets CatchReady
// when the timer elapses.
func (fs *FishingState) UpdateBobber(w *ecs.World, dt float64) {
	if !fs.Active || fs.CatchReady {
		return
	}

	fs.WaitTicks--
	if fs.WaitTicks <= 0 {
		fs.CatchReady = true
	}
}

// reset destroys the bobber and clears the fishing state.
func (fs *FishingState) reset(w *ecs.World) {
	if fs.Active {
		w.DestroyEntity(fs.BobberEntity)
	}
	fs.Active = false
	fs.BobberEntity = 0
	fs.WaitTicks = 0
	fs.CatchReady = false
}

// rollLoot picks a random item from the loot tables using the probability
// distribution: 85% fish, 10% junk, 5% treasure.
func rollLoot(rng *rand.Rand) item.ItemID {
	roll := rng.Intn(1000)
	switch {
	case roll < fishThreshold:
		return fishLoot[rng.Intn(len(fishLoot))]
	case roll < junkThreshold:
		return junkLoot[rng.Intn(len(junkLoot))]
	default:
		return treasureLoot[rng.Intn(len(treasureLoot))]
	}
}
