package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Breeding timing and range constants.
const (
	// breedLoveDuration is how long (in ticks) an animal stays in love mode
	// after being fed. At 20 TPS this is 30 seconds.
	breedLoveDuration = 600

	// breedCooldownTicks is the cooldown after breeding before the animal
	// can enter love mode again (5 minutes at 20 TPS).
	breedCooldownTicks = 6000

	// breedRange is the maximum distance (in blocks) between two animals
	// for them to pair up for breeding.
	breedRange float32 = 3.0

	// babyGrowthTicks is the time (in ticks) for a baby to mature into an
	// adult (20 minutes at 20 TPS).
	babyGrowthTicks = 24000

	// babyScale is the visual scale multiplier applied to baby entities.
	babyScale float32 = 0.5
)

// Breedable is an ECS component that tracks breeding state for passive mobs.
type Breedable struct {
	InLove        bool
	LoveTicks     int
	CooldownTicks int
	Baby          bool
	GrowthTicks   int
	Scale         float32
}

// ---------------------------------------------------------------------------
// BreedingSystem
// ---------------------------------------------------------------------------

// BreedingSystem processes love mode, mate pairing, baby spawning, and growth
// for entities with Breedable, Transform, and EntityTypeComp components.
type BreedingSystem struct{}

// breedCandidate holds snapshot data from an in-love adult entity.
type breedCandidate struct {
	entity   ecs.Entity
	mobType  uint8
	position mcmath.Vec3
}

// Update runs one tick of the breeding simulation.
func (s *BreedingSystem) Update(w *ecs.World, dt float64) {
	breedStore := ecs.GetStore[Breedable](w)
	transformStore := ecs.GetStore[Transform](w)
	etStore := ecs.GetStore[EntityTypeComp](w)

	// --- Phase 1: tick timers and collect love-mode candidates. ---
	var candidates []breedCandidate

	breedStore.Each(func(e ecs.Entity, b *Breedable) {
		// Tick cooldown.
		if b.CooldownTicks > 0 {
			b.CooldownTicks--
		}

		// Tick baby growth.
		if b.Baby {
			b.GrowthTicks--
			if b.GrowthTicks <= 0 {
				b.Baby = false
				b.Scale = 1.0
			}
			return // babies cannot breed
		}

		// Tick love mode.
		if b.InLove {
			b.LoveTicks--
			if b.LoveTicks <= 0 {
				b.InLove = false
			}
		}

		// Collect still-in-love adults for pairing.
		if !b.InLove {
			return
		}
		t, tok := transformStore.Get(e)
		et, etok := etStore.Get(e)
		if !tok || !etok {
			return
		}
		candidates = append(candidates, breedCandidate{
			entity:   e,
			mobType:  et.Type,
			position: t.Position,
		})
	})

	// --- Phase 2: pair up candidates of the same type within range. ---
	if len(candidates) < 2 {
		return
	}

	breedRangeSq := breedRange * breedRange
	paired := make(map[ecs.Entity]bool)

	for i := 0; i < len(candidates); i++ {
		a := candidates[i]
		if paired[a.entity] {
			continue
		}
		for j := i + 1; j < len(candidates); j++ {
			b := candidates[j]
			if paired[b.entity] {
				continue
			}
			if a.mobType != b.mobType {
				continue
			}
			if a.position.DistanceSq(b.position) > breedRangeSq {
				continue
			}

			// Mate found -- mark both as paired.
			paired[a.entity] = true
			paired[b.entity] = true

			// Reset parents.
			ba, _ := breedStore.Get(a.entity)
			ba.InLove = false
			ba.LoveTicks = 0
			ba.CooldownTicks = breedCooldownTicks

			bb, _ := breedStore.Get(b.entity)
			bb.InLove = false
			bb.LoveTicks = 0
			bb.CooldownTicks = breedCooldownTicks

			// Spawn baby at midpoint.
			midpoint := a.position.Lerp(b.position, 0.5)
			spawnBaby(w, a.mobType, midpoint)

			break // a is done, move to next i
		}
	}
}

// spawnBaby creates a baby entity of the given mob type at the given position.
func spawnBaby(w *ecs.World, mobType uint8, pos mcmath.Vec3) ecs.Entity {
	var baby ecs.Entity
	switch mobType {
	case TypeCow:
		baby = SpawnCow(w, pos)
	case TypePig:
		baby = SpawnPig(w, pos)
	case TypeSheep:
		baby = SpawnSheep(w, pos)
	case TypeChicken:
		baby = SpawnChicken(w, pos)
	default:
		return 0
	}

	b, ok := ecs.GetStore[Breedable](w).Get(baby)
	if ok {
		b.Baby = true
		b.GrowthTicks = babyGrowthTicks
		b.Scale = babyScale
	}
	return baby
}

// ---------------------------------------------------------------------------
// TryFeed
// ---------------------------------------------------------------------------

// breedingFoods maps entity types to the item IDs that trigger love mode.
var breedingFoods = map[uint8]uint16{
	TypeCow:     item.Wheat,
	TypeSheep:   item.Wheat,
	TypeChicken: item.WheatSeeds,
	TypePig:     item.Carrot,
}

// TryFeed attempts to feed the given entity with the specified item. If the
// item matches the entity's breeding food and the entity is an adult not on
// cooldown, it enters love mode and returns true. Otherwise it returns false.
func TryFeed(w *ecs.World, entity ecs.Entity, itemID uint16) bool {
	et, ok := ecs.GetStore[EntityTypeComp](w).Get(entity)
	if !ok {
		return false
	}

	wantedItem, ok := breedingFoods[et.Type]
	if !ok || itemID != wantedItem {
		return false
	}

	b, ok := ecs.GetStore[Breedable](w).Get(entity)
	if !ok {
		return false
	}

	// Babies and entities on cooldown or already in love cannot be fed.
	if b.Baby || b.CooldownTicks > 0 || b.InLove {
		return false
	}

	b.InLove = true
	b.LoveTicks = breedLoveDuration
	return true
}
