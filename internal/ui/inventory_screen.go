package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	invRows        = 4
	invCols        = 9
	invTotalSlots  = invRows * invCols // 36
	invSlotSize    = 36.0
	invSlotPadding = 4.0
	craftGridSize  = 2
	invBgPadding   = 16.0
)

// InventoryScreen displays the player inventory (36 slots in a 4x9 grid)
// with a 2x2 crafting grid and a crafting result slot. Items can be
// dragged and dropped between slots.
type InventoryScreen struct {
	// Inv is the player inventory (36 slots).
	Inv *inventory.Inventory

	// CraftGrid is the 2x2 (uses top-left of 3x3) crafting grid.
	CraftGrid *inventory.CraftingGrid

	// HeldItem is the item stack currently held by the cursor.
	HeldItem item.ItemStack

	// mouseX and mouseY track the last known cursor position.
	mouseX, mouseY float64

	// closed indicates the screen should be removed.
	closed bool

	// onClose is called when the screen closes, so the manager can pop it.
	onClose func()

	// screenWidth, screenHeight for hit testing when no renderer is available.
	screenWidth, screenHeight float32
}

// NewInventoryScreen creates an inventory screen bound to the given
// inventory and crafting grid.
func NewInventoryScreen(inv *inventory.Inventory, grid *inventory.CraftingGrid, onClose func()) *InventoryScreen {
	return &InventoryScreen{
		Inv:       inv,
		CraftGrid: grid,
		onClose:   onClose,
	}
}

// Update handles input for the inventory screen: clicking slots, closing
// with E or Escape.
func (s *InventoryScreen) Update(inp *input.Manager, _ float64) {
	s.mouseX, s.mouseY = inp.MousePos()

	// Close on E or Escape.
	if inp.IsKeyJustPressed(input.KeyE) || inp.IsKeyJustPressed(input.KeyEscape) {
		s.Close()
		return
	}

	// Handle left-click for slot interaction.
	if inp.IsMouseJustPressed(input.MouseButtonLeft) {
		if inp.IsKeyDown(input.KeyLeftShift) {
			s.handleShiftClick(float32(s.mouseX), float32(s.mouseY))
		} else {
			s.handleClick(float32(s.mouseX), float32(s.mouseY))
		}
	}
}

// Draw renders the inventory screen: background, inventory grid, crafting
// grid, crafting result, and the held item following the cursor.
func (s *InventoryScreen) Draw(r *UIRenderer) {
	// Semi-transparent background overlay.
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.6)

	// Inventory grid panel.
	invWidth := float32(invCols)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	invHeight := float32(invRows)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	invX := (r.ScreenWidth - invWidth) / 2
	invY := r.ScreenHeight/2 - invBgPadding

	// Panel background.
	r.DrawRect(invX, invY, invWidth, invHeight, 0.4, 0.4, 0.4, 0.9)

	// Draw inventory slots.
	for row := 0; row < invRows; row++ {
		for col := 0; col < invCols; col++ {
			slotIdx := row*invCols + col
			sx, sy := s.inventorySlotPos(r, row, col)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
			stack := s.Inv.GetSlot(slotIdx)
			if !stack.IsEmpty() {
				r.DrawItemSlot(sx+2, sy+2, stack)
			}
		}
	}

	// Crafting grid (2x2) — positioned above and to the right of inventory.
	s.drawCraftingGrid(r)

	// Held item follows cursor.
	if !s.HeldItem.IsEmpty() {
		r.DrawItemSlot(float32(s.mouseX)-invSlotSize/2, float32(s.mouseY)-invSlotSize/2, s.HeldItem)
	}
}

// IsOverlay returns true: the inventory is drawn on top of the HUD.
func (s *InventoryScreen) IsOverlay() bool {
	return true
}

// HandleKey processes a raw key code.
func (s *InventoryScreen) HandleKey(key int) {
	if key == input.KeyE || key == input.KeyEscape {
		s.Close()
	}
}

// IsClosed reports whether the screen has been closed.
func (s *InventoryScreen) IsClosed() bool {
	return s.closed
}

// Close closes the inventory screen, dropping any held item back into the
// inventory.
func (s *InventoryScreen) Close() {
	if s.closed {
		return
	}
	s.closed = true
	// Return held item to inventory.
	if !s.HeldItem.IsEmpty() && s.Inv != nil {
		s.Inv.AddItem(s.HeldItem)
		s.HeldItem = item.ItemStack{}
	}
	if s.onClose != nil {
		s.onClose()
	}
}

// handleClick processes a click at screen coordinates (mx, my).
func (s *InventoryScreen) handleClick(mx, my float32) {
	// Check inventory slots.
	slotIdx := s.hitTestInventory(mx, my)
	if slotIdx >= 0 {
		s.swapWithSlot(slotIdx)
		return
	}

	// Check crafting grid slots.
	cr, cc := s.hitTestCraftGrid(mx, my)
	if cr >= 0 && cc >= 0 {
		s.swapWithCraftSlot(cr, cc)
		return
	}

	// Check crafting result slot.
	if s.hitTestCraftResult(mx, my) {
		s.takeCraftResult()
	}
}

// handleShiftClick processes a shift-click at screen coordinates (mx, my).
// Shift-clicking moves items directly into the player inventory without
// picking them up on the cursor.
func (s *InventoryScreen) handleShiftClick(mx, my float32) {
	// Check crafting result slot: craft and add result to inventory.
	if s.hitTestCraftResult(mx, my) {
		if s.CraftGrid == nil || s.Inv == nil {
			return
		}
		result, ok := s.CraftGrid.Craft()
		if ok {
			s.Inv.AddItem(result)
		}
		return
	}

	// Check crafting grid slots: move item to inventory.
	cr, cc := s.hitTestCraftGrid(mx, my)
	if cr >= 0 && cc >= 0 {
		if s.CraftGrid == nil || s.Inv == nil {
			return
		}
		stack := s.CraftGrid.GetSlot(cr, cc)
		if !stack.IsEmpty() {
			s.CraftGrid.SetSlot(cr, cc, item.ItemStack{})
			s.Inv.AddItem(stack)
		}
		return
	}

	// Check player inventory slots: move between hotbar (row 0) and main
	// inventory (rows 1-3).
	slotIdx := s.hitTestInventory(mx, my)
	if slotIdx >= 0 && s.Inv != nil {
		stack := s.Inv.GetSlot(slotIdx)
		if stack.IsEmpty() {
			return
		}
		hotbarEnd := invCols // slots 0..8 are row 0 (hotbar)
		if slotIdx < hotbarEnd {
			// In hotbar: move to first available main inventory slot (rows 1-3).
			s.Inv.SetSlot(slotIdx, item.ItemStack{})
			remainder := s.addItemToRange(stack, hotbarEnd, invTotalSlots)
			if !remainder.IsEmpty() {
				// Could not fit; put it back.
				s.Inv.SetSlot(slotIdx, remainder)
			}
		} else {
			// In main inventory: move to first available hotbar slot (row 0).
			s.Inv.SetSlot(slotIdx, item.ItemStack{})
			remainder := s.addItemToRange(stack, 0, hotbarEnd)
			if !remainder.IsEmpty() {
				s.Inv.SetSlot(slotIdx, remainder)
			}
		}
	}
}

// addItemToRange tries to add a stack into inventory slots [start, end).
// It first merges with compatible stacks, then fills empty slots.
// Returns whatever could not fit.
func (s *InventoryScreen) addItemToRange(stack item.ItemStack, start, end int) item.ItemStack {
	// First pass: merge into existing compatible stacks.
	for i := start; i < end; i++ {
		current := s.Inv.GetSlot(i)
		if current.IsEmpty() {
			continue
		}
		if current.CanStackWith(stack) {
			merged := current
			stack = merged.Merge(stack)
			s.Inv.SetSlot(i, merged)
			if stack.IsEmpty() {
				return stack
			}
		}
	}
	// Second pass: place into empty slots.
	for i := start; i < end; i++ {
		if s.Inv.GetSlot(i).IsEmpty() {
			s.Inv.SetSlot(i, stack)
			return item.ItemStack{}
		}
	}
	return stack
}

// swapWithSlot swaps the held item with the item in the given inventory slot.
func (s *InventoryScreen) swapWithSlot(slotIdx int) {
	if s.Inv == nil {
		return
	}
	current := s.Inv.GetSlot(slotIdx)

	// If holding nothing, pick up the slot's item.
	if s.HeldItem.IsEmpty() {
		s.HeldItem = current
		s.Inv.SetSlot(slotIdx, item.ItemStack{})
		return
	}

	// If the slot is empty, place the held item.
	if current.IsEmpty() {
		s.Inv.SetSlot(slotIdx, s.HeldItem)
		s.HeldItem = item.ItemStack{}
		return
	}

	// If same item type, try to merge.
	if s.HeldItem.CanStackWith(current) {
		merged := current
		remaining := merged.Merge(s.HeldItem)
		s.Inv.SetSlot(slotIdx, merged)
		s.HeldItem = remaining
		return
	}

	// Different items: swap.
	s.Inv.SetSlot(slotIdx, s.HeldItem)
	s.HeldItem = current
}

// swapWithCraftSlot swaps the held item with a crafting grid slot.
func (s *InventoryScreen) swapWithCraftSlot(row, col int) {
	if s.CraftGrid == nil {
		return
	}
	current := s.CraftGrid.GetSlot(row, col)

	if s.HeldItem.IsEmpty() {
		s.HeldItem = current
		s.CraftGrid.SetSlot(row, col, item.ItemStack{})
		return
	}

	if current.IsEmpty() {
		s.CraftGrid.SetSlot(row, col, s.HeldItem)
		s.HeldItem = item.ItemStack{}
		return
	}

	// Swap.
	s.CraftGrid.SetSlot(row, col, s.HeldItem)
	s.HeldItem = current
}

// takeCraftResult picks up the crafting result if one exists.
func (s *InventoryScreen) takeCraftResult() {
	if s.CraftGrid == nil {
		return
	}
	// Only pick up if not already holding something.
	if !s.HeldItem.IsEmpty() {
		return
	}
	result, ok := s.CraftGrid.Craft()
	if ok {
		s.HeldItem = result
	}
}

// drawCraftingGrid renders the 2x2 crafting grid and result slot.
func (s *InventoryScreen) drawCraftingGrid(r *UIRenderer) {
	baseX, baseY := s.craftGridOrigin(r)

	// Label.
	r.DrawText(baseX, baseY-20, "Crafting", 0.8, 0.9, 0.9, 0.9)

	// 2x2 grid slots.
	for row := 0; row < craftGridSize; row++ {
		for col := 0; col < craftGridSize; col++ {
			sx := baseX + float32(col)*(invSlotSize+invSlotPadding)
			sy := baseY + float32(row)*(invSlotSize+invSlotPadding)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.25, 0.25, 0.25, 0.8)
			stack := s.CraftGrid.GetSlot(row, col)
			if !stack.IsEmpty() {
				r.DrawItemSlot(sx+2, sy+2, stack)
			}
		}
	}

	// Arrow.
	arrowX := baseX + float32(craftGridSize)*(invSlotSize+invSlotPadding) + 8
	arrowY := baseY + (invSlotSize+invSlotPadding)/2
	r.DrawText(arrowX, arrowY, "=>", 1.0, 1, 1, 1)

	// Result slot.
	resultX := arrowX + 32
	resultY := baseY + (invSlotSize+invSlotPadding)/2 - invSlotSize/2 + invSlotSize/4
	r.DrawRect(resultX, resultY, invSlotSize, invSlotSize, 0.3, 0.3, 0.15, 0.8)
	if s.CraftGrid != nil {
		result := s.CraftGrid.GetResult()
		if !result.IsEmpty() {
			r.DrawItemSlot(resultX+2, resultY+2, result)
		}
	}
}

// craftGridOrigin returns the top-left corner of the crafting grid area.
func (s *InventoryScreen) craftGridOrigin(r *UIRenderer) (float32, float32) {
	invWidth := float32(invCols)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	invX := (r.ScreenWidth - invWidth) / 2
	baseX := invX + invWidth - float32(craftGridSize)*(invSlotSize+invSlotPadding) - invBgPadding - 60
	baseY := r.ScreenHeight/2 - invBgPadding - float32(craftGridSize)*(invSlotSize+invSlotPadding) - 30
	return baseX, baseY
}

// inventorySlotPos returns the screen position for an inventory grid slot.
func (s *InventoryScreen) inventorySlotPos(r *UIRenderer, row, col int) (float32, float32) {
	invWidth := float32(invCols)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding
	invX := (r.ScreenWidth-invWidth)/2 + invBgPadding
	invY := r.ScreenHeight/2 - invBgPadding + invBgPadding
	sx := invX + float32(col)*(invSlotSize+invSlotPadding)
	sy := invY + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

// hitTestInventory returns the inventory slot index under (mx, my), or -1.
func (s *InventoryScreen) hitTestInventory(mx, my float32) int {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	for row := 0; row < invRows; row++ {
		for col := 0; col < invCols; col++ {
			sx, sy := s.inventorySlotPos(r, row, col)
			if mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize {
				return row*invCols + col
			}
		}
	}
	return -1
}

// SetScreenSize sets the screen dimensions used for hit testing.
func (s *InventoryScreen) SetScreenSize(w, h float32) {
	s.screenWidth = w
	s.screenHeight = h
}

// hitTestCraftGrid returns the crafting grid (row, col) under (mx, my),
// or (-1, -1) if no slot is hit.
func (s *InventoryScreen) hitTestCraftGrid(mx, my float32) (int, int) {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	baseX, baseY := s.craftGridOrigin(r)
	for row := 0; row < craftGridSize; row++ {
		for col := 0; col < craftGridSize; col++ {
			sx := baseX + float32(col)*(invSlotSize+invSlotPadding)
			sy := baseY + float32(row)*(invSlotSize+invSlotPadding)
			if mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize {
				return row, col
			}
		}
	}
	return -1, -1
}

// hitTestCraftResult reports whether (mx, my) hits the crafting result slot.
func (s *InventoryScreen) hitTestCraftResult(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	baseX, baseY := s.craftGridOrigin(r)
	arrowX := baseX + float32(craftGridSize)*(invSlotSize+invSlotPadding) + 8
	resultX := arrowX + 32
	resultY := baseY + (invSlotSize+invSlotPadding)/2 - invSlotSize/2 + invSlotSize/4
	return mx >= resultX && mx <= resultX+invSlotSize && my >= resultY && my <= resultY+invSlotSize
}
