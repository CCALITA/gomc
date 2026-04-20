package player

import (
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

// Eating animation constants.
const (
	// MaxEatTicks is the number of ticks required to finish eating (32 ticks = 1.6 s at 20 TPS).
	MaxEatTicks = 32
)

// EatingState tracks whether the player is currently eating, how far
// along the eating animation has progressed, and which item is being consumed.
type EatingState struct {
	Eating      bool
	EatTicks    int
	MaxEatTicks int
	ItemID      uint16
}

// NewEatingState returns an EatingState with default values.
func NewEatingState() EatingState {
	return EatingState{
		MaxEatTicks: MaxEatTicks,
	}
}

// StartEating begins the eating animation for the given food item.
// It returns a new EatingState with the animation started.
func StartEating(itemID uint16) EatingState {
	return EatingState{
		Eating:      true,
		EatTicks:    0,
		MaxEatTicks: MaxEatTicks,
		ItemID:      itemID,
	}
}

// Cancel returns a reset EatingState with the eating animation stopped.
func (es EatingState) Cancel() EatingState {
	return NewEatingState()
}

// Tick advances the eating timer by one tick and returns the updated state.
// Returns the new state and whether eating just completed.
func (es EatingState) Tick() (EatingState, bool) {
	if !es.Eating {
		return es, false
	}
	next := EatingState{
		Eating:      true,
		EatTicks:    es.EatTicks + 1,
		MaxEatTicks: es.MaxEatTicks,
		ItemID:      es.ItemID,
	}
	if next.EatTicks >= next.MaxEatTicks {
		return NewEatingState(), true
	}
	return next, false
}

// UpdateEating processes the eating mechanic for a single tick.
// It reads input to start, continue, or cancel eating, and applies
// hunger/saturation restoration when eating completes.
//
// The returned EatingState should replace the caller's copy.
func UpdateEating(
	es EatingState,
	inp *input.Manager,
	keyMap *input.KeyMap,
	inv *inventory.Inventory,
	selectedSlot int,
	hunger *entity.Hunger,
) EatingState {
	useBtn := keyMap.GetKey(input.Use)
	attackBtn := keyMap.GetKey(input.Attack)

	if es.Eating {
		// Cancel on left-click or right-click release.
		if inp.IsMouseDown(attackBtn) || !inp.IsMouseDown(useBtn) {
			return es.Cancel()
		}

		// Advance eating timer.
		next, completed := es.Tick()
		if completed {
			consumeFood(inv, selectedSlot, hunger, es.ItemID)
		}
		return next
	}

	// Try to start eating: right-click held, food item selected, not full.
	if !inp.IsMouseDown(useBtn) {
		return es
	}
	if hunger.FoodLevel >= entity.MaxFoodLevel {
		return es
	}

	stack := inv.GetSlot(selectedSlot)
	if stack.IsEmpty() || !item.IsFood(stack.ItemID) {
		return es
	}

	return StartEating(stack.ItemID)
}

// consumeFood removes one food item from the inventory slot and restores
// the player's hunger and saturation via entity.EatItem.
func consumeFood(inv *inventory.Inventory, slot int, hunger *entity.Hunger, itemID uint16) {
	stack := inv.GetSlot(slot)
	if stack.IsEmpty() || stack.ItemID != itemID {
		return
	}

	if entity.EatItem(hunger, itemID) {
		inv.RemoveItem(slot, 1)
	}
}
