package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	ctGridSize = 3 // 3x3 crafting table grid
)

// CraftingTableScreen displays a full 3x3 crafting grid with a result slot
// and the player inventory (36 slots, 4x9) at the bottom.
type CraftingTableScreen struct {
	// CraftGrid is the 3x3 crafting grid.
	CraftGrid *inventory.CraftingGrid

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

// NewCraftingTableScreen creates a crafting table screen bound to the given
// crafting grid and player inventory.
func NewCraftingTableScreen(grid *inventory.CraftingGrid, playerInv *inventory.Inventory, onClose func()) *CraftingTableScreen {
	return &CraftingTableScreen{
		CraftGrid: grid,
		PlayerInv: playerInv,
		onClose:   onClose,
	}
}

// Update handles input for the crafting table screen.
func (s *CraftingTableScreen) Update(inp *input.Manager, _ float64) {
	s.mouseX, s.mouseY = inp.MousePos()

	if inp.IsKeyJustPressed(input.KeyE) || inp.IsKeyJustPressed(input.KeyEscape) {
		s.Close()
		return
	}

	if inp.IsMouseJustPressed(input.MouseButtonLeft) {
		s.handleClick(float32(s.mouseX), float32(s.mouseY))
	}

	if inp.IsMouseJustPressed(input.MouseButtonRight) {
		s.handleRightClick(float32(s.mouseX), float32(s.mouseY))
	}
}

// Draw renders the crafting table screen.
func (s *CraftingTableScreen) Draw(r *UIRenderer) {
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.6)

	panelWidth, gridAreaHeight, _, baseX, baseY := s.ctLayout(r)

	// Crafting panel background.
	r.DrawRect(baseX, baseY, panelWidth, gridAreaHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, baseY+4, "Crafting", 0.8, 0.9, 0.9, 0.9)

	// 3x3 crafting grid.
	gridBaseX, gridBaseY := s.craftGridOrigin(r)
	for row := 0; row < ctGridSize; row++ {
		for col := 0; col < ctGridSize; col++ {
			sx := gridBaseX + float32(col)*(invSlotSize+invSlotPadding)
			sy := gridBaseY + float32(row)*(invSlotSize+invSlotPadding)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.25, 0.25, 0.25, 0.8)
			if s.CraftGrid != nil {
				stack := s.CraftGrid.GetSlot(row, col)
				if !stack.IsEmpty() {
					r.DrawItemSlot(sx+2, sy+2, stack)
				}
			}
		}
	}

	// Arrow.
	arrowX := gridBaseX + float32(ctGridSize)*(invSlotSize+invSlotPadding) + 8
	arrowY := gridBaseY + float32(ctGridSize-1)*(invSlotSize+invSlotPadding)/2
	r.DrawText(arrowX, arrowY, "=>", 1.0, 1, 1, 1)

	// Result slot.
	resultX, resultY := s.resultSlotPos(r)
	r.DrawRect(resultX, resultY, invSlotSize, invSlotSize, 0.3, 0.3, 0.15, 0.8)
	if s.CraftGrid != nil {
		result := s.CraftGrid.GetResult()
		if !result.IsEmpty() {
			r.DrawItemSlot(resultX+2, resultY+2, result)
		}
	}

	// Player inventory panel.
	pHeight := playerPanelHeight()
	playerBaseY := baseY + gridAreaHeight + containerPanelGap
	r.DrawRect(baseX, playerBaseY, panelWidth, pHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, playerBaseY+4, "Inventory", 0.8, 0.9, 0.9, 0.9)

	for row := 0; row < invRows; row++ {
		for col := 0; col < invCols; col++ {
			sx, sy := s.playerSlotPos(r, row, col)
			r.DrawRect(sx, sy, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
			if s.PlayerInv != nil {
				stack := s.PlayerInv.GetSlot(row*invCols + col)
				if !stack.IsEmpty() {
					r.DrawItemSlot(sx+2, sy+2, stack)
				}
			}
		}
	}

	if !s.HeldItem.IsEmpty() {
		r.DrawItemSlot(float32(s.mouseX)-invSlotSize/2, float32(s.mouseY)-invSlotSize/2, s.HeldItem)
	}
}

// IsOverlay returns true.
func (s *CraftingTableScreen) IsOverlay() bool { return true }

// HandleKey processes a raw key code.
func (s *CraftingTableScreen) HandleKey(key int) {
	if key == input.KeyE || key == input.KeyEscape {
		s.Close()
	}
}

// IsClosed reports whether the screen has been closed.
func (s *CraftingTableScreen) IsClosed() bool { return s.closed }

// Close closes the crafting table screen, returning any held item to the
// player inventory and clearing the crafting grid back into the inventory.
func (s *CraftingTableScreen) Close() {
	if s.closed {
		return
	}
	s.closed = true

	if !s.HeldItem.IsEmpty() && s.PlayerInv != nil {
		s.PlayerInv.AddItem(s.HeldItem)
		s.HeldItem = item.ItemStack{}
	}

	if s.CraftGrid != nil && s.PlayerInv != nil {
		for row := 0; row < ctGridSize; row++ {
			for col := 0; col < ctGridSize; col++ {
				stack := s.CraftGrid.GetSlot(row, col)
				if !stack.IsEmpty() {
					s.PlayerInv.AddItem(stack)
					s.CraftGrid.SetSlot(row, col, item.ItemStack{})
				}
			}
		}
	}

	if s.onClose != nil {
		s.onClose()
	}
}

// SetScreenSize sets the screen dimensions for hit testing.
func (s *CraftingTableScreen) SetScreenSize(w, h float32) {
	s.screenWidth = w
	s.screenHeight = h
}

// handleClick processes a click at screen coordinates.
func (s *CraftingTableScreen) handleClick(mx, my float32) {
	cr, cc := s.hitTestCraftGrid(mx, my)
	if cr >= 0 && cc >= 0 {
		s.swapWithCraftSlot(cr, cc)
		return
	}

	if s.hitTestResult(mx, my) {
		s.takeCraftResult()
		return
	}

	playerIdx := s.hitTestPlayer(mx, my)
	if playerIdx >= 0 {
		s.swapWithPlayerSlot(playerIdx)
	}
}

// swapWithCraftSlot swaps the held item with a crafting grid slot.
func (s *CraftingTableScreen) swapWithCraftSlot(row, col int) {
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

	s.CraftGrid.SetSlot(row, col, s.HeldItem)
	s.HeldItem = current
}

// takeCraftResult picks up the crafting result if hands are empty.
func (s *CraftingTableScreen) takeCraftResult() {
	if s.CraftGrid == nil || !s.HeldItem.IsEmpty() {
		return
	}
	result, ok := s.CraftGrid.Craft()
	if ok {
		s.HeldItem = result
	}
}

// swapWithPlayerSlot swaps the held item with a player inventory slot.
func (s *CraftingTableScreen) swapWithPlayerSlot(slotIdx int) {
	s.HeldItem = swapHeldWithSlot(s.HeldItem, s.PlayerInv, slotIdx)
}

// handleRightClick processes a right-click at screen coordinates.
// Grid/player slots: holding nothing -> pick up half; holding items -> place 1.
// Result slot: same as left-click (take result).
func (s *CraftingTableScreen) handleRightClick(mx, my float32) {
	cr, cc := s.hitTestCraftGrid(mx, my)
	if cr >= 0 && cc >= 0 {
		s.rightClickCraftSlot(cr, cc)
		return
	}

	if s.hitTestResult(mx, my) {
		s.takeCraftResult()
		return
	}

	playerIdx := s.hitTestPlayer(mx, my)
	if playerIdx >= 0 {
		s.rightClickPlayerSlot(playerIdx)
	}
}

// rightClickCraftSlot handles right-click on a crafting grid slot.
// If not holding anything, picks up half the stack. If holding items,
// places one item into the slot (if compatible or empty).
func (s *CraftingTableScreen) rightClickCraftSlot(row, col int) {
	if s.CraftGrid == nil {
		return
	}
	current := s.CraftGrid.GetSlot(row, col)

	if s.HeldItem.IsEmpty() {
		// Pick up half (rounded up).
		if current.IsEmpty() {
			return
		}
		half := (current.Count + 1) / 2
		taken, remaining := current.Split(half)
		s.HeldItem = taken
		s.CraftGrid.SetSlot(row, col, remaining)
		return
	}

	// Holding items: place one into the slot.
	if current.IsEmpty() {
		placed, remaining := s.HeldItem.Split(1)
		s.CraftGrid.SetSlot(row, col, placed)
		s.HeldItem = remaining
		return
	}

	if current.CanStackWith(s.HeldItem) {
		max := item.MaxStack(current.ItemID)
		if current.Count < max {
			placed := current
			placed.Count++
			s.CraftGrid.SetSlot(row, col, placed)
			_, remaining := s.HeldItem.Split(1)
			s.HeldItem = remaining
		}
	}
}

// rightClickPlayerSlot handles right-click on a player inventory slot.
// If not holding anything, picks up half the stack. If holding items,
// places one item into the slot (if compatible or empty).
func (s *CraftingTableScreen) rightClickPlayerSlot(slotIdx int) {
	if s.PlayerInv == nil {
		return
	}
	current := s.PlayerInv.GetSlot(slotIdx)

	if s.HeldItem.IsEmpty() {
		// Pick up half (rounded up).
		if current.IsEmpty() {
			return
		}
		half := (current.Count + 1) / 2
		taken, remaining := current.Split(half)
		s.HeldItem = taken
		s.PlayerInv.SetSlot(slotIdx, remaining)
		return
	}

	// Holding items: place one into the slot.
	if current.IsEmpty() {
		placed, remaining := s.HeldItem.Split(1)
		s.PlayerInv.SetSlot(slotIdx, placed)
		s.HeldItem = remaining
		return
	}

	if current.CanStackWith(s.HeldItem) {
		max := item.MaxStack(current.ItemID)
		if current.Count < max {
			placed := current
			placed.Count++
			s.PlayerInv.SetSlot(slotIdx, placed)
			_, remaining := s.HeldItem.Split(1)
			s.HeldItem = remaining
		}
	}
}

// --- Layout helpers ---

func (s *CraftingTableScreen) ctLayout(r *UIRenderer) (panelWidth, gridAreaHeight, totalHeight, baseX, baseY float32) {
	panelWidth = playerPanelWidth()
	gridAreaHeight = float32(ctGridSize)*(invSlotSize+invSlotPadding) - invSlotPadding + 2*invBgPadding + 20
	pHeight := playerPanelHeight()
	totalHeight = gridAreaHeight + containerPanelGap + pHeight
	baseX = (r.ScreenWidth - panelWidth) / 2
	baseY = (r.ScreenHeight - totalHeight) / 2
	return
}

func (s *CraftingTableScreen) craftGridOrigin(r *UIRenderer) (float32, float32) {
	_, _, _, baseX, baseY := s.ctLayout(r)
	return baseX + invBgPadding, baseY + invBgPadding + 20
}

func (s *CraftingTableScreen) resultSlotPos(r *UIRenderer) (float32, float32) {
	gridBaseX, gridBaseY := s.craftGridOrigin(r)
	arrowX := gridBaseX + float32(ctGridSize)*(invSlotSize+invSlotPadding) + 8
	resultX := arrowX + 32
	resultY := gridBaseY + float32(ctGridSize-1)*(invSlotSize+invSlotPadding)/2 - invSlotSize/2 + invSlotSize/4
	return resultX, resultY
}

func (s *CraftingTableScreen) playerSlotPos(r *UIRenderer, row, col int) (float32, float32) {
	_, gridAreaH, _, baseX, baseY := s.ctLayout(r)
	playerBaseY := baseY + gridAreaH + containerPanelGap + invBgPadding
	sx := baseX + invBgPadding + float32(col)*(invSlotSize+invSlotPadding)
	sy := playerBaseY + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

// --- Hit testing ---

func (s *CraftingTableScreen) hitTestCraftGrid(mx, my float32) (int, int) {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	gridBaseX, gridBaseY := s.craftGridOrigin(r)
	for row := 0; row < ctGridSize; row++ {
		for col := 0; col < ctGridSize; col++ {
			sx := gridBaseX + float32(col)*(invSlotSize+invSlotPadding)
			sy := gridBaseY + float32(row)*(invSlotSize+invSlotPadding)
			if mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize {
				return row, col
			}
		}
	}
	return -1, -1
}

func (s *CraftingTableScreen) hitTestResult(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	resultX, resultY := s.resultSlotPos(r)
	return mx >= resultX && mx <= resultX+invSlotSize && my >= resultY && my <= resultY+invSlotSize
}

func (s *CraftingTableScreen) hitTestPlayer(mx, my float32) int {
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
