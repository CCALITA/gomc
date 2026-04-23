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

// ---------- ChestScreen right-click tests ----------

func TestChestScreen_RightClick_PickUpHalf(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 10})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	// Right-click with empty hand on a slot with 10 items: pick up half (rounded up = 5).
	screen.HeldItem = rightClickSlot(screen.HeldItem, chest, 0)

	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 5, screen.HeldItem.Count)
	remaining := chest.GetSlot(0)
	assert.Equal(t, item.Stone, remaining.ItemID)
	assert.Equal(t, 5, remaining.Count)
}

func TestChestScreen_RightClick_PlaceOne(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	// Holding 10 stone, right-click on empty slot: place exactly 1.
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.HeldItem = rightClickSlot(screen.HeldItem, chest, 0)

	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 9, screen.HeldItem.Count)
	placed := chest.GetSlot(0)
	assert.Equal(t, item.Stone, placed.ItemID)
	assert.Equal(t, 1, placed.Count)
}

func TestChestScreen_RightClick_Incompatible(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 5})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	// Holding stone, right-click on slot with dirt: no-op.
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.HeldItem = rightClickSlot(screen.HeldItem, chest, 0)

	// Held item unchanged.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 10, screen.HeldItem.Count)
	// Slot unchanged.
	slot := chest.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
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

// ---------- ChestScreen shift-click (quick-move) tests ----------

func TestChestScreen_ShiftClick_ChestToPlayer(t *testing.T) {
	chest := inventory.NewInventory(27)
	chest.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 32})
	player := inventory.NewInventory(36)
	screen := NewChestScreen(chest, player, nil)

	screen.quickMove(chest, 0, player)

	assert.True(t, chest.GetSlot(0).IsEmpty(), "chest slot should be empty after quick move")
	slot := player.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 32, slot.Count)
}

func TestChestScreen_ShiftClick_PlayerToChest(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 16})
	screen := NewChestScreen(chest, player, nil)

	screen.quickMove(player, 0, chest)

	assert.True(t, player.GetSlot(0).IsEmpty(), "player slot should be empty after quick move")
	slot := chest.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 16, slot.Count)
}

func TestChestScreen_ShiftClick_PartialTransfer(t *testing.T) {
	chest := inventory.NewInventory(27)
	player := inventory.NewInventory(36)

	// Fill every player slot to capacity with stone so no room remains.
	for i := 0; i < player.Size(); i++ {
		player.SetSlot(i, item.ItemStack{ItemID: item.Stone, Count: item.MaxStack(item.Stone)})
	}

	// Place dirt in chest slot 0 — it cannot merge with stone, and all slots are full.
	chest.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 10})
	screen := NewChestScreen(chest, player, nil)

	screen.quickMove(chest, 0, player)

	// The entire stack should remain in the chest because the player inventory is full.
	remaining := chest.GetSlot(0)
	assert.Equal(t, item.Dirt, remaining.ItemID)
	assert.Equal(t, 10, remaining.Count, "all items should remain when target is full")
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

func TestFurnaceScreen_RightClickPickUpHalf_FurnaceSlot(t *testing.T) {
	furnace := inventory.NewFurnace()
	furnace.InputSlot = item.ItemStack{ItemID: item.IronOre, Count: 10}
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	// Right-click with empty hand picks up half (floor = 5 taken, ceiling = 5 stays).
	screen.rightClickSlot(&furnace.InputSlot)

	assert.Equal(t, item.IronOre, screen.HeldItem.ItemID)
	assert.Equal(t, 5, screen.HeldItem.Count)
	assert.Equal(t, item.IronOre, furnace.InputSlot.ItemID)
	assert.Equal(t, 5, furnace.InputSlot.Count)
}

func TestFurnaceScreen_RightClickPlaceOne_FurnaceSlot(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	screen := NewFurnaceScreen(furnace, player, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Coal, Count: 8}

	// Right-click with held items on empty slot places one.
	screen.rightClickSlot(&furnace.FuelSlot)

	assert.Equal(t, item.Coal, furnace.FuelSlot.ItemID)
	assert.Equal(t, 1, furnace.FuelSlot.Count)
	assert.Equal(t, item.Coal, screen.HeldItem.ItemID)
	assert.Equal(t, 7, screen.HeldItem.Count)
}

func TestFurnaceScreen_RightClickPlayerSlot_PickUpHalf(t *testing.T) {
	furnace := inventory.NewFurnace()
	player := inventory.NewInventory(36)
	player.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 16})
	screen := NewFurnaceScreen(furnace, player, nil)

	// Right-click with empty hand picks up half from player slot.
	screen.rightClickPlayerSlot(0)

	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 8, screen.HeldItem.Count)
	slot := player.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 8, slot.Count)
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

// ---------- InventoryScreen right-click tests ----------

func TestInventoryScreen_RightClick_PickUpHalf(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 10})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	screen.rightClickSlot(0)

	// Should pick up ceil(10/2) = 5 items.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 5, screen.HeldItem.Count)
	// Slot should have the remaining 5.
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
}

func TestInventoryScreen_RightClick_PickUpHalfOdd(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 7})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	screen.rightClickSlot(0)

	// Should pick up ceil(7/2) = 4 items.
	assert.Equal(t, item.Dirt, screen.HeldItem.ItemID)
	assert.Equal(t, 4, screen.HeldItem.Count)
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 3, slot.Count)
}

func TestInventoryScreen_RightClick_PlaceOneInEmptySlot(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.rightClickSlot(5)

	// Slot 5 should have 1 item.
	slot := inv.GetSlot(5)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 1, slot.Count)
	// Held should have 9.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 9, screen.HeldItem.Count)
}

func TestInventoryScreen_RightClick_IncompatibleDoesNothing(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 5})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)
	screen.SetScreenSize(800, 600)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 3}
	screen.rightClickSlot(0)

	// Nothing should change.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 3, screen.HeldItem.Count)
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
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

// ---------- AnvilScreen tests ----------

func TestAnvilScreen_New(t *testing.T) {
	player := inventory.NewInventory(36)
	closed := false
	screen := NewAnvilScreen(player, func() { closed = true })

	assert.NotNil(t, screen)
	assert.False(t, screen.IsClosed())
	assert.True(t, screen.HeldItem.IsEmpty())
	assert.True(t, screen.LeftInput.IsEmpty())
	assert.True(t, screen.RightInput.IsEmpty())
	assert.True(t, screen.Output.IsEmpty())
	assert.True(t, screen.IsOverlay())
	assert.False(t, closed)
}

func TestAnvilScreen_CloseReturnsInputItems(t *testing.T) {
	player := inventory.NewInventory(36)
	screen := NewAnvilScreen(player, nil)

	screen.LeftInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 100}
	screen.RightInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 50}
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.Close()

	assert.True(t, screen.LeftInput.IsEmpty())
	assert.True(t, screen.RightInput.IsEmpty())
	assert.True(t, screen.HeldItem.IsEmpty())

	// All three stacks should be returned to the player inventory.
	foundPickaxe1 := false
	foundPickaxe2 := false
	foundStone := false
	for i := 0; i < player.Size(); i++ {
		s := player.GetSlot(i)
		if s.ItemID == item.IronPickaxe && s.Durability == 100 {
			foundPickaxe1 = true
		}
		if s.ItemID == item.IronPickaxe && s.Durability == 50 {
			foundPickaxe2 = true
		}
		if s.ItemID == item.Stone && s.Count == 10 {
			foundStone = true
		}
	}
	assert.True(t, foundPickaxe1, "Left input should be returned to player inventory")
	assert.True(t, foundPickaxe2, "Right input should be returned to player inventory")
	assert.True(t, foundStone, "Held item should be returned to player inventory")
}

func TestAnvilScreen_RepairComputesOutput(t *testing.T) {
	player := inventory.NewInventory(36)
	screen := NewAnvilScreen(player, nil)

	// Place two iron pickaxes with partial durability into the input slots.
	screen.LeftInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 100}
	screen.RightInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 80}
	screen.recomputeOutput()

	assert.False(t, screen.Output.IsEmpty(), "Output should be computed for valid repair")
	assert.Equal(t, item.IronPickaxe, screen.Output.ItemID)
	assert.Greater(t, screen.Output.Durability, 0)
	assert.Greater(t, screen.XPCost, 0)
}

func TestAnvilScreen_TakeOutputConsumesInputs(t *testing.T) {
	player := inventory.NewInventory(36)
	screen := NewAnvilScreen(player, nil)

	screen.LeftInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 100}
	screen.RightInput = item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 80}
	screen.recomputeOutput()

	savedOutput := screen.Output
	screen.takeOutput()

	assert.Equal(t, savedOutput.ItemID, screen.HeldItem.ItemID)
	assert.True(t, screen.LeftInput.IsEmpty(), "Left input should be consumed")
	assert.True(t, screen.RightInput.IsEmpty(), "Right input should be consumed")
	assert.True(t, screen.Output.IsEmpty(), "Output slot should be cleared")
	assert.Equal(t, 0, screen.XPCost)
}

func TestAnvilScreen_RenameOnlyOutput(t *testing.T) {
	player := inventory.NewInventory(36)
	screen := NewAnvilScreen(player, nil)

	screen.LeftInput = item.ItemStack{ItemID: item.DiamondSword, Count: 1, Durability: 1561}
	screen.RenameTo = "Excalibur"
	screen.recomputeOutput()

	assert.False(t, screen.Output.IsEmpty(), "Rename-only should produce output")
	assert.Equal(t, "Excalibur", screen.Output.CustomName)
	assert.Equal(t, item.DiamondSword, screen.Output.ItemID)
	assert.Equal(t, 1, screen.XPCost, "Rename-only costs 1 XP level")
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

// ---------- InventoryScreen Armor Slot tests ----------

func TestInventoryScreen_ArmorSlot_EquipCorrectSlot(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	// Hold an iron helmet (ArmorSlot=0) and click armor slot 0.
	screen.HeldItem = item.NewItemStack(item.IronHelmet, 1)
	screen.handleArmorSlotClick(0)

	assert.True(t, screen.HeldItem.IsEmpty(), "Held item should be empty after equipping")
	assert.Equal(t, item.IronHelmet, screen.ArmorSlots[0].ItemID)
	assert.Equal(t, 1, screen.ArmorSlots[0].Count)
}

func TestInventoryScreen_ArmorSlot_UnequipEmptyHands(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	// Place boots in armor slot 3, then click with empty hands.
	screen.ArmorSlots[3] = item.NewItemStack(item.IronBoots, 1)
	screen.handleArmorSlotClick(3)

	assert.Equal(t, item.IronBoots, screen.HeldItem.ItemID, "Should pick up boots")
	assert.True(t, screen.ArmorSlots[3].IsEmpty(), "Armor slot should be empty after unequip")
}

func TestInventoryScreen_ArmorSlot_RejectWrongSlot(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	// Hold a chestplate (ArmorSlot=1) and try to place in slot 0 (helmet).
	screen.HeldItem = item.NewItemStack(item.IronChestplate, 1)
	screen.handleArmorSlotClick(0)

	assert.Equal(t, item.IronChestplate, screen.HeldItem.ItemID, "Should still be holding chestplate")
	assert.True(t, screen.ArmorSlots[0].IsEmpty(), "Helmet slot should remain empty")
}
