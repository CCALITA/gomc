package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

const (
	// itemDespawnTicks is the tick count after which a dropped item despawns
	// (6000 ticks = 5 minutes at 20 TPS).
	itemDespawnTicks int = 6000

	// itemPickupRangeSq is the squared maximum distance in blocks for a
	// player to pick up a dropped item (1.5^2 = 2.25). Squared to avoid
	// a per-pair sqrt in the hot loop.
	itemPickupRangeSq float32 = 1.5 * 1.5

	// itemDefaultPickupDelay is the number of ticks before a newly spawned
	// item can be picked up.
	itemDefaultPickupDelay int = 10
)

// ItemDrop marks an entity as a dropped item eligible for player pickup.
type ItemDrop struct {
	ItemID      uint16
	Count       int
	PickupDelay int // ticks remaining before pickup is allowed
	AgeTicks    int // ticks since spawn; used for despawn timer
}

// OnPickupFunc is a callback invoked when a player picks up a dropped item.
// It returns true if the pickup was accepted (e.g. inventory had space),
// false otherwise.
type OnPickupFunc func(playerEntity ecs.Entity, itemID uint16, count int) bool

// ItemPickupSystem processes dropped item entities each tick: it decrements
// the pickup delay, ages items toward despawn, and checks proximity to
// players for pickup.
type ItemPickupSystem struct {
	// OnPickup is called when a player is in range and the delay has elapsed.
	// If nil or if it returns false, the item is not consumed.
	OnPickup OnPickupFunc
}

// Update implements ecs.System.
func (s *ItemPickupSystem) Update(w *ecs.World, dt float64) {
	// Collect player positions.
	var players []playerEntry
	ecs.Query2[Transform, EntityTypeComp](w, func(e ecs.Entity, t *Transform, et *EntityTypeComp) {
		if et.Type == TypePlayer {
			players = append(players, playerEntry{entity: e, pos: t.Position})
		}
	})

	var toDestroy []ecs.Entity

	ecs.Query2[ItemDrop, Transform](w, func(e ecs.Entity, drop *ItemDrop, t *Transform) {
		// Age the item toward despawn.
		drop.AgeTicks++
		if drop.AgeTicks >= itemDespawnTicks {
			toDestroy = append(toDestroy, e)
			return
		}

		// Decrement pickup delay.
		if drop.PickupDelay > 0 {
			drop.PickupDelay--
			return
		}

		// Check proximity to players using squared distance to avoid sqrt.
		for _, p := range players {
			distSq := t.Position.DistanceSq(p.pos)
			if distSq < itemPickupRangeSq {
				if s.OnPickup != nil && s.OnPickup(p.entity, drop.ItemID, drop.Count) {
					toDestroy = append(toDestroy, e)
					return
				}
			}
		}
	})

	for _, e := range toDestroy {
		w.DestroyEntity(e)
	}
}

// SpawnItemDrop creates a dropped item entity at the given position.
// The entity has a physics body subject to gravity so the item falls
// and lands on the ground.
func SpawnItemDrop(w *ecs.World, pos mcmath.Vec3, itemID uint16, count int) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(itemBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeItem})

	ecs.GetStore[ItemDrop](w).Set(e, ItemDrop{
		ItemID:      itemID,
		Count:       count,
		PickupDelay: itemDefaultPickupDelay,
	})

	return e
}
