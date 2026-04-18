package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	chestRows       = 3
	chestCols       = 9
	chestTotalSlots = chestRows * chestCols // 27
)

// ChestScreen displays a chest inventory (27 slots, 3x9) at the top and the
// player inventory (36 slots, 4x9) at the bottom with drag-and-drop between
// the two.
type ChestScreen struct {
	// ChestInv is the chest's inventory (27 slots).
	ChestInv *inventory.Inventory

	// PlayerInv is the player's inventory (36 slots).
	PlayerInv *inventory.Inventory

	// HeldItem is the item stack currently held by the cursor.
	HeldItem item.ItemStack

	mouseX, mouseY float64
	closed         bool
	onClose        func()
	screenWidth    float32
	screenHeight   float32
}

// NewChestScreen creates a chest screen bound to the given chest and player
// inventories.
func NewChestScreen(chestInv, playerInv *inventory.Inventory, onClose func()) *ChestScreen {
	return &ChestScreen{
		ChestInv:  chestInv,
		PlayerInv: playerInv,
		onClose:   onClose,
	}
}

// Update handles input for the chest screen.
func (s *ChestScreen) Update(inp *input.Manager, _ float64) {
	s.mouseX, s.mouseY = inp.MousePos()

	if inp.IsKeyJustPressed(input.KeyE) || inp.IsKeyJustPressed(input.KeyEscape) {
		s.Close()
		return
	}

	if inp.IsMouseJustPressed(input.MouseButtonLeft) {
		s.handleClick(float32(s.mouseX), float32(s.mouseY))
	}
}

// Draw renders the chest screen.
func (s *ChestScreen) Draw(r *UIRenderer) {
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.6)

	gridWidth := playerPanelWidth()
	chestHeight := float32(chestRows)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	pHeight := playerPanelHeight()
	totalHeight := chestHeight + containerPanelGap + pHeight

	baseX := (r.ScreenWidth - gridWidth) / 2
	baseY := (r.ScreenHeight - totalHeight) / 2

	// Chest panel.
	r.DrawRect(baseX, baseY, gridWidth, chestHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, baseY+4, "Chest", 0.8, 0.9, 0.9, 0.9)

	for row := 0; row < chestRows; row++ {
		for col := 0; col < chestCols; col++ {
			sx, sy := s.chestSlotPos(r, row, col)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
			stack := s.ChestInv.GetSlot(row*chestCols + col)
			if !stack.IsEmpty() {
				r.DrawItemSlot(sx+2, sy+2, stack)
			}
		}
	}

	// Player inventory panel.
	playerY := baseY + chestHeight + containerPanelGap
	r.DrawRect(baseX, playerY, gridWidth, pHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, playerY+4, "Inventory", 0.8, 0.9, 0.9, 0.9)

	for row := 0; row < invRows; row++ {
		for col := 0; col < invCols; col++ {
			sx, sy := s.playerSlotPos(r, row, col)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
			stack := s.PlayerInv.GetSlot(row*invCols + col)
			if !stack.IsEmpty() {
				r.DrawItemSlot(sx+2, sy+2, stack)
			}
		}
	}

	if !s.HeldItem.IsEmpty() {
		r.DrawItemSlot(float32(s.mouseX)-invSlotSize/2, float32(s.mouseY)-invSlotSize/2, s.HeldItem)
	}
}

// IsOverlay returns true.
func (s *ChestScreen) IsOverlay() bool { return true }

// HandleKey processes a raw key code.
func (s *ChestScreen) HandleKey(key int) {
	if key == input.KeyE || key == input.KeyEscape {
		s.Close()
	}
}

// IsClosed reports whether the screen has been closed.
func (s *ChestScreen) IsClosed() bool { return s.closed }

// Close closes the chest screen, returning any held item to the player
// inventory.
func (s *ChestScreen) Close() {
	if s.closed {
		return
	}
	s.closed = true
	if !s.HeldItem.IsEmpty() && s.PlayerInv != nil {
		s.PlayerInv.AddItem(s.HeldItem)
		s.HeldItem = item.ItemStack{}
	}
	if s.onClose != nil {
		s.onClose()
	}
}

// SetScreenSize sets the screen dimensions for hit testing.
func (s *ChestScreen) SetScreenSize(w, h float32) {
	s.screenWidth = w
	s.screenHeight = h
}

// handleClick processes a click at screen coordinates (mx, my).
func (s *ChestScreen) handleClick(mx, my float32) {
	chestIdx := s.hitTestChest(mx, my)
	if chestIdx >= 0 {
		s.HeldItem = swapHeldWithSlot(s.HeldItem, s.ChestInv, chestIdx)
		return
	}

	playerIdx := s.hitTestPlayer(mx, my)
	if playerIdx >= 0 {
		s.HeldItem = swapHeldWithSlot(s.HeldItem, s.PlayerInv, playerIdx)
	}
}

// swapWithInventory swaps the held item with a slot in the given inventory.
// Kept as a convenience method for direct testing.
func (s *ChestScreen) swapWithInventory(inv *inventory.Inventory, slotIdx int) {
	s.HeldItem = swapHeldWithSlot(s.HeldItem, inv, slotIdx)
}

// --- Layout helpers ---

func (s *ChestScreen) chestLayout(r *UIRenderer) (gridWidth, chestHeight, totalHeight, baseX, baseY float32) {
	gridWidth = playerPanelWidth()
	chestHeight = float32(chestRows)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	pHeight := playerPanelHeight()
	totalHeight = chestHeight + containerPanelGap + pHeight
	baseX = (r.ScreenWidth - gridWidth) / 2
	baseY = (r.ScreenHeight - totalHeight) / 2
	return
}

func (s *ChestScreen) chestSlotPos(r *UIRenderer, row, col int) (float32, float32) {
	_, _, _, baseX, baseY := s.chestLayout(r)
	sx := baseX + invBgPadding + float32(col)*(invSlotSize+invSlotPadding)
	sy := baseY + invBgPadding + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

func (s *ChestScreen) playerSlotPos(r *UIRenderer, row, col int) (float32, float32) {
	_, chestH, _, baseX, baseY := s.chestLayout(r)
	playerBaseY := baseY + chestH + containerPanelGap + invBgPadding
	sx := baseX + invBgPadding + float32(col)*(invSlotSize+invSlotPadding)
	sy := playerBaseY + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

func (s *ChestScreen) hitTestChest(mx, my float32) int {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	for row := 0; row < chestRows; row++ {
		for col := 0; col < chestCols; col++ {
			sx, sy := s.chestSlotPos(r, row, col)
			if mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize {
				return row*chestCols + col
			}
		}
	}
	return -1
}

func (s *ChestScreen) hitTestPlayer(mx, my float32) int {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	for row := 0; row < invRows; row++ {
		for col := 0; col < invCols; col++ {
			sx, sy := s.playerSlotPos(r, row, col)
			if mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize {
				return row*invCols + col
			}
		}
	}
	return -1
}
