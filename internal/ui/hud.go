package ui

import (
	"fmt"
	"math"

	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	hotbarSlots    = 9
	maxHearts      = 10
	maxHunger      = 10
	maxArmorIcons  = 10
	maxBubbles     = 10
	slotSize       = 40.0
	slotPadding    = 4.0
	heartSize      = 16.0
	heartPadding   = 2.0
	hudBarHeight   = 48.0
	crosshairSize  = 16.0
	crosshairWidth = 2.0
	xpBarHeight    = 6.0
	xpBarGap       = 4.0
)

// HUD is the always-visible heads-up display: crosshair, hotbar, health
// and hunger bars, and an optional debug overlay (F3).
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

	// ArmorPoints is the current armor value (0-20, displayed as shield icons).
	ArmorPoints int

	// AirBubbles is the current air supply (0-10, displayed as bubble icons).
	AirBubbles int

	// MaxAirBubbles is the maximum air supply (default 10).
	MaxAirBubbles int

	// ShowFPS toggles the debug overlay in the top-left corner.
	ShowFPS bool

	// FPS is the current frames-per-second value for display.
	FPS int

	// PlayerX, PlayerY, PlayerZ hold the player's world position.
	PlayerX, PlayerY, PlayerZ float32

	// PlayerYaw and PlayerPitch hold the player's rotation in degrees.
	PlayerYaw, PlayerPitch float32

	// ChunkX and ChunkZ are the chunk coordinates the player is in.
	ChunkX, ChunkZ int32

	// LoadedChunks is the number of currently loaded chunks.
	LoadedChunks int

	// EntityCount is the total number of active entities.
	EntityCount int

	// MemoryMB is the current memory usage in megabytes.
	MemoryMB float64

	// FacingDirection is the cardinal direction computed from PlayerYaw.
	FacingDirection string

	// SpawnX and SpawnZ hold the world spawn point (for compass display).
	SpawnX, SpawnZ float32

	// GameTick is the current game tick (for clock display).
	GameTick int64

	// XPLevel is the player's current experience level (displayed as text).
	XPLevel int

	// XPProgress is the progress toward the next level (0.0 to 1.0).
	XPProgress float32
}

// NewHUD creates a HUD bound to the given hotbar inventory.
func NewHUD(hotbar *inventory.Inventory) *HUD {
	return &HUD{
		Hotbar:        hotbar,
		Health:        20,
		MaxHealth:     20,
		Hunger:        20,
		MaxHunger:     20,
		AirBubbles:    10,
		MaxAirBubbles: 10,
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
	h.drawXPBar(r)
	h.drawHealthBar(r)
	h.drawHungerBar(r)
	h.drawArmorBar(r)
	h.drawBreathBar(r)
	h.drawFunctionalItemOverlay(r)
	if h.ShowFPS {
		h.drawDebugOverlay(r)
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

// drawXPBar renders a green experience progress bar below the hotbar with
// the current XP level displayed as centered text.
func (h *HUD) drawXPBar(r *UIRenderer) {
	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	startX := (r.ScreenWidth - totalWidth) / 2
	y := r.ScreenHeight - hudBarHeight + slotSize + xpBarGap

	// Dark background bar.
	r.DrawRect(startX, y, totalWidth, xpBarHeight, 0.1, 0.1, 0.1, 0.8)

	// Green filled portion.
	progress := h.XPProgress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	if progress > 0 {
		r.DrawRect(startX, y, totalWidth*progress, xpBarHeight, 0.3, 0.9, 0.1, 0.9)
	}

	// XP level text centered above the bar.
	if h.XPLevel > 0 {
		label := fmt.Sprintf("%d", h.XPLevel)
		textY := y - 14
		textX := r.ScreenWidth/2 - float32(len(label))*4
		r.DrawText(textX, textY, label, 1.0, 0.3, 0.9, 0.1)
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

// drawArmorBar renders shield icons above the health bar on the left side.
// The bar is only visible when ArmorPoints > 0.
func (h *HUD) drawArmorBar(r *UIRenderer) {
	if h.ArmorPoints <= 0 {
		return
	}

	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	startX := (r.ScreenWidth - totalWidth) / 2
	// Position one row above the health bar.
	y := r.ScreenHeight - hudBarHeight - 2*(heartSize+4)

	fullIcons := h.ArmorPoints / 2
	halfIcon := h.ArmorPoints%2 == 1
	maxIconCount := maxArmorIcons

	for i := 0; i < maxIconCount; i++ {
		x := startX + float32(i)*(heartSize+heartPadding)
		// Background (empty shield).
		r.DrawRect(x, y, heartSize, heartSize, 0.3, 0.3, 0.3, 0.4)

		if i < fullIcons {
			// Full shield.
			r.DrawRect(x+2, y+2, heartSize-4, heartSize-4, 0.7, 0.7, 0.7, 1.0)
		} else if i == fullIcons && halfIcon {
			// Half shield (left half filled).
			r.DrawRect(x+2, y+2, (heartSize-4)/2, heartSize-4, 0.7, 0.7, 0.7, 1.0)
		}
	}
}

// drawBreathBar renders bubble icons above the hunger bar on the right side.
// The bar is only visible when AirBubbles < MaxAirBubbles (i.e. underwater).
func (h *HUD) drawBreathBar(r *UIRenderer) {
	maxBubbleCount := h.MaxAirBubbles
	if maxBubbleCount <= 0 {
		maxBubbleCount = maxBubbles
	}
	if maxBubbleCount > maxBubbles {
		maxBubbleCount = maxBubbles
	}

	if h.AirBubbles >= maxBubbleCount {
		return
	}

	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	endX := (r.ScreenWidth+totalWidth)/2 - heartSize
	// Position one row above the hunger bar.
	y := r.ScreenHeight - hudBarHeight - 2*(heartSize+4)

	for i := 0; i < maxBubbleCount; i++ {
		// Draw right-to-left, matching hunger bar direction.
		x := endX - float32(i)*(heartSize+heartPadding)
		// Background (empty bubble).
		r.DrawRect(x, y, heartSize, heartSize, 0.1, 0.2, 0.4, 0.4)

		if i < h.AirBubbles {
			// Full bubble.
			r.DrawRect(x+2, y+2, heartSize-4, heartSize-4, 0.3, 0.6, 0.9, 1.0)
		}
	}
}

// debugTextLines returns the lines of text displayed in the debug overlay.
func (h *HUD) debugTextLines() []string {
	frameTime := float64(0)
	if h.FPS > 0 {
		frameTime = 1000.0 / float64(h.FPS)
	}

	return []string{
		"GoMC (Vulkan)",
		fmt.Sprintf("FPS: %d (%.1fms)", h.FPS, frameTime),
		"",
		fmt.Sprintf("XYZ: %.1f / %.1f / %.1f", h.PlayerX, h.PlayerY, h.PlayerZ),
		fmt.Sprintf("Chunk: %d / %d", h.ChunkX, h.ChunkZ),
		fmt.Sprintf("Facing: %s (Yaw: %.1f, Pitch: %.1f)", h.FacingDirection, h.PlayerYaw, h.PlayerPitch),
		"",
		fmt.Sprintf("Loaded Chunks: %d", h.LoadedChunks),
		fmt.Sprintf("Entities: %d", h.EntityCount),
		fmt.Sprintf("Memory: %.0f MB", h.MemoryMB),
	}
}

// drawDebugOverlay renders the F3 debug panel on the left side of the screen.
func (h *HUD) drawDebugOverlay(r *UIRenderer) {
	lines := h.debugTextLines()

	const (
		lineHeight = 14.0
		padX       = 8.0
		padY       = 8.0
		panelWidth = 260.0
	)

	panelHeight := padY*2 + float32(len(lines))*lineHeight

	// Semi-transparent dark background panel.
	r.DrawRect(0, 0, panelWidth, panelHeight, 0, 0, 0, 0.5)

	// Draw each line of debug text.
	for i, line := range lines {
		if line == "" {
			continue
		}
		y := padY + float32(i)*lineHeight
		r.DrawText(padX, y, line, 1.0, 1, 1, 1)
	}
}

// SetPlayerPos updates the player's world position on the debug overlay.
func (h *HUD) SetPlayerPos(x, y, z float32) {
	h.PlayerX = x
	h.PlayerY = y
	h.PlayerZ = z
}

// SetPlayerRotation updates the player's rotation and recomputes the
// facing direction from the yaw angle.
func (h *HUD) SetPlayerRotation(yaw, pitch float32) {
	h.PlayerYaw = yaw
	h.PlayerPitch = pitch
	h.FacingDirection = facingDirectionFromYaw(yaw)
}

// SetChunkPos updates the chunk coordinates on the debug overlay.
func (h *HUD) SetChunkPos(x, z int32) {
	h.ChunkX = x
	h.ChunkZ = z
}

// SetLoadedChunks updates the loaded chunk count on the debug overlay.
func (h *HUD) SetLoadedChunks(n int) {
	h.LoadedChunks = n
}

// SetEntityCount updates the entity count on the debug overlay.
func (h *HUD) SetEntityCount(n int) {
	h.EntityCount = n
}

// SetMemoryMB updates the memory usage on the debug overlay.
func (h *HUD) SetMemoryMB(mb float64) {
	h.MemoryMB = mb
}

// facingDirectionFromYaw returns a cardinal direction string for the given
// yaw angle in degrees. The mapping follows Minecraft conventions:
//
//	-45 to 45   -> South  (toward +Z)
//	 45 to 135  -> West   (toward -X)
//	135 to 180 or -180 to -135 -> North (toward -Z)
//	-135 to -45 -> East   (toward +X)
func facingDirectionFromYaw(yaw float32) string {
	// Normalize yaw to [-180, 180).
	y := normalizeYaw(yaw)

	switch {
	case y >= -45 && y < 45:
		return "South"
	case y >= 45 && y < 135:
		return "West"
	case y >= -135 && y < -45:
		return "East"
	default:
		// y >= 135 || y < -135
		return "North"
	}
}

// normalizeYaw reduces a yaw angle to the range [-180, 180).
func normalizeYaw(yaw float32) float32 {
	r := math.Remainder(float64(yaw), 360)
	return float32(r)
}

// GetSelectedItem returns the item stack in the currently selected hotbar slot.
func (h *HUD) GetSelectedItem() item.ItemStack {
	if h.Hotbar == nil {
		return item.ItemStack{}
	}
	return h.Hotbar.GetSlot(h.SelectedSlot)
}

// SetSpawnPoint updates the world spawn coordinates used by the compass overlay.
func (h *HUD) SetSpawnPoint(x, z float32) {
	h.SpawnX = x
	h.SpawnZ = z
}

// SetGameTick updates the current game tick used by the clock overlay.
func (h *HUD) SetGameTick(tick int64) {
	h.GameTick = tick
}

// drawFunctionalItemOverlay renders context text above the hotbar when the
// selected item is a compass or clock.
func (h *HUD) drawFunctionalItemOverlay(r *UIRenderer) {
	selected := h.GetSelectedItem()
	if selected.IsEmpty() {
		return
	}
	props := item.GetProperties(selected.ItemID)

	var text string
	switch {
	case props.IsCompass:
		text = fmt.Sprintf("-> Spawn: %.0f %.0f", h.SpawnX, h.SpawnZ)
	case props.IsClock:
		hour, minute := item.TickToHoursMinutes(h.GameTick)
		dayNight := "Day"
		if !item.TickIsDaytime(h.GameTick) {
			dayNight = "Night"
		}
		text = fmt.Sprintf("Time: %s (%02d:%02d)", dayNight, hour, minute)
	default:
		return
	}

	totalWidth := float32(hotbarSlots)*slotSize + float32(hotbarSlots-1)*slotPadding
	startX := (r.ScreenWidth - totalWidth) / 2
	y := r.ScreenHeight - hudBarHeight - heartSize - 24
	r.DrawText(startX, y, text, 1.0, 1, 1, 1)
}
