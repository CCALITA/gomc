package player

import (
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	// HotbarSlots is the number of hotbar slots.
	HotbarSlots = 9
)

// SelectSlot sets the selected hotbar index, clamped to [0, 8].
func (c *Controller) SelectSlot(index int) {
	if index < 0 {
		index = 0
	}
	if index >= HotbarSlots {
		index = HotbarSlots - 1
	}
	c.SelectedSlot = index
}

// ScrollSlot adjusts the selected slot by delta, wrapping around.
func (c *Controller) ScrollSlot(delta int) {
	c.SelectedSlot = ((c.SelectedSlot + delta) % HotbarSlots + HotbarSlots) % HotbarSlots
}

// GetSelectedItem returns the item stack in the selected hotbar slot
// from the given inventory.
func (c *Controller) GetSelectedItem(inv *inventory.Inventory) item.ItemStack {
	return inv.GetSlot(c.SelectedSlot)
}

// updateHotbar reads hotbar number keys (1-9) and scroll wheel input.
func (c *Controller) updateHotbar(inp *input.Manager) {
	// Check number keys 1-9.
	hotbarActions := []input.Action{
		input.Hotbar1, input.Hotbar2, input.Hotbar3,
		input.Hotbar4, input.Hotbar5, input.Hotbar6,
		input.Hotbar7, input.Hotbar8, input.Hotbar9,
	}
	for i, action := range hotbarActions {
		key := c.KeyMap.GetKey(action)
		if inp.IsKeyJustPressed(key) {
			c.SelectSlot(i)
			return
		}
	}

	// Check scroll wheel.
	scroll := inp.ScrollDelta()
	if scroll > 0 {
		c.ScrollSlot(-1)
	} else if scroll < 0 {
		c.ScrollSlot(1)
	}
}
