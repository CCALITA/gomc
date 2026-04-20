package inventory

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Inventory tests
// ---------------------------------------------------------------------------

func TestNewInventory(t *testing.T) {
	inv := NewInventory(36)
	assert.Equal(t, 36, inv.Size())
	assert.True(t, inv.IsEmpty())
}

func TestNewInventory_ZeroAndNegative(t *testing.T) {
	inv := NewInventory(0)
	assert.Equal(t, 0, inv.Size())

	inv2 := NewInventory(-5)
	assert.Equal(t, 0, inv2.Size())
}

func TestSetSlot_GetSlot(t *testing.T) {
	inv := NewInventory(5)
	stack := item.NewItemStack(item.Dirt, 10)

	inv.SetSlot(2, stack)
	got := inv.GetSlot(2)
	assert.Equal(t, item.Dirt, got.ItemID)
	assert.Equal(t, 10, got.Count)
}

func TestGetSlot_OutOfBounds(t *testing.T) {
	inv := NewInventory(3)
	assert.True(t, inv.GetSlot(-1).IsEmpty())
	assert.True(t, inv.GetSlot(5).IsEmpty())
}

func TestSetSlot_OutOfBounds(t *testing.T) {
	inv := NewInventory(3)
	// Should not panic.
	inv.SetSlot(-1, item.NewItemStack(item.Dirt, 1))
	inv.SetSlot(10, item.NewItemStack(item.Dirt, 1))
	assert.True(t, inv.IsEmpty())
}

func TestAddItem_IntoEmptySlots(t *testing.T) {
	inv := NewInventory(3)
	remaining := inv.AddItem(item.NewItemStack(item.Dirt, 10))
	assert.True(t, remaining.IsEmpty())
	assert.Equal(t, 10, inv.GetSlot(0).Count)
}

func TestAddItem_MergeIntoExisting(t *testing.T) {
	inv := NewInventory(3)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 60))

	remaining := inv.AddItem(item.NewItemStack(item.Dirt, 10))
	assert.True(t, remaining.IsEmpty())
	// Should have merged: 60 + 4 = 64 (max), then 6 in next slot.
	assert.Equal(t, 64, inv.GetSlot(0).Count)
	assert.Equal(t, 6, inv.GetSlot(1).Count)
}

func TestAddItem_InventoryFull(t *testing.T) {
	inv := NewInventory(1)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 64))

	remaining := inv.AddItem(item.NewItemStack(item.Dirt, 10))
	assert.Equal(t, 10, remaining.Count)
}

func TestAddItem_EmptyStack(t *testing.T) {
	inv := NewInventory(3)
	remaining := inv.AddItem(item.ItemStack{})
	assert.True(t, remaining.IsEmpty())
}

func TestAddItem_LargeStack_SpansMultipleSlots(t *testing.T) {
	inv := NewInventory(3)
	remaining := inv.AddItem(item.NewItemStack(item.Cobblestone, 150))
	// 64 + 64 + 22 = 150, all fit in 3 slots.
	assert.True(t, remaining.IsEmpty())
	assert.Equal(t, 64, inv.GetSlot(0).Count)
	assert.Equal(t, 64, inv.GetSlot(1).Count)
	assert.Equal(t, 22, inv.GetSlot(2).Count)
}

func TestAddItem_Unstackable(t *testing.T) {
	inv := NewInventory(2)
	// Tools have max stack 1 and are not stackable.
	remaining := inv.AddItem(item.NewItemStack(item.WoodenPickaxe, 1))
	assert.True(t, remaining.IsEmpty())

	remaining = inv.AddItem(item.NewItemStack(item.WoodenPickaxe, 1))
	assert.True(t, remaining.IsEmpty())

	// Inventory full for tools now.
	remaining = inv.AddItem(item.NewItemStack(item.WoodenPickaxe, 1))
	assert.Equal(t, 1, remaining.Count)
}

func TestRemoveItem(t *testing.T) {
	inv := NewInventory(5)
	inv.SetSlot(1, item.NewItemStack(item.Stone, 32))

	taken := inv.RemoveItem(1, 10)
	assert.Equal(t, 10, taken.Count)
	assert.Equal(t, item.Stone, taken.ItemID)
	assert.Equal(t, 22, inv.GetSlot(1).Count)
}

func TestRemoveItem_EntireSlot(t *testing.T) {
	inv := NewInventory(5)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 5))

	taken := inv.RemoveItem(0, 10) // more than available
	assert.Equal(t, 5, taken.Count)
	assert.True(t, inv.GetSlot(0).IsEmpty())
}

func TestRemoveItem_InvalidIndex(t *testing.T) {
	inv := NewInventory(5)
	taken := inv.RemoveItem(-1, 5)
	assert.True(t, taken.IsEmpty())

	taken = inv.RemoveItem(100, 5)
	assert.True(t, taken.IsEmpty())
}

func TestRemoveItem_ZeroCount(t *testing.T) {
	inv := NewInventory(5)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 10))
	taken := inv.RemoveItem(0, 0)
	assert.True(t, taken.IsEmpty())
	assert.Equal(t, 10, inv.GetSlot(0).Count) // unchanged
}

func TestRemoveItem_EmptySlot(t *testing.T) {
	inv := NewInventory(5)
	taken := inv.RemoveItem(0, 5)
	assert.True(t, taken.IsEmpty())
}

func TestFindItem(t *testing.T) {
	inv := NewInventory(10)
	inv.SetSlot(3, item.NewItemStack(item.Diamond, 5))

	idx, found := inv.FindItem(item.Diamond)
	assert.True(t, found)
	assert.Equal(t, 3, idx)
}

func TestFindItem_NotFound(t *testing.T) {
	inv := NewInventory(5)
	_, found := inv.FindItem(item.Diamond)
	assert.False(t, found)
}

func TestClear(t *testing.T) {
	inv := NewInventory(5)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 10))
	inv.SetSlot(4, item.NewItemStack(item.Stone, 20))

	inv.Clear()
	assert.True(t, inv.IsEmpty())
}

func TestIsEmpty(t *testing.T) {
	inv := NewInventory(3)
	assert.True(t, inv.IsEmpty())

	inv.SetSlot(1, item.NewItemStack(item.Sand, 1))
	assert.False(t, inv.IsEmpty())
}

// ---------------------------------------------------------------------------
// CraftingGrid tests
// ---------------------------------------------------------------------------

func TestCraftingGrid_SetSlot_GetSlot(t *testing.T) {
	var g CraftingGrid
	stack := item.NewItemStack(item.OakPlanks, 1)
	g.SetSlot(0, 0, stack)
	got := g.GetSlot(0, 0)
	assert.Equal(t, item.OakPlanks, got.ItemID)
	assert.Equal(t, 1, got.Count)
}

func TestCraftingGrid_OutOfBounds(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(-1, 0, item.NewItemStack(item.Dirt, 1))
	g.SetSlot(0, 5, item.NewItemStack(item.Dirt, 1))
	assert.True(t, g.GetSlot(-1, 0).IsEmpty())
	assert.True(t, g.GetSlot(0, 5).IsEmpty())
}

func TestCraftingGrid_Clear(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(1, 1, item.NewItemStack(item.Dirt, 1))
	g.Clear()
	assert.True(t, g.GetSlot(1, 1).IsEmpty())
}

func TestCraftPlanks(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakLog, 1))

	result := g.GetResult()
	assert.Equal(t, item.OakPlanks, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftPlanks_Shapeless_AnyPosition(t *testing.T) {
	var g CraftingGrid
	// Place OakLog in center instead of top-left.
	g.SetSlot(1, 1, item.NewItemStack(item.OakLog, 1))

	result := g.GetResult()
	assert.Equal(t, item.OakPlanks, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftSticks(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.Stick, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftCraftingTable(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.CraftingTable, result.ItemID)
	assert.Equal(t, 1, result.Count)
}

func TestCraftWoodenPickaxe(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Stick, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.WoodenPickaxe, result.ItemID)
	assert.Equal(t, 1, result.Count)
}

func TestCraftDiamondPickaxe(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Diamond, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.Diamond, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.Diamond, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Stick, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.DiamondPickaxe, result.ItemID)
}

func TestCraftFurnace(t *testing.T) {
	var g CraftingGrid
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if r == 1 && c == 1 {
				continue // center is empty
			}
			g.SetSlot(r, c, item.NewItemStack(item.Cobblestone, 1))
		}
	}

	result := g.GetResult()
	assert.Equal(t, item.Furnace, result.ItemID)
	assert.Equal(t, 1, result.Count)
}

func TestCraftChest(t *testing.T) {
	var g CraftingGrid
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if r == 1 && c == 1 {
				continue
			}
			g.SetSlot(r, c, item.NewItemStack(item.OakPlanks, 1))
		}
	}

	result := g.GetResult()
	assert.Equal(t, item.Chest, result.ItemID)
}

func TestCraftTorch(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Coal, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.Torch, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftAxe(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Stick, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.WoodenAxe, result.ItemID)
}

func TestCraftShovel(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.Stick, 1))
	g.SetSlot(2, 0, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.WoodenShovel, result.ItemID)
}

func TestCraftSword(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 0, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.WoodenSword, result.ItemID)
}

func TestInvalidRecipe_ReturnsEmpty(t *testing.T) {
	var g CraftingGrid
	// Random arrangement that matches no recipe.
	g.SetSlot(0, 0, item.NewItemStack(item.Dirt, 1))
	g.SetSlot(2, 2, item.NewItemStack(item.Sand, 1))

	result := g.GetResult()
	assert.True(t, result.IsEmpty())
}

func TestCraft_DecrementsGrid(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakLog, 3))

	result, ok := g.Craft()
	assert.True(t, ok)
	assert.Equal(t, item.OakPlanks, result.ItemID)
	assert.Equal(t, 4, result.Count)

	// OakLog count should have decreased by 1.
	assert.Equal(t, 2, g.GetSlot(0, 0).Count)
}

func TestCraft_RemovesEmptySlots(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakLog, 1))

	result, ok := g.Craft()
	assert.True(t, ok)
	assert.Equal(t, item.OakPlanks, result.ItemID)
	assert.True(t, g.GetSlot(0, 0).IsEmpty())
}

func TestCraft_NoMatch(t *testing.T) {
	var g CraftingGrid
	result, ok := g.Craft()
	assert.False(t, ok)
	assert.True(t, result.IsEmpty())
}

func TestShapedRecipe_OffsetMatching(t *testing.T) {
	// Sticks recipe (2 planks vertical) placed at col 2 instead of col 0.
	var g CraftingGrid
	g.SetSlot(0, 2, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 2, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.Stick, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

// ---------------------------------------------------------------------------
// Smelting tests
// ---------------------------------------------------------------------------

func TestFindSmeltingRecipe(t *testing.T) {
	recipe, found := FindSmeltingRecipe(item.IronOre)
	assert.True(t, found)
	assert.Equal(t, item.IronIngot, recipe.Output)
	assert.Equal(t, 10.0, recipe.Duration)
}

func TestFindSmeltingRecipe_NotFound(t *testing.T) {
	_, found := FindSmeltingRecipe(item.Diamond)
	assert.False(t, found)
}

func TestFurnace_BasicSmelting(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	// Smelt for the full duration.
	f.Update(10.0)

	assert.True(t, f.InputSlot.IsEmpty())
	assert.Equal(t, item.IronIngot, f.OutputSlot.ItemID)
	assert.Equal(t, 1, f.OutputSlot.Count)
}

func TestFurnace_PartialProgress(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(5.0)
	assert.Equal(t, 5.0, f.Progress)
	assert.Equal(t, 1, f.InputSlot.Count) // not yet consumed
	assert.True(t, f.OutputSlot.IsEmpty())

	f.Update(5.0)
	assert.Equal(t, 0.0, f.Progress)
	assert.True(t, f.InputSlot.IsEmpty())
	assert.Equal(t, 1, f.OutputSlot.Count)
}

func TestFurnace_MultipleSmelts(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 3)
	f.FuelSlot = item.NewItemStack(item.Coal, 3)

	f.Update(30.0)

	assert.True(t, f.InputSlot.IsEmpty())
	assert.Equal(t, 3, f.OutputSlot.Count)
	assert.Equal(t, item.IronIngot, f.OutputSlot.ItemID)
}

func TestFurnace_NoFuel(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	// No fuel.

	f.Update(10.0)
	assert.Equal(t, 0.0, f.Progress)
	assert.Equal(t, 1, f.InputSlot.Count) // unchanged
	assert.True(t, f.OutputSlot.IsEmpty())
}

func TestFurnace_NoInput(t *testing.T) {
	f := NewFurnace()
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(10.0)
	assert.Equal(t, 0.0, f.Progress)
	assert.Equal(t, 1, f.FuelSlot.Count) // fuel not consumed
}

func TestFurnace_FuelConsumption(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 2)
	f.FuelSlot = item.NewItemStack(item.Coal, 2)

	// First smelt consumes 1 fuel.
	f.Update(10.0)
	assert.Equal(t, 1, f.FuelSlot.Count)
	assert.Equal(t, 1, f.OutputSlot.Count)

	// Second smelt consumes another fuel.
	f.Update(10.0)
	assert.True(t, f.FuelSlot.IsEmpty())
	assert.Equal(t, 2, f.OutputSlot.Count)
}

func TestFurnace_InvalidRecipe(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.Diamond, 1) // No smelting recipe
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(10.0)
	assert.Equal(t, 0.0, f.Progress)
	assert.True(t, f.OutputSlot.IsEmpty())
}

func TestFurnace_OutputFull(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)
	f.OutputSlot = item.NewItemStack(item.IronIngot, 64) // full

	f.Update(10.0)
	assert.Equal(t, 1, f.InputSlot.Count) // unchanged
	assert.Equal(t, 64, f.OutputSlot.Count)
}

func TestFurnace_OutputMismatch(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)
	f.OutputSlot = item.NewItemStack(item.GoldIngot, 1) // wrong type

	f.Update(10.0)
	assert.Equal(t, 1, f.InputSlot.Count) // unchanged
}

func TestFurnace_ZeroDt(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(0)
	assert.Equal(t, 0.0, f.Progress)
}

func TestFurnace_NegativeDt(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(-5.0)
	assert.Equal(t, 0.0, f.Progress)
}

func TestSmeltingRecipes_AllRegistered(t *testing.T) {
	expectedInputs := []uint16{item.IronOre, item.GoldOre, item.Sand, item.Cobblestone, item.OakLog}
	for _, id := range expectedInputs {
		_, found := FindSmeltingRecipe(id)
		assert.True(t, found, "smelting recipe not found for item ID %d", id)
	}
}

func TestGetSmeltingRecipes(t *testing.T) {
	all := GetSmeltingRecipes()
	assert.GreaterOrEqual(t, len(all), 5)
}

func TestGetRecipes(t *testing.T) {
	all := GetRecipes()
	// We registered at least: planks, sticks, crafting table, 4 pickaxes,
	// 4 axes, 4 shovels, 4 swords, furnace, chest, torch, 4 hoes,
	// bucket, ladder, oak door, oak fence, oak fence gate, shears, boat,
	// 3 mineral blocks, 3 decompositions, 6 stairs, 6 slabs = 45+
	assert.GreaterOrEqual(t, len(all), 45)
}

// ---------------------------------------------------------------------------
// Inventory full handling edge cases
// ---------------------------------------------------------------------------

func TestAddItem_DifferentItems_MixedSlots(t *testing.T) {
	inv := NewInventory(2)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 32))

	// Add stone; should go in slot 1 (not merge with dirt).
	remaining := inv.AddItem(item.NewItemStack(item.Stone, 10))
	assert.True(t, remaining.IsEmpty())
	assert.Equal(t, item.Stone, inv.GetSlot(1).ItemID)
	assert.Equal(t, 10, inv.GetSlot(1).Count)
}

func TestAddItem_FullInventory_DifferentItem(t *testing.T) {
	inv := NewInventory(1)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 32))

	remaining := inv.AddItem(item.NewItemStack(item.Stone, 5))
	assert.Equal(t, 5, remaining.Count)
	assert.Equal(t, item.Stone, remaining.ItemID)
}

func TestInventory_FindItem_FirstOccurrence(t *testing.T) {
	inv := NewInventory(5)
	inv.SetSlot(1, item.NewItemStack(item.Coal, 5))
	inv.SetSlot(3, item.NewItemStack(item.Coal, 10))

	idx, found := inv.FindItem(item.Coal)
	assert.True(t, found)
	assert.Equal(t, 1, idx) // first occurrence
}

// ---------------------------------------------------------------------------
// Shaped recipe offset matching
// ---------------------------------------------------------------------------

func TestCraftingTable_BottomRight(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(1, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 2, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 2, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.CraftingTable, result.ItemID)
}

func TestPickaxe_Offset(t *testing.T) {
	// Stone pickaxe placed starting at row 0 col 0.
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Stick, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Stick, 1))

	result := g.GetResult()
	assert.Equal(t, item.StonePickaxe, result.ItemID)
}

func TestFurnace_SmeltSand(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.Sand, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(10.0)
	assert.Equal(t, item.Glass, f.OutputSlot.ItemID)
	assert.Equal(t, 1, f.OutputSlot.Count)
}

func TestFurnace_SmeltOakLog(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.OakLog, 1)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)

	f.Update(10.0)
	assert.Equal(t, item.Coal, f.OutputSlot.ItemID)
	assert.Equal(t, 1, f.OutputSlot.Count)
}

func TestFurnace_BurnTimeCarriesOver(t *testing.T) {
	f := NewFurnace()
	f.InputSlot = item.NewItemStack(item.IronOre, 2)
	f.FuelSlot = item.NewItemStack(item.Coal, 1)
	// One coal burns for 10s, smelting takes 10s per item.
	// First item should smelt; second should fail (no fuel left).
	f.Update(10.0)
	assert.Equal(t, 1, f.OutputSlot.Count)

	// Try to smelt second: no fuel, no burn time remaining.
	f.Update(10.0)
	assert.Equal(t, 1, f.OutputSlot.Count) // unchanged
}

// ---------------------------------------------------------------------------
// Stair / slab recipe tests
// ---------------------------------------------------------------------------

func TestCraftOakStairs(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(2, 2, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.OakStairs, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftCobblestoneStairs(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(2, 0, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(2, 2, item.NewItemStack(item.Cobblestone, 1))

	result := g.GetResult()
	assert.Equal(t, item.CobblestoneStairs, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftStoneStairs(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Stone, 1))
	g.SetSlot(1, 0, item.NewItemStack(item.Stone, 1))
	g.SetSlot(1, 1, item.NewItemStack(item.Stone, 1))
	g.SetSlot(2, 0, item.NewItemStack(item.Stone, 1))
	g.SetSlot(2, 1, item.NewItemStack(item.Stone, 1))
	g.SetSlot(2, 2, item.NewItemStack(item.Stone, 1))

	result := g.GetResult()
	assert.Equal(t, item.StoneStairs, result.ItemID)
	assert.Equal(t, 4, result.Count)
}

func TestCraftOakSlab(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.OakPlanks, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.OakPlanks, 1))

	result := g.GetResult()
	assert.Equal(t, item.OakSlab, result.ItemID)
	assert.Equal(t, 6, result.Count)
}

func TestCraftCobblestoneSlab(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.Cobblestone, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.Cobblestone, 1))

	result := g.GetResult()
	assert.Equal(t, item.CobblestoneSlab, result.ItemID)
	assert.Equal(t, 6, result.Count)
}

func TestCraftStoneSlab(t *testing.T) {
	var g CraftingGrid
	g.SetSlot(0, 0, item.NewItemStack(item.Stone, 1))
	g.SetSlot(0, 1, item.NewItemStack(item.Stone, 1))
	g.SetSlot(0, 2, item.NewItemStack(item.Stone, 1))

	result := g.GetResult()
	assert.Equal(t, item.StoneSlab, result.ItemID)
	assert.Equal(t, 6, result.Count)
}

func TestStairRecipesRegistered(t *testing.T) {
	stairItems := []uint16{
		item.OakStairs, item.CobblestoneStairs, item.StoneStairs,
		item.BirchStairs, item.SpruceStairs, item.SandstoneStairs,
	}
	all := GetRecipes()
	for _, stairID := range stairItems {
		found := false
		for _, r := range all {
			if r.Result.ItemID == stairID {
				found = true
				assert.Equal(t, 4, r.Result.Count, "stair recipe for item %d should yield 4", stairID)
				break
			}
		}
		assert.True(t, found, "stair recipe not found for item ID %d", stairID)
	}
}

func TestSlabRecipesRegistered(t *testing.T) {
	slabItems := []uint16{
		item.OakSlab, item.CobblestoneSlab, item.StoneSlab,
		item.BirchSlab, item.SpruceSlab, item.SandstoneSlab,
	}
	all := GetRecipes()
	for _, slabID := range slabItems {
		found := false
		for _, r := range all {
			if r.Result.ItemID == slabID {
				found = true
				assert.Equal(t, 6, r.Result.Count, "slab recipe for item %d should yield 6", slabID)
				break
			}
		}
		assert.True(t, found, "slab recipe not found for item ID %d", slabID)
	}
}
