package ui

import (
	"fmt"

	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

// FurnaceScreen displays a furnace GUI with input, fuel, and output slots,
// a progress arrow, a flame indicator, and the player inventory at the bottom.
type FurnaceScreen struct {
	// Furnace is the furnace state (input, fuel, output, progress, burn time).
	Furnace *inventory.Furnace

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

// NewFurnaceScreen creates a furnace screen bound to the given furnace and
// player inventory.
func NewFurnaceScreen(furnace *inventory.Furnace, playerInv *inventory.Inventory, onClose func()) *FurnaceScreen {
	return &FurnaceScreen{
		Furnace:   furnace,
		PlayerInv: playerInv,
		onClose:   onClose,
	}
}

// Update handles input for the furnace screen.
func (s *FurnaceScreen) Update(inp *input.Manager, _ float64) {
	s.mouseX, s.mouseY = inp.MousePos()

	if inp.IsKeyJustPressed(input.KeyE) || inp.IsKeyJustPressed(input.KeyEscape) {
		s.Close()
		return
	}

	if inp.IsMouseJustPressed(input.MouseButtonLeft) {
		s.handleClick(float32(s.mouseX), float32(s.mouseY))
	}
}

// Draw renders the furnace screen.
func (s *FurnaceScreen) Draw(r *UIRenderer) {
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.6)

	panelWidth, furnaceAreaHeight, _, _, baseX, baseY := s.furnaceLayout(r)

	// Furnace panel background.
	r.DrawRect(baseX, baseY, panelWidth, furnaceAreaHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, baseY+4, "Furnace", 0.8, 0.9, 0.9, 0.9)

	centerX := baseX + panelWidth/2

	// Input slot (top center).
	inputX := centerX - invSlotSize/2
	inputY := baseY + invBgPadding + 10
	r.DrawRect(inputX, inputY, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
	if s.Furnace != nil && !s.Furnace.InputSlot.IsEmpty() {
		r.DrawItemSlot(inputX+2, inputY+2, s.Furnace.InputSlot)
	}

	// Flame icon (below input, left of center).
	flameX := centerX - invSlotSize - 10
	flameY := inputY + invSlotSize + invSlotPadding + 4
	s.drawFlame(r, flameX, flameY)

	// Fuel slot (below flame).
	fuelX := flameX
	fuelY := flameY + invSlotSize + invSlotPadding
	r.DrawRect(fuelX, fuelY, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
	if s.Furnace != nil && !s.Furnace.FuelSlot.IsEmpty() {
		r.DrawItemSlot(fuelX+2, fuelY+2, s.Furnace.FuelSlot)
	}

	// Progress arrow (right of center, vertically between input and fuel).
	arrowX := centerX + 10
	arrowY := inputY + invSlotSize + invSlotPadding + 4
	s.drawProgressArrow(r, arrowX, arrowY)

	// Output slot (right of arrow).
	outputX := arrowX + invSlotSize + 10
	outputY := arrowY
	r.DrawRect(outputX, outputY, invSlotSize, invSlotSize, 0.3, 0.3, 0.15, 0.8)
	if s.Furnace != nil && !s.Furnace.OutputSlot.IsEmpty() {
		r.DrawItemSlot(outputX+2, outputY+2, s.Furnace.OutputSlot)
	}

	// Player inventory panel.
	_, _, pHeight, _, _, _ := s.furnaceLayout(r)
	playerBaseY := baseY + furnaceAreaHeight + containerPanelGap
	r.DrawRect(baseX, playerBaseY, panelWidth, pHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, playerBaseY+4, "Inventory", 0.8, 0.9, 0.9, 0.9)

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

// drawFlame renders a flame indicator. The fill level is based on BurnTime.
func (s *FurnaceScreen) drawFlame(r *UIRenderer, x, y float32) {
	r.DrawRect(x, y, invSlotSize, invSlotSize, 0.15, 0.15, 0.15, 0.8)

	if s.Furnace == nil {
		return
	}

	fill := float32(0)
	if s.Furnace.BurnTime > 0 {
		fill = float32(s.Furnace.BurnTime / inventory.DefaultFuelBurnTime)
		if fill > 1 {
			fill = 1
		}
	}

	if fill > 0 {
		fillHeight := invSlotSize * fill
		fillY := y + invSlotSize - fillHeight
		r.DrawRect(x, fillY, invSlotSize, fillHeight, 0.9, 0.4, 0.1, 0.9)
	}

	burnPct := int(fill * 100)
	r.DrawText(x+4, y+invSlotSize/2-6, fmt.Sprintf("%d%%", burnPct), 0.6, 1, 1, 1)
}

// drawProgressArrow renders a progress arrow. The fill is based on Progress.
func (s *FurnaceScreen) drawProgressArrow(r *UIRenderer, x, y float32) {
	r.DrawRect(x, y, invSlotSize, invSlotSize, 0.15, 0.15, 0.15, 0.8)

	if s.Furnace == nil {
		r.DrawText(x+8, y+invSlotSize/2-6, "=>", 0.7, 0.5, 0.5, 0.5)
		return
	}

	fill := float32(0)
	if s.Furnace.Progress > 0 {
		recipe, found := inventory.FindSmeltingRecipe(s.Furnace.InputSlot.ItemID)
		if found && recipe.Duration > 0 {
			fill = float32(s.Furnace.Progress / recipe.Duration)
			if fill > 1 {
				fill = 1
			}
		}
	}

	if fill > 0 {
		fillWidth := invSlotSize * fill
		r.DrawRect(x, y, fillWidth, invSlotSize, 0.2, 0.7, 0.2, 0.9)
	}

	r.DrawText(x+8, y+invSlotSize/2-6, "=>", 0.7, 1, 1, 1)
}

// IsOverlay returns true.
func (s *FurnaceScreen) IsOverlay() bool { return true }

// HandleKey processes a raw key code.
func (s *FurnaceScreen) HandleKey(key int) {
	if key == input.KeyE || key == input.KeyEscape {
		s.Close()
	}
}

// IsClosed reports whether the screen has been closed.
func (s *FurnaceScreen) IsClosed() bool { return s.closed }

// Close closes the furnace screen, returning any held item to the player
// inventory.
func (s *FurnaceScreen) Close() {
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
func (s *FurnaceScreen) SetScreenSize(w, h float32) {
	s.screenWidth = w
	s.screenHeight = h
}

// handleClick processes a click at screen coordinates.
func (s *FurnaceScreen) handleClick(mx, my float32) {
	if s.hitTestInput(mx, my) {
		s.swapWithFurnaceSlot(&s.Furnace.InputSlot)
		return
	}

	if s.hitTestFuel(mx, my) {
		s.swapWithFurnaceSlot(&s.Furnace.FuelSlot)
		return
	}

	if s.hitTestOutput(mx, my) {
		s.takeOutput()
		return
	}

	playerIdx := s.hitTestPlayer(mx, my)
	if playerIdx >= 0 {
		s.swapWithPlayerSlot(playerIdx)
	}
}

// swapWithFurnaceSlot swaps the held item with a furnace slot.
func (s *FurnaceScreen) swapWithFurnaceSlot(slot *item.ItemStack) {
	if s.Furnace == nil || slot == nil {
		return
	}

	if s.HeldItem.IsEmpty() {
		s.HeldItem = *slot
		*slot = item.ItemStack{}
		return
	}

	if slot.IsEmpty() {
		*slot = s.HeldItem
		s.HeldItem = item.ItemStack{}
		return
	}

	if s.HeldItem.CanStackWith(*slot) {
		merged := *slot
		remaining := merged.Merge(s.HeldItem)
		*slot = merged
		s.HeldItem = remaining
		return
	}

	*slot, s.HeldItem = s.HeldItem, *slot
}

// takeOutput picks up the furnace output if hands are empty.
func (s *FurnaceScreen) takeOutput() {
	if s.Furnace == nil || !s.HeldItem.IsEmpty() || s.Furnace.OutputSlot.IsEmpty() {
		return
	}
	s.HeldItem = s.Furnace.OutputSlot
	s.Furnace.OutputSlot = item.ItemStack{}
}

// swapWithPlayerSlot swaps the held item with a player inventory slot.
func (s *FurnaceScreen) swapWithPlayerSlot(slotIdx int) {
	s.HeldItem = swapHeldWithSlot(s.HeldItem, s.PlayerInv, slotIdx)
}

// --- Layout helpers ---

func (s *FurnaceScreen) furnaceLayout(r *UIRenderer) (panelWidth, furnaceAreaHeight, pHeight, totalHeight, baseX, baseY float32) {
	panelWidth = playerPanelWidth()
	furnaceAreaHeight = float32(3)*(invSlotSize+invSlotPadding) + 2*invBgPadding
	pHeight = playerPanelHeight()
	totalHeight = furnaceAreaHeight + containerPanelGap + pHeight
	baseX = (r.ScreenWidth - panelWidth) / 2
	baseY = (r.ScreenHeight - totalHeight) / 2
	return
}

func (s *FurnaceScreen) inputSlotPos(r *UIRenderer) (float32, float32) {
	panelWidth, _, _, _, baseX, baseY := s.furnaceLayout(r)
	centerX := baseX + panelWidth/2
	return centerX - invSlotSize/2, baseY + invBgPadding + 10
}

func (s *FurnaceScreen) fuelSlotPos(r *UIRenderer) (float32, float32) {
	panelWidth, _, _, _, baseX, baseY := s.furnaceLayout(r)
	centerX := baseX + panelWidth/2
	inputY := baseY + invBgPadding + 10
	flameX := centerX - invSlotSize - 10
	flameY := inputY + invSlotSize + invSlotPadding + 4
	return flameX, flameY + invSlotSize + invSlotPadding
}

func (s *FurnaceScreen) outputSlotPos(r *UIRenderer) (float32, float32) {
	panelWidth, _, _, _, baseX, baseY := s.furnaceLayout(r)
	centerX := baseX + panelWidth/2
	inputY := baseY + invBgPadding + 10
	arrowX := centerX + 10
	arrowY := inputY + invSlotSize + invSlotPadding + 4
	return arrowX + invSlotSize + 10, arrowY
}

func (s *FurnaceScreen) playerSlotPos(r *UIRenderer, row, col int) (float32, float32) {
	_, furnaceAreaH, _, _, baseX, baseY := s.furnaceLayout(r)
	playerBaseY := baseY + furnaceAreaH + containerPanelGap + invBgPadding
	sx := baseX + invBgPadding + float32(col)*(invSlotSize+invSlotPadding)
	sy := playerBaseY + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

func (s *FurnaceScreen) hitTestInput(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.inputSlotPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *FurnaceScreen) hitTestFuel(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.fuelSlotPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *FurnaceScreen) hitTestOutput(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.outputSlotPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *FurnaceScreen) hitTestPlayer(mx, my float32) int {
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
