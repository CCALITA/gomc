package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
)

// Difficulty constants control starvation behaviour.
const (
	// DifficultyPeaceful disables starvation damage entirely.
	DifficultyPeaceful uint8 = 0
	// DifficultyEasy stops starvation damage at 10 HP.
	DifficultyEasy uint8 = 1
	// DifficultyNormal stops starvation damage at 1 HP.
	DifficultyNormal uint8 = 2
	// DifficultyHard allows starvation to kill the player.
	DifficultyHard uint8 = 3
)

// Hunger defaults.
const (
	// DefaultFoodLevel is the starting food level for a new player.
	DefaultFoodLevel = 20
	// DefaultSaturation is the starting saturation for a new player.
	DefaultSaturation = 5.0
	// MaxFoodLevel is the upper bound for food level.
	MaxFoodLevel = 20
	// MaxSaturation is the upper bound for saturation.
	MaxSaturation = 20.0
)

// Exhaustion thresholds and rates.
const (
	// ExhaustionThreshold is the exhaustion level that triggers saturation or food drain.
	ExhaustionThreshold = 4.0
	// ExhaustionSprint is the exhaustion added per metre of sprinting.
	ExhaustionSprint = 0.1
	// ExhaustionJump is the exhaustion added per jump.
	ExhaustionJump = 0.05
	// ExhaustionAttack is the exhaustion added per attack.
	ExhaustionAttack = 0.1
	// HealthRegenFoodMin is the minimum food level required to regenerate HP.
	HealthRegenFoodMin = 18
	// HealthRegenRate is the HP regenerated per second when food >= 18 and saturation > 0.
	HealthRegenRate = 1.0
	// HealthRegenExhaustion is the exhaustion cost per HP regenerated.
	HealthRegenExhaustion = 6.0
	// StarvationDamage is the HP lost per starvation tick on Normal difficulty.
	StarvationDamage = 1.0
	// StarvationInterval is the number of seconds between starvation damage ticks.
	StarvationInterval = 4.0
	// StarvationMinHealthEasy is the HP floor below which starvation stops on Easy.
	StarvationMinHealthEasy float32 = 10.0
	// StarvationMinHealthNormal is the HP floor below which starvation stops on Normal.
	StarvationMinHealthNormal float32 = 1.0
)

// Hunger is an ECS component tracking the player's food, saturation,
// and exhaustion levels.
type Hunger struct {
	FoodLevel       int     // 0-20, default 20
	Saturation      float64 // 0-20, default 5
	Exhaustion      float64 // accumulated exhaustion
	StarvationTimer float64 // seconds since last starvation tick
	RegenTimer      float64 // seconds since last health regen tick
}

// NewHunger returns a Hunger component with default values.
func NewHunger() Hunger {
	return Hunger{
		FoodLevel:  DefaultFoodLevel,
		Saturation: DefaultSaturation,
	}
}

// AddExhaustion adds exhaustion to the hunger component.
// When exhaustion reaches the threshold, saturation (or food level) is
// drained and exhaustion resets.
func (h *Hunger) AddExhaustion(amount float64) {
	h.Exhaustion += amount
	for h.Exhaustion >= ExhaustionThreshold {
		h.Exhaustion -= ExhaustionThreshold
		if h.Saturation > 0 {
			h.Saturation -= 1.0
			if h.Saturation < 0 {
				h.Saturation = 0
			}
		} else {
			h.FoodLevel--
			if h.FoodLevel < 0 {
				h.FoodLevel = 0
			}
		}
	}
}

// Eat restores food level and saturation from a food item, clamping to maximums.
func (h *Hunger) Eat(foodRestore int, foodSaturation float64) {
	h.FoodLevel += foodRestore
	if h.FoodLevel > MaxFoodLevel {
		h.FoodLevel = MaxFoodLevel
	}
	h.Saturation += foodSaturation
	if h.Saturation > MaxSaturation {
		h.Saturation = MaxSaturation
	}
}

// ---------------------------------------------------------------------------
// HungerSystem
// ---------------------------------------------------------------------------

// HungerSystem processes hunger mechanics each tick: health regeneration
// when well-fed, and starvation damage when food is depleted.
type HungerSystem struct {
	Difficulty uint8
}

// Update processes hunger for every entity with both Hunger and Health
// components.
func (s *HungerSystem) Update(w *ecs.World, dt float64) {
	ecs.Query2[Hunger, Health](w, func(e ecs.Entity, h *Hunger, hp *Health) {
		// Health regeneration: food >= 18 and saturation > 0.
		if h.FoodLevel >= HealthRegenFoodMin && h.Saturation > 0 && hp.Current < hp.Max {
			h.RegenTimer += dt
			if h.RegenTimer >= 1.0 {
				h.RegenTimer -= 1.0
				hp.Current += float32(HealthRegenRate)
				if hp.Current > hp.Max {
					hp.Current = hp.Max
				}
				// Regeneration costs exhaustion.
				h.AddExhaustion(HealthRegenExhaustion)
			}
		} else {
			h.RegenTimer = 0
		}

		// Starvation damage when food level is 0.
		if h.FoodLevel <= 0 {
			h.StarvationTimer += dt
			if h.StarvationTimer >= StarvationInterval {
				h.StarvationTimer -= StarvationInterval
				applyStarvation(s.Difficulty, hp)
			}
		} else {
			h.StarvationTimer = 0
		}
	})
}

// applyStarvation applies starvation damage based on difficulty.
func applyStarvation(difficulty uint8, hp *Health) {
	switch difficulty {
	case DifficultyPeaceful:
		return
	case DifficultyEasy:
		damageWithFloor(hp, StarvationMinHealthEasy)
	case DifficultyNormal:
		damageWithFloor(hp, StarvationMinHealthNormal)
	case DifficultyHard:
		hp.Current -= float32(StarvationDamage)
	}
}

// damageWithFloor applies starvation damage but prevents health from
// dropping below the given floor.
func damageWithFloor(hp *Health, floor float32) {
	if hp.Current > floor {
		hp.Current -= float32(StarvationDamage)
		if hp.Current < floor {
			hp.Current = floor
		}
	}
}

// EatItem attempts to consume a food item, restoring hunger and saturation.
// Returns true if the item was food and was eaten.
func EatItem(h *Hunger, itemID item.ItemID) bool {
	props := item.GetProperties(itemID)
	if props.FoodRestore <= 0 {
		return false
	}
	if h.FoodLevel >= MaxFoodLevel {
		return false
	}
	h.Eat(props.FoodRestore, props.FoodSaturation)
	return true
}
