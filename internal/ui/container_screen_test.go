package ui

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
)

// ---------- ChestScreen tests ----------

func TestChestScreen_New(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	closed := false
	screen := NewChestScreen(chest, player, func() { closed = true })

	assert.NotNil(t, screen)
	assert.False(t, screen.IsClosed())
	assert.True(t, screen.HeldItem.IsEmpty())
	assert.True(t, screen.IsOverlay())
	assert.False(t, closed)
}

func TestChestScreen_CloseOnEscape(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	closed := false
	screen := NewChestScreen(chest, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestChestScreen_CloseOnE(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	closed := false
	screen := NewChestScreen(chest, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestChestScreen_CloseReturnsHeldItem(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.Close()

	assert.True(t, screen.HeldItem.IsEmpty())
	slot := player.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 10, slot.Count)
}

func TestChestScreen_CloseIdempotent(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	callCount := 0
	screen := NewChestScreen(chest, player, func() { callCount++ })

	screen.Close()
	screen.Close()
	assert.Equal(t, 1, callCount)
}

func TestChestScreen_HandleKey(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	closed := false
	screen := NewChestScreen(chest, player, func() { closed = true })

	screen.HandleKey(input.KeyE)
	assert.True(t, closed)
}

func TestChestScreen_SwapWithChest_PickUp(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 32})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	screen.swapWithInventory(chest, 0)
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 32, screen.HeldItem.Count)
	assert.True(t, chest.GetSlot(0).IsEmpty())
}

func TestChestScreen_SwapWithChest_PlaceDown(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Dirt, Count: 16}
	screen.swapWithInventory(chest, 5)

	assert.True(t, screen.HeldItem.IsEmpty())
	slot := chest.GetSlot(5)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 16, slot.Count)
}

func TestChestScreen_SwapWithChest_Swap(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 10})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Dirt, Count: 5}
	screen.swapWithInventory(chest, 0)

	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 10, screen.HeldItem.Count)
	slot := chest.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
}

func TestChestScreen_SwapWithNilInventory(t *testing.T) {
	screen := NewChestScreen(nil, nil, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.swapWithInventory(nil, 0)
	assert.Equal(t, 1, screen.HeldItem.Count)
}

func TestChestScreen_Draw(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 64})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	r := NewUIRenderer(800, 600)
	screen.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestChestScreen_SetScreenSize(t *testing.T) {
	screen := NewChestScreen(inventory.NewInventory(27), inventory.NewInventory(36), nil)
	screen.SetScreenSize(1920, 1080)
	assert.Equal(t, float32(1920), screen.screenWidth)
	assert.Equal(t, float32(1080), screen.screenHeight)
}

func TestChestScreen_SlotCounts(t *testing.T) {
	// Verify that chestTotalSlots = 27 and player = 36.
	assert.Equal(t, 27, chestTotalSlots)
	assert.Equal(t, 36, invTotalSlots)
}

// ---------- FurnaceScreen tests ----------

func TestFurnaceScreen_New(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	closed := false
	screen := NewFurnaceScreen(furnace, player, func() { closed = true })

	assert.NotNil(t, screen)
	assert.False(t, screen.IsClosed())
	assert.True(t, screen.HeldItem.IsEmpty())
	assert.True(t, screen.IsOverlay())
	assert.False(t, closed)
}

func TestFurnaceScreen_CloseOnEscape(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	closed := false
	screen := NewFurnaceScreen(furnace, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestFurnaceScreen_CloseOnE(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	closed := false
	screen := NewFurnaceScreen(furnace, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestFurnaceScreen_CloseReturnsHeldItem(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Coal, Count: 5}
	screen.Close()

	assert.True(t, screen.HeldItem.IsEmpty())
	slot := player.GetSlot(0)
	assert.Equal(t, item.Coal, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
}

func TestFurnaceScreen_CloseIdempotent(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	callCount := 0
	screen := NewFurnaceScreen(furnace, player, func() { callCount++ })

	screen.Close()
	screen.Close()
	assert.Equal(t, 1, callCount)
}

func TestFurnaceScreen_HandleKey(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	closed := false
	screen := NewFurnaceScreen(furnace, player, func() { closed = true })

	screen.HandleKey(input.KeyEscape)
	assert.True(t, closed)
}

func TestFurnaceScreen_SwapFurnaceInput(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.IronOre, Count: 8}
	screen.swapWithFurnaceSlot(&furnace.InputSlot)

	assert.True(t, screen.HeldItem.IsEmpty())
	assert.Equal(t, item.IronOre, furnace.InputSlot.ItemID)
	assert.Equal(t, 8, furnace.InputSlot.Count)
}

func TestFurnaceScreen_SwapFurnaceFuel(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.FuelSlot = item.ItemStack{ItemID: item.Coal, Count: 10}
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.swapWithFurnaceSlot(&furnace.FuelSlot)
	assert.Equal(t, item.Coal, screen.HeldItem.ItemID)
	assert.Equal(t, 10, screen.HeldItem.Count)
	assert.True(t, furnace.FuelSlot.IsEmpty())
}

func TestFurnaceScreen_TakeOutput_EmptyHands(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.OutputSlot = item.ItemStack{ItemID: item.IronIngot, Count: 3}
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.takeOutput()
	assert.Equal(t, item.IronIngot, screen.HeldItem.ItemID)
	assert.Equal(t, 3, screen.HeldItem.Count)
	assert.True(t, furnace.OutputSlot.IsEmpty())
}

func TestFurnaceScreen_TakeOutput_HandsFull(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.OutputSlot = item.ItemStack{ItemID: item.IronIngot, Count: 3}
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.takeOutput()
	// Should not take output when holding something.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 3, furnace.OutputSlot.Count)
}

func TestFurnaceScreen_TakeOutput_NilFurnace(t *testing.T) {
	screen := NewFurnaceScreen(nil, inventory.NewInventory(36), nil)
	screen.takeOutput()
	assert.True(t, screen.HeldItem.IsEmpty())
}

func TestFurnaceScreen_SwapPlayerSlot(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 32})
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.swapWithPlayerSlot(0)
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.True(t, player.GetSlot(0).IsEmpty())
}

func TestFurnaceScreen_SwapPlayerSlot_NilInv(t *testing.T) {
	furnace := inventory.NewFurnace()
	screen := NewFurnaceScreen(furnace, nil, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.swapWithPlayerSlot(0)
	assert.Equal(t, 1, screen.HeldItem.Count)
}

func TestFurnaceScreen_Draw(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.InputSlot = item.ItemStack{ItemID: item.IronOre, Count: 8}
	furnace.FuelSlot = item.ItemStack{ItemID: item.Coal, Count: 4}
	furnace.BurnTime = 5.0
	furnace.Progress = 3.0
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	r := NewUIRenderer(800, 600)
	screen.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestFurnaceScreen_Draw_NilFurnace(t *testing.T) {
	screen := NewFurnaceScreen(nil, inventory.NewInventory(36), nil)
	r := NewUIRenderer(800, 600)
	// Should not panic even with nil furnace.
	screen.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestFurnaceScreen_SetScreenSize(t *testing.T) {
	screen := NewFurnaceScreen(inventory.NewFurnace(), inventory.NewInventory(36), nil)
	screen.SetScreenSize(1920, 1080)
	assert.Equal(t, float32(1920), screen.screenWidth)
	assert.Equal(t, float32(1080), screen.screenHeight)
}

func TestFurnaceScreen_ShiftClickOutputMovesToPlayer(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.OutputSlot = item.ItemStack{ItemID: item.IronIngot, Count: 5}
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)
	screen.SetScreenSize(800, 600)

	// Locate the output slot center using hit-test layout.
	r := makeHitRenderer(800, 600)
	ox, oy := screen.outputSlotPos(r)
	cx, cy := ox+invSlotSize/2, oy+invSlotSize/2

	screen.handleShiftClick(cx, cy)

	// Output should be empty and player should have the ingots.
	assert.True(t, furnace.OutputSlot.IsEmpty(), "output slot should be cleared")
	slot := player.GetSlot(0)
	assert.Equal(t, item.IronIngot, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
}

func TestFurnaceScreen_ShiftClickPlayerFuelMovesToFuelSlot(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Coal, Count: 10})
	screen := NewFurnaceScreen(furnace, player, nil)
	screen.SetScreenSize(800, 600)

	// Locate player slot 0 center.
	r := makeHitRenderer(800, 600)
	px, py := screen.playerSlotPos(r, 0, 0)
	cx, cy := px+invSlotSize/2, py+invSlotSize/2

	screen.handleShiftClick(cx, cy)

	// Coal should move to the fuel slot.
	assert.Equal(t, item.Coal, furnace.FuelSlot.ItemID)
	assert.Equal(t, 10, furnace.FuelSlot.Count)
	assert.True(t, player.GetSlot(0).IsEmpty(), "player slot should be cleared")
}

func TestFurnaceScreen_ShiftClickPlayerSmeltableMovesToInputSlot(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.IronOre, Count: 8})
	screen := NewFurnaceScreen(furnace, player, nil)
	screen.SetScreenSize(800, 600)

	// Locate player slot 0 center.
	r := makeHitRenderer(800, 600)
	px, py := screen.playerSlotPos(r, 0, 0)
	cx, cy := px+invSlotSize/2, py+invSlotSize/2

	screen.handleShiftClick(cx, cy)

	// Iron ore should move to the input slot.
	assert.Equal(t, item.IronOre, furnace.InputSlot.ItemID)
	assert.Equal(t, 8, furnace.InputSlot.Count)
	assert.True(t, player.GetSlot(0).IsEmpty(), "player slot should be cleared")
}

// ---------- CraftingTableScreen tests ----------

func TestCraftingTableScreen_New(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	closed := false
	screen := NewCraftingTableScreen(grid, player, func() { closed = true })

	assert.NotNil(t, screen)
	assert.False(t, screen.IsClosed())
	assert.True(t, screen.HeldItem.IsEmpty())
	assert.True(t, screen.IsOverlay())
	assert.False(t, closed)
}

func TestCraftingTableScreen_CloseOnEscape(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	closed := false
	screen := NewCraftingTableScreen(grid, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestCraftingTableScreen_CloseOnE(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	closed := false
	screen := NewCraftingTableScreen(grid, player, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestCraftingTableScreen_CloseReturnsHeldItem(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.Close()

	assert.True(t, screen.HeldItem.IsEmpty())
	slot := player.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 10, slot.Count)
}

func TestCraftingTableScreen_CloseReturnsCraftGridItems(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	grid.SetSlot(0, 0, item.ItemStack{ItemID: item.OakPlanks, Count: 4})
	grid.SetSlot(1, 1, item.ItemStack{ItemID: item.Stick, Count: 2})
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.Close()

	// Grid items should be returned to the player inventory.
	assert.True(t, grid.GetSlot(0, 0).IsEmpty())
	assert.True(t, grid.GetSlot(1, 1).IsEmpty())

	// Player should have received both stacks.
	foundPlanks := false
	foundSticks := false
	for i := 0; i < player.Size(); i++ {
		s := player.GetSlot(i)
		if s.ItemID == item.OakPlanks && s.Count == 4 {
			foundPlanks = true
		}
		if s.ItemID == item.Stick && s.Count == 2 {
			foundSticks = true
		}
	}
	assert.True(t, foundPlanks, "Planks should be returned to player inventory")
	assert.True(t, foundSticks, "Sticks should be returned to player inventory")
}

func TestCraftingTableScreen_CloseIdempotent(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	callCount := 0
	screen := NewCraftingTableScreen(grid, player, func() { callCount++ })

	screen.Close()
	screen.Close()
	assert.Equal(t, 1, callCount)
}

func TestCraftingTableScreen_HandleKey(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	closed := false
	screen := NewCraftingTableScreen(grid, player, func() { closed = true })

	screen.HandleKey(input.KeyE)
	assert.True(t, closed)
}

func TestCraftingTableScreen_SwapWithCraftSlot_PlaceDown(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.OakPlanks, Count: 4}
	screen.swapWithCraftSlot(0, 0)

	assert.True(t, screen.HeldItem.IsEmpty())
	placed := grid.GetSlot(0, 0)
	assert.Equal(t, item.OakPlanks, placed.ItemID)
	assert.Equal(t, 4, placed.Count)
}

func TestCraftingTableScreen_SwapWithCraftSlot_PickUp(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	grid.SetSlot(1, 1, item.ItemStack{ItemID: item.Stick, Count: 2})
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.swapWithCraftSlot(1, 1)
	assert.Equal(t, item.Stick, screen.HeldItem.ItemID)
	assert.True(t, grid.GetSlot(1, 1).IsEmpty())
}

func TestCraftingTableScreen_SwapWithCraftSlot_NilGrid(t *testing.T) {
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(nil, player, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.swapWithCraftSlot(0, 0)
	assert.Equal(t, 1, screen.HeldItem.Count)
}

func TestCraftingTableScreen_TakeCraftResult_EmptyHands(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	// No recipe set, so Craft() returns false.
	screen.takeCraftResult()
	assert.True(t, screen.HeldItem.IsEmpty())
}

func TestCraftingTableScreen_TakeCraftResult_HandsFull(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.takeCraftResult()
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
}

func TestCraftingTableScreen_TakeCraftResult_NilGrid(t *testing.T) {
	screen := NewCraftingTableScreen(nil, inventory.NewInventory(36), nil)
	screen.takeCraftResult()
	assert.True(t, screen.HeldItem.IsEmpty())
}

func TestCraftingTableScreen_SwapWithPlayerSlot(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 32})
	screen := NewCraftingTableScreen(grid, player, nil)

	screen.swapWithPlayerSlot(0)
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.True(t, player.GetSlot(0).IsEmpty())
}

func TestCraftingTableScreen_SwapWithPlayerSlot_NilInv(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	screen := NewCraftingTableScreen(grid, nil, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.swapWithPlayerSlot(0)
	assert.Equal(t, 1, screen.HeldItem.Count)
}

func TestCraftingTableScreen_Draw(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 64})
	screen := NewCraftingTableScreen(grid, player, nil)

	r := NewUIRenderer(800, 600)
	screen.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestCraftingTableScreen_SetScreenSize(t *testing.T) {
	screen := NewCraftingTableScreen(&inventory.CraftingGrid{}, inventory.NewInventory(36), nil)
	screen.SetScreenSize(1920, 1080)
	assert.Equal(t, float32(1920), screen.screenWidth)
	assert.Equal(t, float32(1080), screen.screenHeight)
}

func TestCraftingTableScreen_GridSize(t *testing.T) {
	assert.Equal(t, 3, ctGridSize, "Crafting table should use 3x3 grid")
}

// ---------- Integration: container screens in UIManager ----------

func TestUIManager_IsBlockingInput_ChestScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	chest := NewChestScreen(inventory.NewInventory(27), inventory.NewInventory(36), nil)
	m.PushScreen(hud)
	m.PushScreen(chest)
	assert.True(t, m.IsBlockingInput(), "Chest screen should block input")
}

func TestUIManager_IsBlockingInput_FurnaceScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	furnace := NewFurnaceScreen(inventory.NewFurnace(), inventory.NewInventory(36), nil)
	m.PushScreen(hud)
	m.PushScreen(furnace)
	assert.True(t, m.IsBlockingInput(), "Furnace screen should block input")
}

func TestUIManager_IsBlockingInput_CraftingTableScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	ct := NewCraftingTableScreen(&inventory.CraftingGrid{}, inventory.NewInventory(36), nil)
	m.PushScreen(hud)
	m.PushScreen(ct)
	assert.True(t, m.IsBlockingInput(), "Crafting table screen should block input")
}

func TestIntegration_ChestOpenClose(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	m.PushScreen(hud)

	chest := NewChestScreen(
		inventory.NewInventory(27),
		inventory.NewInventory(36),
		func() { m.PopScreen() },
	)
	m.PushScreen(chest)
	assert.Equal(t, 2, m.ScreenCount())
	assert.True(t, m.IsBlockingInput())

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)

	assert.Equal(t, 1, m.ScreenCount())
	assert.False(t, m.IsBlockingInput())
}

func TestIntegration_FurnaceOpenClose(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	m.PushScreen(hud)

	furnace := NewFurnaceScreen(
		inventory.NewFurnace(),
		inventory.NewInventory(36),
		func() { m.PopScreen() },
	)
	m.PushScreen(furnace)
	assert.Equal(t, 2, m.ScreenCount())

	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)

	assert.Equal(t, 1, m.ScreenCount())
}

func TestIntegration_CraftingTableOpenClose(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	m.PushScreen(hud)

	ct := NewCraftingTableScreen(
		&inventory.CraftingGrid{},
		inventory.NewInventory(36),
		func() { m.PopScreen() },
	)
	m.PushScreen(ct)
	assert.Equal(t, 2, m.ScreenCount())

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)

	assert.Equal(t, 1, m.ScreenCount())
}
