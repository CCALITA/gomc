package ui

import (
	"fmt"

	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	hotbarSlots    = 9
	maxHearts      = 10
	maxHunger      = 10
	slotSize       = 40.0
	slotPadding    = 4.0
	heartSize      = 16.0
	heartPadding   = 2.0
	hudBarHeight   = 48.0
	crosshairSize  = 16.0
	crosshairWidth = 2.0
)

// HUD is the always-visible heads-up display: crosshair, hotbar, health
// and hunger bars, and an optional FPS counter.
type HUD struct {
	// Hotbar references the player's hotbar inventory (9 slots).
	Hotbar *inventory.Inventory

	// SelectedSlot is the currently selected hotbar slot (0-8).
	SelectedSlot int

	// Health is the current health value (0-20, displayed as hearts).
	Health int

	// MaxHealth is the maximum health value.
	MaxHealth int

	// Hunger is the current hunger value (0-20, displayed as drumsticks).
	Hunger int

	// MaxHunger is the maximum hunger value.
	MaxHunger int

	// ShowFPS toggles the FPS counter in the top-left corner.
	ShowFPS bool

	// FPS is the current frames-per-second value for display.
	FPS int
}

// NewHUD creates a HUD bound to the given hotbar inventory.
func NewHUD(hotbar *inventory.Inventory) *HUD {
	return &HUD{
		Hotbar:    hotbar,
		Health:    20,
		MaxHealth: 20,
		Hunger:    20,
		MaxHunger: 20,
	}
}

// Update processes HUD-specific input: hotbar selection via number keys
// and FPS counter toggle via F3.
func (h *HUD) Update(inp *input.Manager, _ float64) {
	// Hotbar slot selection via 1-9 keys.
	hotbarKeys := []int{
		input.Key1, input.Key2, input.Key3,
		input.Key4, input.Key5, input.Key6,
		input.Key7, input.Key8, input.Key9,
	}
	for i, key := range hotbarKeys {
		if inp.IsKeyJustPressed(key) {
			h.SelectedSlot = i
			break
		}
	}

	// Toggle FPS counter with F3.
	if inp.IsKeyJustPressed(input.KeyF3) {
		h.ShowFPS = !h.ShowFPS
	}
}

// Draw emits draw commands for all HUD elements.
func (h *HUD) Draw(r *UIRenderer) {
	h.drawCrosshair(r)
	h.drawHotbar(r)
	h.drawHealthBar(r)
	h.drawHungerBar(r)
	if h.ShowFPS {
		h.drawFPS(r)
	}
}

// IsOverlay returns true: the HUD is drawn on top of the game world.
func (h *HUD) IsOverlay() bool {
	return false
}

// HandleKey is a no-op for the HUD; input is handled in Update.
func (h *HUD) HandleKey(_ int) {}

// drawCrosshair renders a + shape at the center of the screen.
func (h *HUD) drawCrosshair(r *UIRenderer) {
	cx := r.ScreenWidth / 2
	cy := r.ScreenHeight / 2

	// Horizontal bar.
	r.DrawRect(
		cx-crosshairSize/2, cy-crosshairWidth/2,
		crosshairSize, crosshairWidth,
		1, 1, 1, 0.8,
	)
	// Vertical bar.
	r.DrawRect(
		cx-crosshairWidth/2, cy-crosshairSize/2,
		crosshairWidth, crosshairSize,
		1, 1, 1, 0.8,
	)
}

// drawHotbar renders the 9-slot hotbar at the bottom center.
func (h *HUD) drawHotbar(r *UIRenderer) {
	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	startX := (r.ScreenWidth - totalWidth) / 2
	startY := r.ScreenHeight - hudBarHeight

	for i := 0; i < hotbarSlots; i++ {
		x := startX + float32(i)*(slotSize+slotPadding)
		y := startY

		// Slot background.
		if i == h.SelectedSlot {
			// Highlighted slot: brighter border.
			r.DrawRect(x-2, y-2, slotSize+4, slotSize+4, 1, 1, 1, 0.6)
		}
		r.DrawRect(x, y, slotSize, slotSize, 0.2, 0.2, 0.2, 0.8)

		// Item icon.
		if h.Hotbar != nil {
			stack := h.Hotbar.GetSlot(i)
			if !stack.IsEmpty() {
				r.DrawItemSlot(x+4, y+4, stack)
			}
		}
	}
}

// drawHealthBar renders hearts above the hotbar on the left side.
func (h *HUD) drawHealthBar(r *UIRenderer) {
	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	startX := (r.ScreenWidth - totalWidth) / 2
	y := r.ScreenHeight - hudBarHeight - heartSize - 4

	hearts := h.Health / 2
	halfHeart := h.Health%2 == 1
	maxHeartCount := h.MaxHealth / 2
	if maxHeartCount > maxHearts {
		maxHeartCount = maxHearts
	}

	for i := 0; i < maxHeartCount; i++ {
		x := startX + float32(i)*(heartSize+heartPadding)
		// Background (empty heart).
		r.DrawRect(x, y, heartSize, heartSize, 0.3, 0.0, 0.0, 0.6)

		if i < hearts {
			// Full heart.
			r.DrawRect(x+2, y+2, heartSize-4, heartSize-4, 0.9, 0.1, 0.1, 1.0)
		} else if i == hearts && halfHeart {
			// Half heart (left half filled).
			r.DrawRect(x+2, y+2, (heartSize-4)/2, heartSize-4, 0.9, 0.1, 0.1, 1.0)
		}
	}
}

// drawHungerBar renders drumstick icons above the hotbar on the right side.
func (h *HUD) drawHungerBar(r *UIRenderer) {
	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	endX := (r.ScreenWidth+totalWidth)/2 - heartSize
	y := r.ScreenHeight - hudBarHeight - heartSize - 4

	drumsticks := h.Hunger / 2
	halfDrumstick := h.Hunger%2 == 1
	maxDrumstickCount := h.MaxHunger / 2
	if maxDrumstickCount > maxHunger {
		maxDrumstickCount = maxHunger
	}

	for i := 0; i < maxDrumstickCount; i++ {
		// Draw right-to-left.
		x := endX - float32(i)*(heartSize+heartPadding)
		// Background (empty drumstick).
		r.DrawRect(x, y, heartSize, heartSize, 0.3, 0.2, 0.0, 0.6)

		if i < drumsticks {
			// Full drumstick.
			r.DrawRect(x+2, y+2, heartSize-4, heartSize-4, 0.8, 0.6, 0.2, 1.0)
		} else if i == drumsticks && halfDrumstick {
			// Half drumstick.
			r.DrawRect(x+2, y+2, (heartSize-4)/2, heartSize-4, 0.8, 0.6, 0.2, 1.0)
		}
	}
}

// drawFPS renders the FPS counter in the top-left corner.
func (h *HUD) drawFPS(r *UIRenderer) {
	text := fmt.Sprintf("FPS: %d", h.FPS)
	r.DrawText(8, 8, text, 1.0, 1, 1, 1)
}

// GetSelectedItem returns the item stack in the currently selected hotbar slot.
func (h *HUD) GetSelectedItem() item.ItemStack {
	if h.Hotbar == nil {
		return item.ItemStack{}
	}
	return h.Hotbar.GetSlot(h.SelectedSlot)
}
