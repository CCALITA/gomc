package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

const (
	anvilInputSlots = 2 // left material, right material
)

// AnvilScreen displays an anvil GUI with two input slots (left and right),
// one output slot, a rename field, and the player inventory at the bottom.
type AnvilScreen struct {
	// LeftInput is the primary item being modified.
	LeftInput item.ItemStack

	// RightInput is the sacrifice/material item.
	RightInput item.ItemStack

	// Output is the computed result (read-only, recomputed on input change).
	Output item.ItemStack

	// XPCost is the experience level cost of the current operation.
	XPCost int

	// RenameTo is the desired custom name for the output item.
	RenameTo string

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

// NewAnvilScreen creates an anvil screen bound to the given player inventory.
func NewAnvilScreen(playerInv *inventory.Inventory, onClose func()) *AnvilScreen {
	return &AnvilScreen{
		PlayerInv: playerInv,
		onClose:   onClose,
	}
}

// Update handles input for the anvil screen.
func (s *AnvilScreen) Update(inp *input.Manager, _ float64) {
	s.mouseX, s.mouseY = inp.MousePos()

	if inp.IsKeyJustPressed(input.KeyE) || inp.IsKeyJustPressed(input.KeyEscape) {
		s.Close()
		return
	}

	if inp.IsMouseJustPressed(input.MouseButtonLeft) {
		s.handleClick(float32(s.mouseX), float32(s.mouseY))
	}
}

// Draw renders the anvil screen.
func (s *AnvilScreen) Draw(r *UIRenderer) {
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.6)

	panelWidth, anvilAreaHeight, _, baseX, baseY := s.anvilLayout(r)

	// Anvil panel background.
	r.DrawRect(baseX, baseY, panelWidth, anvilAreaHeight, 0.4, 0.4, 0.4, 0.9)
	r.DrawText(baseX+invBgPadding, baseY+4, "Anvil", 0.8, 0.9, 0.9, 0.9)

	// Left input slot.
	leftX, leftY := s.leftInputPos(r)
	r.DrawRect(leftX, leftY, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
	if !s.LeftInput.IsEmpty() {
		r.DrawItemSlot(leftX+2, leftY+2, s.LeftInput)
	}

	// "+" label between inputs.
	plusX := leftX + invSlotSize + invSlotPadding + 8
	plusY := leftY + invSlotSize/2 - 6
	r.DrawText(plusX, plusY, "+", 1.0, 1, 1, 1)

	// Right input slot.
	rightX, rightY := s.rightInputPos(r)
	r.DrawRect(rightX, rightY, invSlotSize, invSlotSize, 0.2, 0.2, 0.2, 0.8)
	if !s.RightInput.IsEmpty() {
		r.DrawItemSlot(rightX+2, rightY+2, s.RightInput)
	}

	// Arrow between inputs and output.
	arrowX := rightX + invSlotSize + invSlotPadding + 8
	arrowY := rightY + invSlotSize/2 - 6
	r.DrawText(arrowX, arrowY, "=>", 1.0, 1, 1, 1)

	// Output slot.
	outputX, outputY := s.outputPos(r)
	r.DrawRect(outputX, outputY, invSlotSize, invSlotSize, 0.3, 0.3, 0.15, 0.8)
	if !s.Output.IsEmpty() {
		r.DrawItemSlot(outputX+2, outputY+2, s.Output)
	}

	// Player inventory panel.
	pHeight := playerPanelHeight()
	playerBaseY := baseY + anvilAreaHeight + containerPanelGap
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
func (s *AnvilScreen) IsOverlay() bool { return true }

// HandleKey processes a raw key code.
func (s *AnvilScreen) HandleKey(key int) {
	if key == input.KeyE || key == input.KeyEscape {
		s.Close()
	}
}

// IsClosed reports whether the screen has been closed.
func (s *AnvilScreen) IsClosed() bool { return s.closed }

// Close closes the anvil screen, returning input items and any held item
// to the player inventory.
func (s *AnvilScreen) Close() {
	if s.closed {
		return
	}
	s.closed = true

	// Return held item.
	if !s.HeldItem.IsEmpty() && s.PlayerInv != nil {
		s.PlayerInv.AddItem(s.HeldItem)
		s.HeldItem = item.ItemStack{}
	}

	// Return left input.
	if !s.LeftInput.IsEmpty() && s.PlayerInv != nil {
		s.PlayerInv.AddItem(s.LeftInput)
		s.LeftInput = item.ItemStack{}
	}

	// Return right input.
	if !s.RightInput.IsEmpty() && s.PlayerInv != nil {
		s.PlayerInv.AddItem(s.RightInput)
		s.RightInput = item.ItemStack{}
	}

	if s.onClose != nil {
		s.onClose()
	}
}

// SetScreenSize sets the screen dimensions for hit testing.
func (s *AnvilScreen) SetScreenSize(w, h float32) {
	s.screenWidth = w
	s.screenHeight = h
}

// recomputeOutput recalculates the output using CalculateAnvilOutput.
func (s *AnvilScreen) recomputeOutput() {
	op := inventory.AnvilOperation{
		Input:      s.LeftInput,
		Material:   s.RightInput,
		OutputName: s.RenameTo,
	}
	s.Output, s.XPCost = inventory.CalculateAnvilOutput(op)
}

// handleClick processes a click at screen coordinates.
func (s *AnvilScreen) handleClick(mx, my float32) {
	if s.hitTestLeftInput(mx, my) {
		s.swapWithInputSlot(&s.LeftInput)
		return
	}

	if s.hitTestRightInput(mx, my) {
		s.swapWithInputSlot(&s.RightInput)
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

// swapWithInputSlot swaps the held item with an anvil input slot and
// recomputes the output.
func (s *AnvilScreen) swapWithInputSlot(slot *item.ItemStack) {
	if slot == nil {
		return
	}

	if s.HeldItem.IsEmpty() {
		s.HeldItem = *slot
		*slot = item.ItemStack{}
		s.recomputeOutput()
		return
	}

	if slot.IsEmpty() {
		*slot = s.HeldItem
		s.HeldItem = item.ItemStack{}
		s.recomputeOutput()
		return
	}

	// Swap.
	*slot, s.HeldItem = s.HeldItem, *slot
	s.recomputeOutput()
}

// takeOutput picks up the anvil output if hands are empty and an output
// exists. Consumes the input items.
func (s *AnvilScreen) takeOutput() {
	if s.HeldItem.IsEmpty() && !s.Output.IsEmpty() {
		s.HeldItem = s.Output
		s.LeftInput = item.ItemStack{}
		s.RightInput = item.ItemStack{}
		s.Output = item.ItemStack{}
		s.XPCost = 0
	}
}

// swapWithPlayerSlot swaps the held item with a player inventory slot.
func (s *AnvilScreen) swapWithPlayerSlot(slotIdx int) {
	s.HeldItem = swapHeldWithSlot(s.HeldItem, s.PlayerInv, slotIdx)
}

// --- Layout helpers ---

func (s *AnvilScreen) anvilLayout(r *UIRenderer) (panelWidth, anvilAreaHeight, totalHeight, baseX, baseY float32) {
	panelWidth = playerPanelWidth()
	anvilAreaHeight = invSlotSize + 2*invBgPadding + 20
	pHeight := playerPanelHeight()
	totalHeight = anvilAreaHeight + containerPanelGap + pHeight
	baseX = (r.ScreenWidth - panelWidth) / 2
	baseY = (r.ScreenHeight - totalHeight) / 2
	return
}

func (s *AnvilScreen) leftInputPos(r *UIRenderer) (float32, float32) {
	_, _, _, baseX, baseY := s.anvilLayout(r)
	x := baseX + invBgPadding
	y := baseY + invBgPadding + 20
	return x, y
}

func (s *AnvilScreen) rightInputPos(r *UIRenderer) (float32, float32) {
	leftX, leftY := s.leftInputPos(r)
	x := leftX + invSlotSize + invSlotPadding + 32
	return x, leftY
}

func (s *AnvilScreen) outputPos(r *UIRenderer) (float32, float32) {
	rightX, rightY := s.rightInputPos(r)
	x := rightX + invSlotSize + invSlotPadding + 40
	return x, rightY
}

func (s *AnvilScreen) playerSlotPos(r *UIRenderer, row, col int) (float32, float32) {
	_, anvilAreaH, _, baseX, baseY := s.anvilLayout(r)
	playerBaseY := baseY + anvilAreaH + containerPanelGap + invBgPadding
	sx := baseX + invBgPadding + float32(col)*(invSlotSize+invSlotPadding)
	sy := playerBaseY + float32(row)*(invSlotSize+invSlotPadding)
	return sx, sy
}

// --- Hit testing ---

func (s *AnvilScreen) hitTestLeftInput(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.leftInputPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *AnvilScreen) hitTestRightInput(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.rightInputPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *AnvilScreen) hitTestOutput(mx, my float32) bool {
	r := makeHitRenderer(s.screenWidth, s.screenHeight)
	sx, sy := s.outputPos(r)
	return mx >= sx && mx <= sx+invSlotSize && my >= sy && my <= sy+invSlotSize
}

func (s *AnvilScreen) hitTestPlayer(mx, my float32) int {
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
