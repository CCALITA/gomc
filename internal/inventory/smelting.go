package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// SmeltingRecipe defines a single smelting transformation.
type SmeltingRecipe struct {
	Input    uint16  // input item ID
	Output   uint16  // output item ID
	Duration float64 // time in seconds to smelt one item
}

// smeltingRecipes is the global list of registered smelting recipes.
var smeltingRecipes []SmeltingRecipe

func init() {
	registerSmeltingRecipes()
}

// RegisterSmeltingRecipe adds a smelting recipe to the global list.
func RegisterSmeltingRecipe(r SmeltingRecipe) {
	smeltingRecipes = append(smeltingRecipes, r)
}

// GetSmeltingRecipes returns a copy of the global smelting recipe list.
func GetSmeltingRecipes() []SmeltingRecipe {
	out := make([]SmeltingRecipe, len(smeltingRecipes))
	copy(out, smeltingRecipes)
	return out
}

// FindSmeltingRecipe looks up a smelting recipe by input item ID.
// Returns the recipe and true if found, or a zero value and false otherwise.
func FindSmeltingRecipe(inputID uint16) (SmeltingRecipe, bool) {
	for _, r := range smeltingRecipes {
		if r.Input == inputID {
			return r, true
		}
	}
	return SmeltingRecipe{}, false
}

// registerSmeltingRecipes populates the global smelting recipe list.
func registerSmeltingRecipes() {
	RegisterSmeltingRecipe(SmeltingRecipe{Input: item.IronOre, Output: item.IronIngot, Duration: 10.0})
	RegisterSmeltingRecipe(SmeltingRecipe{Input: item.GoldOre, Output: item.GoldIngot, Duration: 10.0})
	RegisterSmeltingRecipe(SmeltingRecipe{Input: item.Sand, Output: item.Glass, Duration: 10.0})
	RegisterSmeltingRecipe(SmeltingRecipe{Input: item.Cobblestone, Output: item.Stone, Duration: 10.0})
	RegisterSmeltingRecipe(SmeltingRecipe{Input: item.OakLog, Output: item.Coal, Duration: 10.0})
}

// DefaultFuelBurnTime is the burn duration (in seconds) for one unit of fuel.
const DefaultFuelBurnTime = 10.0

// Furnace represents a smelting furnace with input, fuel, and output slots.
type Furnace struct {
	InputSlot  item.ItemStack
	FuelSlot   item.ItemStack
	OutputSlot item.ItemStack
	Progress   float64 // current smelting progress in seconds
	BurnTime   float64 // remaining burn time from current fuel
}

// NewFurnace creates a new empty Furnace.
func NewFurnace() *Furnace {
	return &Furnace{}
}

// Update advances the furnace simulation by dt seconds.
// It consumes fuel, processes smelting, and moves results to the output slot.
func (f *Furnace) Update(dt float64) {
	if dt <= 0 {
		return
	}

	// Nothing to smelt if input is empty.
	if f.InputSlot.IsEmpty() {
		f.Progress = 0
		return
	}

	// Look up the smelting recipe.
	recipe, found := FindSmeltingRecipe(f.InputSlot.ItemID)
	if !found {
		f.Progress = 0
		return
	}

	// Check that the output slot can accept the result.
	if !f.OutputSlot.IsEmpty() {
		if f.OutputSlot.ItemID != recipe.Output {
			f.Progress = 0
			return
		}
		if f.OutputSlot.Count >= item.MaxStack(recipe.Output) {
			f.Progress = 0
			return
		}
	}

	remaining := dt

	for remaining > 0 {
		// Try to consume fuel if burn time is depleted.
		if f.BurnTime <= 0 {
			if f.FuelSlot.IsEmpty() {
				f.Progress = 0
				return
			}
			f.FuelSlot.Count--
			if f.FuelSlot.Count <= 0 {
				f.FuelSlot = item.ItemStack{}
			}
			f.BurnTime = DefaultFuelBurnTime
		}

		// Re-check input (may have been consumed).
		if f.InputSlot.IsEmpty() {
			f.Progress = 0
			return
		}

		// Re-check recipe validity.
		recipe, found = FindSmeltingRecipe(f.InputSlot.ItemID)
		if !found {
			f.Progress = 0
			return
		}

		// Re-check output capacity.
		if !f.OutputSlot.IsEmpty() {
			if f.OutputSlot.ItemID != recipe.Output {
				f.Progress = 0
				return
			}
			if f.OutputSlot.Count >= item.MaxStack(recipe.Output) {
				f.Progress = 0
				return
			}
		}

		// Determine how much time we can advance this tick.
		timeNeeded := recipe.Duration - f.Progress
		timeAvailable := remaining
		if timeAvailable > f.BurnTime {
			timeAvailable = f.BurnTime
		}

		if timeAvailable >= timeNeeded {
			// Smelting completes.
			f.BurnTime -= timeNeeded
			remaining -= timeNeeded
			f.Progress = 0

			// Consume one input item.
			f.InputSlot.Count--
			if f.InputSlot.Count <= 0 {
				f.InputSlot = item.ItemStack{}
			}

			// Produce one output item.
			if f.OutputSlot.IsEmpty() {
				f.OutputSlot = item.NewItemStack(recipe.Output, 1)
			} else {
				f.OutputSlot.Count++
			}
		} else {
			// Partial progress.
			f.Progress += timeAvailable
			f.BurnTime -= timeAvailable
			remaining = 0
		}
	}
}
