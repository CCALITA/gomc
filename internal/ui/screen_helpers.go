package ui

import (
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	// containerPanelGap is the vertical gap between the upper panel and the
	// player inventory panel in container screens (chest, furnace, crafting table).
	containerPanelGap = 12.0

	// defaultHitWidth is the fallback screen width for hit-test calculations.
	defaultHitWidth float32 = 800
	// defaultHitHeight is the fallback screen height for hit-test calculations.
	defaultHitHeight float32 = 600
)

// makeHitRenderer returns a minimal UIRenderer with the given screen
// dimensions for hit-test calculations. Falls back to 800x600 if either
// dimension is zero.
func makeHitRenderer(w, h float32) *UIRenderer {
	if w == 0 {
		w = defaultHitWidth
	}
	if h == 0 {
		h = defaultHitHeight
	}
	return &UIRenderer{ScreenWidth: w, ScreenHeight: h}
}

// swapHeldWithSlot swaps held with a slot in the given inventory,
// handling pick-up, place-down, merge, and swap cases. It returns the
// new held item.
func swapHeldWithSlot(held item.ItemStack, inv *inventory.Inventory, slotIdx int) item.ItemStack {
	if inv == nil {
		return held
	}
	current := inv.GetSlot(slotIdx)

	if held.IsEmpty() {
		inv.SetSlot(slotIdx, item.ItemStack{})
		return current
	}

	if current.IsEmpty() {
		inv.SetSlot(slotIdx, held)
		return item.ItemStack{}
	}

	if held.CanStackWith(current) {
		merged := current
		remaining := merged.Merge(held)
		inv.SetSlot(slotIdx, merged)
		return remaining
	}

	inv.SetSlot(slotIdx, held)
	return current
}

// playerPanelWidth returns the standard player inventory panel width.
func playerPanelWidth() float32 {
	return float32(invCols)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
}

// playerPanelHeight returns the standard player inventory panel height.
func playerPanelHeight() float32 {
	return float32(invRows)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
}
