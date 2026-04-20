package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// SpawnItemDrop
// ---------------------------------------------------------------------------

func TestSpawnItemDrop_Position(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 65, Z: 10}
	e := SpawnItemDrop(w, pos, 1, 1)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, TypeItem, et.Type)

	drop, ok := ecs.GetStore[ItemDrop](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, uint16(1), drop.ItemID)
	assert.Equal(t, 1, drop.Count)
	assert.Equal(t, itemDefaultPickupDelay, drop.PickupDelay)
	assert.Equal(t, 0, drop.AgeTicks)
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — pickup delay prevents early pickup
// ---------------------------------------------------------------------------

func TestItemPickupSystem_DelayPreventsPickup(t *testing.T) {
	w := ecs.NewWorld()

	// Player at origin.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Item right next to player but with pickup delay.
	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 0.5, Y: 0, Z: 0}, 42, 1)

	pickupCalled := false
	sys := &ItemPickupSystem{
		OnPickup: func(_ ecs.Entity, _ uint16, _ int) bool {
			pickupCalled = true
			return true
		},
	}

	// Run one tick — delay should prevent pickup.
	sys.Update(w, 0.05)

	assert.False(t, pickupCalled, "pickup should not fire while delay is active")
	assert.True(t, w.Alive(itemE))
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — pickup within range works after delay expires
// ---------------------------------------------------------------------------

func TestItemPickupSystem_PickupAfterDelay(t *testing.T) {
	w := ecs.NewWorld()

	// Player at origin.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Item close to player.
	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 0.5, Y: 0, Z: 0}, 42, 3)

	// Manually set pickup delay to 0 to simulate it having expired.
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.PickupDelay = 0

	var pickedPlayer ecs.Entity
	var pickedItemID uint16
	var pickedCount int

	sys := &ItemPickupSystem{
		OnPickup: func(pe ecs.Entity, id uint16, c int) bool {
			pickedPlayer = pe
			pickedItemID = id
			pickedCount = c
			return true
		},
	}

	sys.Update(w, 0.05)

	assert.Equal(t, player, pickedPlayer)
	assert.Equal(t, uint16(42), pickedItemID)
	assert.Equal(t, 3, pickedCount)
	assert.False(t, w.Alive(itemE), "item should be destroyed after successful pickup")
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — out of range does not trigger pickup
// ---------------------------------------------------------------------------

func TestItemPickupSystem_OutOfRange(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	// Item far away.
	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 10, Y: 0, Z: 0}, 42, 1)
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.PickupDelay = 0

	pickupCalled := false
	sys := &ItemPickupSystem{
		OnPickup: func(_ ecs.Entity, _ uint16, _ int) bool {
			pickupCalled = true
			return true
		},
	}

	sys.Update(w, 0.05)

	assert.False(t, pickupCalled)
	assert.True(t, w.Alive(itemE))
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — despawn after 5 minutes (6000 ticks)
// ---------------------------------------------------------------------------

func TestItemPickupSystem_DespawnAfter5Minutes(t *testing.T) {
	w := ecs.NewWorld()

	// No players needed for despawn test.
	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 5, Y: 64, Z: 5}, 1, 1)

	// Fast-forward the age to just below threshold.
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.AgeTicks = itemDespawnTicks - 1
	drop.PickupDelay = 0

	sys := &ItemPickupSystem{}

	// One more tick pushes age to the threshold.
	sys.Update(w, 0.05)

	assert.False(t, w.Alive(itemE), "item should despawn after reaching 6000 ticks")
}

func TestItemPickupSystem_NotDespawnedBefore5Minutes(t *testing.T) {
	w := ecs.NewWorld()

	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 5, Y: 64, Z: 5}, 1, 1)

	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.AgeTicks = itemDespawnTicks - 2
	drop.PickupDelay = 0

	sys := &ItemPickupSystem{}
	sys.Update(w, 0.05)

	assert.True(t, w.Alive(itemE), "item should still be alive before 6000 ticks")
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — full inventory prevents pickup
// ---------------------------------------------------------------------------

func TestItemPickupSystem_FullInventoryPreventsPickup(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 0.5, Y: 0, Z: 0}, 42, 1)
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.PickupDelay = 0

	sys := &ItemPickupSystem{
		OnPickup: func(_ ecs.Entity, _ uint16, _ int) bool {
			return false // inventory full
		},
	}

	sys.Update(w, 0.05)

	assert.True(t, w.Alive(itemE), "item should remain when inventory rejects pickup")
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — nil OnPickup does not crash
// ---------------------------------------------------------------------------

func TestItemPickupSystem_NilOnPickup(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})

	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 0.5, Y: 0, Z: 0}, 42, 1)
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	drop.PickupDelay = 0

	sys := &ItemPickupSystem{}
	sys.Update(w, 0.05)

	// Item should remain since no callback accepted it.
	assert.True(t, w.Alive(itemE))
}

// ---------------------------------------------------------------------------
// ItemPickupSystem — delay decrements each tick
// ---------------------------------------------------------------------------

func TestItemPickupSystem_DelayDecrementsEachTick(t *testing.T) {
	w := ecs.NewWorld()

	itemE := SpawnItemDrop(w, mcmath.Vec3{X: 5, Y: 64, Z: 5}, 1, 1)
	drop, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
	initialDelay := drop.PickupDelay

	sys := &ItemPickupSystem{}

	for i := range initialDelay {
		sys.Update(w, 0.05)
		d, _ := ecs.GetStore[ItemDrop](w).Get(itemE)
		assert.Equal(t, initialDelay-1-i, d.PickupDelay,
			"delay should decrement by 1 each tick")
	}
}
