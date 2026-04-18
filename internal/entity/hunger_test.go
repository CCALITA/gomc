package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// Hunger component defaults
// ---------------------------------------------------------------------------

func TestNewHunger_Defaults(t *testing.T) {
	h := NewHunger()
	assert.Equal(t, DefaultFoodLevel, h.FoodLevel)
	assert.Equal(t, DefaultSaturation, h.Saturation)
	assert.Equal(t, 0.0, h.Exhaustion)
	assert.Equal(t, 0.0, h.StarvationTimer)
	assert.Equal(t, 0.0, h.RegenTimer)
}

// ---------------------------------------------------------------------------
// Exhaustion accumulation
// ---------------------------------------------------------------------------

func TestAddExhaustion_BelowThreshold(t *testing.T) {
	h := NewHunger()
	h.AddExhaustion(2.0)
	assert.Equal(t, 2.0, h.Exhaustion)
	assert.Equal(t, DefaultFoodLevel, h.FoodLevel)
	assert.Equal(t, DefaultSaturation, h.Saturation)
}

func TestAddExhaustion_DrainsSaturation(t *testing.T) {
	h := NewHunger()
	h.Saturation = 3.0
	h.AddExhaustion(ExhaustionThreshold)
	// Exhaustion should reset, saturation should decrease by 1.
	assert.InDelta(t, 0.0, h.Exhaustion, 0.001)
	assert.InDelta(t, 2.0, h.Saturation, 0.001)
	assert.Equal(t, DefaultFoodLevel, h.FoodLevel)
}

func TestAddExhaustion_DrainsFoodWhenSaturationZero(t *testing.T) {
	h := NewHunger()
	h.Saturation = 0
	h.AddExhaustion(ExhaustionThreshold)
	assert.InDelta(t, 0.0, h.Exhaustion, 0.001)
	assert.Equal(t, 0.0, h.Saturation)
	assert.Equal(t, DefaultFoodLevel-1, h.FoodLevel)
}

func TestAddExhaustion_MultipleThresholds(t *testing.T) {
	h := NewHunger()
	h.Saturation = 1.0
	// 12.0 = 3 * threshold (4.0).
	// First drain: saturation 1->0.
	// Second drain: food 20->19.
	// Third drain: food 19->18.
	h.AddExhaustion(12.0)
	assert.InDelta(t, 0.0, h.Exhaustion, 0.001)
	assert.Equal(t, 0.0, h.Saturation)
	assert.Equal(t, 18, h.FoodLevel)
}

func TestAddExhaustion_FoodDoesNotGoBelowZero(t *testing.T) {
	h := NewHunger()
	h.Saturation = 0
	h.FoodLevel = 0
	h.AddExhaustion(ExhaustionThreshold)
	assert.Equal(t, 0, h.FoodLevel)
	assert.Equal(t, 0.0, h.Saturation)
}

func TestAddExhaustion_SprintSources(t *testing.T) {
	h := NewHunger()
	// Simulate sprinting 40 metres: 40 * 0.1 = 4.0 = threshold.
	for range 40 {
		h.AddExhaustion(ExhaustionSprint)
	}
	// Should have drained saturation by 1.
	assert.InDelta(t, DefaultSaturation-1.0, h.Saturation, 0.001)
}

func TestAddExhaustion_JumpSource(t *testing.T) {
	h := NewHunger()
	// 81 jumps * 0.05 = 4.05, guarantees crossing the 4.0 threshold even
	// with floating-point accumulation.
	for range 81 {
		h.AddExhaustion(ExhaustionJump)
	}
	assert.InDelta(t, DefaultSaturation-1.0, h.Saturation, 0.01)
}

func TestAddExhaustion_AttackSource(t *testing.T) {
	h := NewHunger()
	// 40 attacks * 0.1 = 4.0 = threshold.
	for range 40 {
		h.AddExhaustion(ExhaustionAttack)
	}
	assert.InDelta(t, DefaultSaturation-1.0, h.Saturation, 0.001)
}

// ---------------------------------------------------------------------------
// Eat
// ---------------------------------------------------------------------------

func TestEat_RestoresFoodAndSaturation(t *testing.T) {
	h := NewHunger()
	h.FoodLevel = 10
	h.Saturation = 2.0
	h.Eat(4, 2.4)
	assert.Equal(t, 14, h.FoodLevel)
	assert.InDelta(t, 4.4, h.Saturation, 0.001)
}

func TestEat_ClampsToMaximum(t *testing.T) {
	h := NewHunger()
	h.FoodLevel = 18
	h.Saturation = 19.0
	h.Eat(8, 12.8)
	assert.Equal(t, MaxFoodLevel, h.FoodLevel)
	assert.Equal(t, MaxSaturation, h.Saturation)
}

// ---------------------------------------------------------------------------
// EatItem
// ---------------------------------------------------------------------------

func TestEatItem_FoodItem(t *testing.T) {
	h := NewHunger()
	h.FoodLevel = 10
	h.Saturation = 2.0
	ok := EatItem(&h, item.Apple)
	assert.True(t, ok)
	assert.Equal(t, 14, h.FoodLevel)
	assert.InDelta(t, 4.4, h.Saturation, 0.001)
}

func TestEatItem_NonFoodItem(t *testing.T) {
	h := NewHunger()
	ok := EatItem(&h, item.Dirt)
	assert.False(t, ok)
	assert.Equal(t, DefaultFoodLevel, h.FoodLevel)
}

func TestEatItem_FullHunger(t *testing.T) {
	h := NewHunger()
	h.FoodLevel = MaxFoodLevel
	ok := EatItem(&h, item.Apple)
	assert.False(t, ok)
}

func TestEatItem_AllFoodTypes(t *testing.T) {
	foods := []struct {
		id      item.ItemID
		food    int
		sat     float64
	}{
		{item.Apple, 4, 2.4},
		{item.Bread, 5, 6.0},
		{item.CookedPorkchop, 8, 12.8},
		{item.Steak, 8, 12.8},
		{item.GoldenApple, 4, 9.6},
		{item.Cookie, 2, 0.4},
		{item.Carrot, 3, 3.6},
		{item.BakedPotato, 5, 6.0},
	}
	for _, tc := range foods {
		h := Hunger{FoodLevel: 0, Saturation: 0}
		ok := EatItem(&h, tc.id)
		assert.True(t, ok, "expected eating %d to succeed", tc.id)
		assert.Equal(t, tc.food, h.FoodLevel, "food restore for %d", tc.id)
		assert.InDelta(t, tc.sat, h.Saturation, 0.001, "saturation for %d", tc.id)
	}
}

// ---------------------------------------------------------------------------
// HungerSystem - Health regeneration
// ---------------------------------------------------------------------------

func TestHungerSystem_HealthRegen(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  20,
		Saturation: 10.0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 18, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}

	// Tick 1.0 second: should regenerate 1 HP.
	sys.Update(w, 1.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(19), hp.Current)
}

func TestHungerSystem_HealthRegenClampsToMax(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  20,
		Saturation: 10.0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 19.5, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 1.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current)
}

func TestHungerSystem_NoRegenBelowFood18(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  17,
		Saturation: 10.0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 18, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 1.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(18), hp.Current)
}

func TestHungerSystem_NoRegenZeroSaturation(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  20,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 18, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 1.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(18), hp.Current)
}

func TestHungerSystem_NoRegenAtFullHealth(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  20,
		Saturation: 10.0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 1.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current)
}

// ---------------------------------------------------------------------------
// HungerSystem - Starvation damage
// ---------------------------------------------------------------------------

func TestHungerSystem_StarvationNormal(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  0,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}

	// Need >= 4 seconds for first starvation tick.
	sys.Update(w, 4.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(19), hp.Current)
}

func TestHungerSystem_StarvationNormal_StopsAt1HP(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  0,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 1.5, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 4.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(1.0), hp.Current)
}

func TestHungerSystem_StarvationEasy_StopsAt10HP(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  0,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 10.5, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyEasy}
	sys.Update(w, 4.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(10.0), hp.Current)
}

func TestHungerSystem_StarvationHard_CanKill(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  0,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 0.5, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyHard}
	sys.Update(w, 4.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Less(t, hp.Current, float32(0.5))
}

func TestHungerSystem_StarvationPeaceful_NoDamage(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  0,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyPeaceful}
	sys.Update(w, 8.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current)
}

func TestHungerSystem_NoStarvationWithFood(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  1,
		Saturation: 0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}
	sys.Update(w, 8.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current)
}

// ---------------------------------------------------------------------------
// SpawnPlayer includes Hunger component
// ---------------------------------------------------------------------------

func TestSpawnPlayer_HasHunger(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := SpawnPlayer(w, "Steve", pos)

	h, ok := ecs.GetStore[Hunger](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, DefaultFoodLevel, h.FoodLevel)
	assert.Equal(t, DefaultSaturation, h.Saturation)
	assert.Equal(t, 0.0, h.Exhaustion)
}

// ---------------------------------------------------------------------------
// HungerSystem - starvation timer resets when fed
// ---------------------------------------------------------------------------

func TestHungerSystem_StarvationTimerResetsWhenFed(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	hunger := Hunger{FoodLevel: 0, Saturation: 0}
	ecs.GetStore[Hunger](w).Set(e, hunger)
	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}

	// Accumulate 3 seconds of starvation timer.
	sys.Update(w, 3.0)

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current) // no damage yet

	// Now feed the player.
	h, _ := ecs.GetStore[Hunger](w).Get(e)
	h.FoodLevel = 10

	// Tick again: starvation timer should reset since food > 0.
	sys.Update(w, 2.0)

	hp, _ = ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current) // no damage, timer was reset
}

// ---------------------------------------------------------------------------
// HungerSystem - regen timer accumulates over multiple ticks
// ---------------------------------------------------------------------------

func TestHungerSystem_RegenMultipleTicks(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Hunger](w).Set(e, Hunger{
		FoodLevel:  20,
		Saturation: 20.0,
	})
	ecs.GetStore[Health](w).Set(e, Health{Current: 15, Max: 20})

	sys := &HungerSystem{Difficulty: DifficultyNormal}

	// 10 ticks of 0.5s = 5 seconds, should heal ~5 HP (one per second).
	for range 10 {
		sys.Update(w, 0.5)
	}

	hp, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(20), hp.Current)
}
