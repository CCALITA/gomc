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

// ---------- InventoryScreen shift-click tests ----------

func TestInventoryScreen_ShiftClickCraftResult(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	// Set up a 2x2 planks recipe for a crafting table.
	grid.SetSlot(0, 0, item.ItemStack{ItemID: item.OakPlanks, Count: 1})
	grid.SetSlot(0, 1, item.ItemStack{ItemID: item.OakPlanks, Count: 1})
	grid.SetSlot(1, 0, item.ItemStack{ItemID: item.OakPlanks, Count: 1})
	grid.SetSlot(1, 1, item.ItemStack{ItemID: item.OakPlanks, Count: 1})

	inv := inventory.NewInventory(36)
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	// Determine the craft result position via hit test.
	r := NewUIRenderer(800, 600)
	baseX, baseY := screen.craftGridOrigin(r)
	arrowX := baseX + float32(craftGridSize)*(invSlotSize+invSlotPadding) + 8
	resultX := arrowX + 32
	resultY := baseY + (invSlotSize+invSlotPadding)/2 - invSlotSize/2 + invSlotSize/4

	// Verify a recipe result exists before shift-clicking.
	result := grid.GetResult()
	if result.IsEmpty() {
		t.Skip("No recipe registered for 2x2 planks; skipping shift-click craft result test")
	}

	// Simulate shift+left-click on craft result.
	screen.handleShiftClick(resultX+1, resultY+1)

	// The crafted item should go directly into the inventory, not the cursor.
	assert.True(t, screen.HeldItem.IsEmpty(), "Held item should remain empty after shift-click craft")

	// The item should be in the player inventory.
	found := false
	for i := 0; i < inv.Size(); i++ {
		if !inv.GetSlot(i).IsEmpty() {
			found = true
			break
		}
	}
	assert.True(t, found, "Crafted item should be added to player inventory")
}

func TestInventoryScreen_ShiftClickCraftGrid(t *testing.T) {
	grid := &inventory.CraftingGrid{}
	grid.SetSlot(0, 0, item.ItemStack{ItemID: item.Stone, Count: 5})

	inv := inventory.NewInventory(36)
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	// Determine the craft grid slot (0,0) position.
	r := NewUIRenderer(800, 600)
	baseX, baseY := screen.craftGridOrigin(r)
	slotX := baseX + 1
	slotY := baseY + 1

	screen.handleShiftClick(slotX, slotY)

	// The craft grid slot should now be empty.
	assert.True(t, grid.GetSlot(0, 0).IsEmpty(), "Craft grid slot should be emptied")

	// The item should be in the player inventory.
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 5, slot.Count)

	// Cursor should remain empty.
	assert.True(t, screen.HeldItem.IsEmpty(), "Held item should remain empty")
}

func TestInventoryScreen_ShiftClickInventorySlot_HotbarToMain(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 16}) // hotbar slot 0

	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	// Determine the position of hotbar slot 0 (row=0, col=0).
	r := NewUIRenderer(800, 600)
	sx, sy := screen.inventorySlotPos(r, 0, 0)

	screen.handleShiftClick(sx+1, sy+1)

	// Hotbar slot 0 should now be empty.
	assert.True(t, inv.GetSlot(0).IsEmpty(), "Hotbar slot should be emptied")

	// The item should have moved to main inventory (slots 9+).
	found := false
	for i := invCols; i < invTotalSlots; i++ {
		s := inv.GetSlot(i)
		if s.ItemID == item.Dirt && s.Count == 16 {
			found = true
			break
		}
	}
	assert.True(t, found, "Item should be moved to main inventory area")

	// Cursor should remain empty.
	assert.True(t, screen.HeldItem.IsEmpty(), "Held item should remain empty")
}
