package player

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/item"
)

// tryEatFood attempts to consume the currently held food item. It checks
// that the held item is food, the player's hunger is not full, and then
// delegates to entity.EatItem to restore hunger and saturation. On success
// one item is removed from the selected hotbar slot.
// Returns true if a food item was eaten.
func (c *Controller) tryEatFood() bool {
	held := c.getSelectedHotbarItem()
	if held.IsEmpty() {
		return false
	}

	props := item.GetProperties(held.ItemID)
	if props.FoodRestore <= 0 {
		return false
	}

	// Get the hunger component from ECS.
	hungerStore := ecs.GetStore[entity.Hunger](c.ECSWorld)
	hunger, ok := hungerStore.Get(c.Entity)
	if !ok || hunger.FoodLevel >= entity.MaxFoodLevel {
		return false
	}

	entity.EatItem(hunger, held.ItemID)
	c.Inventory.RemoveItem(c.SelectedSlot, 1)
	return true
}
