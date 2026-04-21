package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// BrewingRecipe defines a single brewing transformation.
type BrewingRecipe struct {
	Ingredient uint16 // ingredient item ID
	Input      uint16 // base potion item ID
	Output     uint16 // resulting potion item ID
}

// brewingRecipes is the global list of registered brewing recipes.
var brewingRecipes []BrewingRecipe

func init() {
	registerBrewingRecipes()
}

// RegisterBrewingRecipe adds a brewing recipe to the global list.
func RegisterBrewingRecipe(r BrewingRecipe) {
	brewingRecipes = append(brewingRecipes, r)
}

// GetBrewingRecipes returns a copy of the global brewing recipe list.
func GetBrewingRecipes() []BrewingRecipe {
	out := make([]BrewingRecipe, len(brewingRecipes))
	copy(out, brewingRecipes)
	return out
}

// FindBrewingRecipe looks up a brewing recipe by ingredient and input potion.
// Returns the recipe and true if found, or a zero value and false otherwise.
func FindBrewingRecipe(ingredientID, inputID uint16) (BrewingRecipe, bool) {
	for _, r := range brewingRecipes {
		if r.Ingredient == ingredientID && r.Input == inputID {
			return r, true
		}
	}
	return BrewingRecipe{}, false
}

// registerBrewingRecipes populates the global brewing recipe list.
func registerBrewingRecipes() {
	RegisterBrewingRecipe(BrewingRecipe{Ingredient: item.NetherWart, Input: item.WaterBottle, Output: item.AwkwardPotion})
	RegisterBrewingRecipe(BrewingRecipe{Ingredient: item.SpiderEye, Input: item.AwkwardPotion, Output: item.PoisonPotion})
	RegisterBrewingRecipe(BrewingRecipe{Ingredient: item.BlazePowder, Input: item.AwkwardPotion, Output: item.StrengthPotion})
	RegisterBrewingRecipe(BrewingRecipe{Ingredient: item.GhastTear, Input: item.AwkwardPotion, Output: item.RegenPotion})
	RegisterBrewingRecipe(BrewingRecipe{Ingredient: item.Sugar, Input: item.AwkwardPotion, Output: item.SpeedPotion})
}

// MaxBrewTicks is the number of ticks required to complete one brew cycle.
const MaxBrewTicks = 400

// BrewingStand represents a brewing stand with three potion slots,
// an ingredient slot, and blaze powder fuel.
type BrewingStand struct {
	PotionSlots  [3]item.ItemStack
	Ingredient   item.ItemStack
	Fuel         int
	BrewTicks    int
	MaxBrewTicks int
}

// NewBrewingStand creates a new empty BrewingStand.
func NewBrewingStand() *BrewingStand {
	return &BrewingStand{MaxBrewTicks: MaxBrewTicks}
}

// hasValidRecipe reports whether at least one potion slot has a valid recipe
// with the current ingredient.
func (b *BrewingStand) hasValidRecipe() bool {
	if b.Ingredient.IsEmpty() {
		return false
	}
	for _, slot := range b.PotionSlots {
		if slot.IsEmpty() {
			continue
		}
		if _, ok := FindBrewingRecipe(b.Ingredient.ItemID, slot.ItemID); ok {
			return true
		}
	}
	return false
}

// Update advances the brewing stand simulation by dt seconds.
// Each tick is 0.05 seconds (20 ticks per second). At MaxBrewTicks (400 ticks = 20s),
// all valid potions are transformed according to their recipes.
func (b *BrewingStand) Update(dt float64) {
	if dt <= 0 {
		return
	}

	if !b.hasValidRecipe() {
		b.BrewTicks = 0
		return
	}

	if b.Fuel <= 0 {
		b.BrewTicks = 0
		return
	}

	ticks := int(dt / 0.05)
	if ticks <= 0 {
		return
	}

	// Start brewing: consume one fuel unit at the start of a brew cycle.
	if b.BrewTicks == 0 {
		b.Fuel--
	}

	b.BrewTicks += ticks

	if b.BrewTicks >= b.MaxBrewTicks {
		// Brew complete: transform all valid potion slots.
		for i, slot := range b.PotionSlots {
			if slot.IsEmpty() {
				continue
			}
			recipe, ok := FindBrewingRecipe(b.Ingredient.ItemID, slot.ItemID)
			if !ok {
				continue
			}
			b.PotionSlots[i] = item.NewItemStack(recipe.Output, slot.Count)
		}

		// Consume one ingredient.
		b.Ingredient.Count--
		if b.Ingredient.Count <= 0 {
			b.Ingredient = item.ItemStack{}
		}

		b.BrewTicks = 0
	}
}
