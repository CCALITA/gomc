package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// TryFeed
// ---------------------------------------------------------------------------

func TestTryFeed_CorrectItem(t *testing.T) {
	tests := []struct {
		name    string
		spawn   func(*ecs.World, mcmath.Vec3) ecs.Entity
		itemID  uint16
	}{
		{"cow with wheat", SpawnCow, item.Wheat},
		{"sheep with wheat", SpawnSheep, item.Wheat},
		{"chicken with seeds", SpawnChicken, item.WheatSeeds},
		{"pig with carrot", SpawnPig, item.Carrot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ecs.NewWorld()
			e := tt.spawn(w, mcmath.Vec3{})

			ok := TryFeed(w, e, tt.itemID)
			assert.True(t, ok, "feeding should succeed")

			b, has := ecs.GetStore[Breedable](w).Get(e)
			assert.True(t, has)
			assert.True(t, b.InLove)
			assert.Equal(t, breedLoveDuration, b.LoveTicks)
		})
	}
}

func TestTryFeed_WrongItem(t *testing.T) {
	tests := []struct {
		name    string
		spawn   func(*ecs.World, mcmath.Vec3) ecs.Entity
		itemID  uint16
	}{
		{"cow with seeds", SpawnCow, item.WheatSeeds},
		{"pig with wheat", SpawnPig, item.Wheat},
		{"chicken with carrot", SpawnChicken, item.Carrot},
		{"sheep with carrot", SpawnSheep, item.Carrot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ecs.NewWorld()
			e := tt.spawn(w, mcmath.Vec3{})

			ok := TryFeed(w, e, tt.itemID)
			assert.False(t, ok, "feeding with wrong item should fail")

			b, has := ecs.GetStore[Breedable](w).Get(e)
			assert.True(t, has)
			assert.False(t, b.InLove)
		})
	}
}

func TestTryFeed_BabyCantBeFed(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnCow(w, mcmath.Vec3{})

	b, _ := ecs.GetStore[Breedable](w).Get(e)
	b.Baby = true
	b.GrowthTicks = babyGrowthTicks

	ok := TryFeed(w, e, item.Wheat)
	assert.False(t, ok)
}

func TestTryFeed_CooldownPreventsFeeding(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnCow(w, mcmath.Vec3{})

	b, _ := ecs.GetStore[Breedable](w).Get(e)
	b.CooldownTicks = 100

	ok := TryFeed(w, e, item.Wheat)
	assert.False(t, ok)
}

func TestTryFeed_AlreadyInLove(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnCow(w, mcmath.Vec3{})

	b, _ := ecs.GetStore[Breedable](w).Get(e)
	b.InLove = true
	b.LoveTicks = 100

	ok := TryFeed(w, e, item.Wheat)
	assert.False(t, ok)
}

func TestTryFeed_NonBreedableEntity(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnZombie(w, mcmath.Vec3{})

	ok := TryFeed(w, e, item.Wheat)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Love mode pairing and baby spawning
// ---------------------------------------------------------------------------

func TestBreedingSystem_PairingSpawnsBaby(t *testing.T) {
	w := ecs.NewWorld()

	cow1 := SpawnCow(w, mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cow2 := SpawnCow(w, mcmath.Vec3{X: 2, Y: 64, Z: 0})

	// Put both in love mode.
	b1, _ := ecs.GetStore[Breedable](w).Get(cow1)
	b1.InLove = true
	b1.LoveTicks = breedLoveDuration

	b2, _ := ecs.GetStore[Breedable](w).Get(cow2)
	b2.InLove = true
	b2.LoveTicks = breedLoveDuration

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	// Parents should no longer be in love and should have cooldown.
	b1, _ = ecs.GetStore[Breedable](w).Get(cow1)
	assert.False(t, b1.InLove)
	assert.Equal(t, breedCooldownTicks, b1.CooldownTicks)

	b2, _ = ecs.GetStore[Breedable](w).Get(cow2)
	assert.False(t, b2.InLove)
	assert.Equal(t, breedCooldownTicks, b2.CooldownTicks)

	// A baby should have been spawned.
	babyCount := 0
	ecs.GetStore[Breedable](w).Each(func(e ecs.Entity, b *Breedable) {
		if b.Baby {
			babyCount++
			assert.Equal(t, babyGrowthTicks, b.GrowthTicks)
			assert.Equal(t, babyScale, b.Scale)

			// Baby should be a cow.
			et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
			assert.True(t, ok)
			assert.Equal(t, TypeCow, et.Type)

			// Baby should be at midpoint.
			tr, ok := ecs.GetStore[Transform](w).Get(e)
			assert.True(t, ok)
			assert.InDelta(t, 1.0, float64(tr.Position.X), 0.01)
		}
	})
	assert.Equal(t, 1, babyCount)
}

func TestBreedingSystem_DifferentTypesDoNotPair(t *testing.T) {
	w := ecs.NewWorld()

	cow := SpawnCow(w, mcmath.Vec3{X: 0, Y: 64, Z: 0})
	pig := SpawnPig(w, mcmath.Vec3{X: 1, Y: 64, Z: 0})

	bc, _ := ecs.GetStore[Breedable](w).Get(cow)
	bc.InLove = true
	bc.LoveTicks = breedLoveDuration

	bp, _ := ecs.GetStore[Breedable](w).Get(pig)
	bp.InLove = true
	bp.LoveTicks = breedLoveDuration

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	// Both should still be in love (no pairing happened).
	bc, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.True(t, bc.InLove)

	bp, _ = ecs.GetStore[Breedable](w).Get(pig)
	assert.True(t, bp.InLove)
}

func TestBreedingSystem_TooFarToBreed(t *testing.T) {
	w := ecs.NewWorld()

	cow1 := SpawnCow(w, mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cow2 := SpawnCow(w, mcmath.Vec3{X: 10, Y: 64, Z: 0}) // far apart

	b1, _ := ecs.GetStore[Breedable](w).Get(cow1)
	b1.InLove = true
	b1.LoveTicks = breedLoveDuration

	b2, _ := ecs.GetStore[Breedable](w).Get(cow2)
	b2.InLove = true
	b2.LoveTicks = breedLoveDuration

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	// Both should still be in love (too far apart).
	b1, _ = ecs.GetStore[Breedable](w).Get(cow1)
	assert.True(t, b1.InLove)

	b2, _ = ecs.GetStore[Breedable](w).Get(cow2)
	assert.True(t, b2.InLove)
}

// ---------------------------------------------------------------------------
// Growth timer
// ---------------------------------------------------------------------------

func TestBreedingSystem_BabyGrowsUp(t *testing.T) {
	w := ecs.NewWorld()

	cow := SpawnCow(w, mcmath.Vec3{})
	b, _ := ecs.GetStore[Breedable](w).Get(cow)
	b.Baby = true
	b.GrowthTicks = 2
	b.Scale = babyScale

	sys := &BreedingSystem{}

	// Tick 1: still a baby.
	sys.Update(w, 0.05)
	b, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.True(t, b.Baby)
	assert.Equal(t, 1, b.GrowthTicks)

	// Tick 2: grows up.
	sys.Update(w, 0.05)
	b, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.False(t, b.Baby)
	assert.Equal(t, float32(1.0), b.Scale)
}

// ---------------------------------------------------------------------------
// Cooldown prevents re-breeding
// ---------------------------------------------------------------------------

func TestBreedingSystem_CooldownPreventsReBreeding(t *testing.T) {
	w := ecs.NewWorld()

	cow1 := SpawnCow(w, mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cow2 := SpawnCow(w, mcmath.Vec3{X: 1, Y: 64, Z: 0})

	// Breed them once.
	b1, _ := ecs.GetStore[Breedable](w).Get(cow1)
	b1.InLove = true
	b1.LoveTicks = breedLoveDuration

	b2, _ := ecs.GetStore[Breedable](w).Get(cow2)
	b2.InLove = true
	b2.LoveTicks = breedLoveDuration

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	// Verify they bred (cooldown set).
	b1, _ = ecs.GetStore[Breedable](w).Get(cow1)
	assert.Equal(t, breedCooldownTicks, b1.CooldownTicks)

	// Try to feed them again -- should fail due to cooldown.
	ok1 := TryFeed(w, cow1, item.Wheat)
	ok2 := TryFeed(w, cow2, item.Wheat)
	assert.False(t, ok1, "should not feed during cooldown")
	assert.False(t, ok2, "should not feed during cooldown")
}

// ---------------------------------------------------------------------------
// Love mode expires
// ---------------------------------------------------------------------------

func TestBreedingSystem_LoveModeExpires(t *testing.T) {
	w := ecs.NewWorld()
	cow := SpawnCow(w, mcmath.Vec3{})

	b, _ := ecs.GetStore[Breedable](w).Get(cow)
	b.InLove = true
	b.LoveTicks = 2

	sys := &BreedingSystem{}

	// Tick 1: still in love (LoveTicks decremented to 1).
	sys.Update(w, 0.05)
	b, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.True(t, b.InLove)

	// Tick 2: love expired (LoveTicks decremented to 0).
	sys.Update(w, 0.05)
	b, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.False(t, b.InLove)
}

// ---------------------------------------------------------------------------
// Factory functions include Breedable
// ---------------------------------------------------------------------------

func TestSpawnAnimalHasBreedable(t *testing.T) {
	tests := []struct {
		name  string
		spawn func(*ecs.World, mcmath.Vec3) ecs.Entity
	}{
		{"cow", SpawnCow},
		{"pig", SpawnPig},
		{"sheep", SpawnSheep},
		{"chicken", SpawnChicken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ecs.NewWorld()
			e := tt.spawn(w, mcmath.Vec3{})

			b, ok := ecs.GetStore[Breedable](w).Get(e)
			assert.True(t, ok, "spawned animal should have Breedable component")
			assert.False(t, b.Baby)
			assert.False(t, b.InLove)
			assert.Equal(t, float32(1.0), b.Scale)
		})
	}
}

// ---------------------------------------------------------------------------
// Cooldown decrement
// ---------------------------------------------------------------------------

func TestBreedingSystem_CooldownDecrements(t *testing.T) {
	w := ecs.NewWorld()
	cow := SpawnCow(w, mcmath.Vec3{})

	b, _ := ecs.GetStore[Breedable](w).Get(cow)
	b.CooldownTicks = 3

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	b, _ = ecs.GetStore[Breedable](w).Get(cow)
	assert.Equal(t, 2, b.CooldownTicks)
}

// ---------------------------------------------------------------------------
// Multiple species breed independently
// ---------------------------------------------------------------------------

func TestBreedingSystem_MultipleSpeciesBreedIndependently(t *testing.T) {
	w := ecs.NewWorld()

	cow1 := SpawnCow(w, mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cow2 := SpawnCow(w, mcmath.Vec3{X: 1, Y: 64, Z: 0})
	pig1 := SpawnPig(w, mcmath.Vec3{X: 0, Y: 64, Z: 10})
	pig2 := SpawnPig(w, mcmath.Vec3{X: 1, Y: 64, Z: 10})

	for _, e := range []ecs.Entity{cow1, cow2, pig1, pig2} {
		b, _ := ecs.GetStore[Breedable](w).Get(e)
		b.InLove = true
		b.LoveTicks = breedLoveDuration
	}

	sys := &BreedingSystem{}
	sys.Update(w, 0.05)

	// Both pairs should have bred -- count babies.
	babyCount := 0
	ecs.GetStore[Breedable](w).Each(func(e ecs.Entity, b *Breedable) {
		if b.Baby {
			babyCount++
		}
	})
	assert.Equal(t, 2, babyCount, "one cow baby and one pig baby")
}
