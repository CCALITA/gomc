package block

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetDrops – specific drop rules
// ---------------------------------------------------------------------------

func TestGetDrops_StoneDropsCobblestone(t *testing.T) {
	drops := GetDrops(Stone, item.ToolPickaxe, item.LevelWood)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Cobblestone, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_CoalOreDropsCoal(t *testing.T) {
	drops := GetDrops(CoalOre, item.ToolPickaxe, item.LevelWood)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Coal, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_DiamondOreWithIronPickaxe(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolPickaxe, item.LevelIron)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Diamond, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_DiamondOreWithDiamondPickaxe(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolPickaxe, item.LevelDiamond)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Diamond, drops[0].ItemID)
}

func TestGetDrops_DiamondOreWithStonePickaxeDropsNothing(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolPickaxe, item.LevelStone)
	assert.Empty(t, drops)
}

func TestGetDrops_DiamondOreWithWoodPickaxeDropsNothing(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolPickaxe, item.LevelWood)
	assert.Empty(t, drops)
}

func TestGetDrops_DiamondOreWithHandDropsNothing(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_DiamondOreWithWrongToolDropsNothing(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolAxe, item.LevelDiamond)
	assert.Empty(t, drops)
}

func TestGetDrops_IronOreWithStonePickaxe(t *testing.T) {
	drops := GetDrops(IronOre, item.ToolPickaxe, item.LevelStone)
	require.Len(t, drops, 1)
	assert.Equal(t, item.IronOre, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
}

func TestGetDrops_IronOreWithWoodPickaxeDropsNothing(t *testing.T) {
	drops := GetDrops(IronOre, item.ToolPickaxe, item.LevelWood)
	assert.Empty(t, drops)
}

func TestGetDrops_IronOreWithHandDropsNothing(t *testing.T) {
	drops := GetDrops(IronOre, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_GoldOreWithStonePickaxe(t *testing.T) {
	drops := GetDrops(GoldOre, item.ToolPickaxe, item.LevelStone)
	require.Len(t, drops, 1)
	assert.Equal(t, item.GoldOre, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
}

func TestGetDrops_GoldOreWithWoodPickaxeDropsNothing(t *testing.T) {
	drops := GetDrops(GoldOre, item.ToolPickaxe, item.LevelWood)
	assert.Empty(t, drops)
}

func TestGetDrops_GrassDropsDirt(t *testing.T) {
	drops := GetDrops(Grass, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Dirt, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_OakLeavesDropsSapling(t *testing.T) {
	drops := GetDrops(OakLeaves, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, uint16(OakLeaves), drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.InDelta(t, 0.1, drops[0].Chance, 1e-9)
}

func TestGetDrops_GlassDropsNothing(t *testing.T) {
	drops := GetDrops(Glass, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_AirDropsNothing(t *testing.T) {
	drops := GetDrops(Air, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_BedrockDropsNothing(t *testing.T) {
	drops := GetDrops(Bedrock, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_WaterDropsNothing(t *testing.T) {
	drops := GetDrops(Water, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

func TestGetDrops_LavaDropsNothing(t *testing.T) {
	drops := GetDrops(Lava, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

// ---------------------------------------------------------------------------
// GetDrops – default (block drops itself)
// ---------------------------------------------------------------------------

func TestGetDrops_DefaultDropsSelf(t *testing.T) {
	tests := []struct {
		name    string
		blockID uint16
		itemID  uint16
	}{
		{"Dirt", Dirt, item.Dirt},
		{"Sand", Sand, item.Sand},
		{"OakLog", OakLog, item.OakLog},
		{"OakPlanks", OakPlanks, item.OakPlanks},
		{"Cobblestone", Cobblestone, item.Cobblestone},
		{"CraftingTable", CraftingTable, item.CraftingTable},
		{"Furnace", Furnace, item.Furnace},
		{"Chest", Chest, item.Chest},
		{"Torch", Torch, item.Torch},
		{"Obsidian", Obsidian, item.Obsidian},
		{"Sandstone", Sandstone, uint16(Sandstone)},
		{"Gravel", Gravel, item.Gravel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(tc.blockID, item.ToolNone, item.LevelHand)
			require.Len(t, drops, 1)
			assert.Equal(t, tc.itemID, drops[0].ItemID)
			assert.Equal(t, 1, drops[0].Count)
			assert.Equal(t, 1.0, drops[0].Chance)
		})
	}
}

func TestGetDrops_UnknownBlockDropsNothing(t *testing.T) {
	drops := GetDrops(255, item.ToolNone, item.LevelHand)
	assert.Empty(t, drops)
}

// ---------------------------------------------------------------------------
// GetToolCategory
// ---------------------------------------------------------------------------

func TestGetToolCategory(t *testing.T) {
	tests := []struct {
		blockID uint16
		want    string
	}{
		{Stone, item.ToolPickaxe},
		{Cobblestone, item.ToolPickaxe},
		{IronOre, item.ToolPickaxe},
		{CoalOre, item.ToolPickaxe},
		{DiamondOre, item.ToolPickaxe},
		{GoldOre, item.ToolPickaxe},
		{Obsidian, item.ToolPickaxe},
		{Sandstone, item.ToolPickaxe},
		{OakLog, item.ToolAxe},
		{OakPlanks, item.ToolAxe},
		{OakLeaves, item.ToolAxe},
		{Dirt, item.ToolShovel},
		{Grass, item.ToolShovel},
		{Sand, item.ToolShovel},
		{Gravel, item.ToolShovel},
		{Glass, item.ToolNone},
		{Air, item.ToolNone},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, GetToolCategory(tc.blockID), "block %d", tc.blockID)
	}
}

// ---------------------------------------------------------------------------
// GetMinToolLevel
// ---------------------------------------------------------------------------

func TestGetMinToolLevel(t *testing.T) {
	tests := []struct {
		blockID uint16
		want    int
	}{
		{DiamondOre, item.LevelIron},
		{IronOre, item.LevelStone},
		{GoldOre, item.LevelStone},
		{Stone, item.LevelHand},
		{Dirt, item.LevelHand},
		{OakLog, item.LevelHand},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, GetMinToolLevel(tc.blockID), "block %d", tc.blockID)
	}
}

// ---------------------------------------------------------------------------
// CalculateBreakSpeed
// ---------------------------------------------------------------------------

func TestCalculateBreakSpeed_BareHand(t *testing.T) {
	// Stone hardness = 1.5, no tool = multiplier 1.0
	// breakSpeed = 1.5 / 1.0 = 1.5
	speed := CalculateBreakSpeed(Stone, item.ToolNone, item.LevelHand)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_MatchingTool(t *testing.T) {
	// Stone hardness = 1.5, pickaxe matches, min level = 0 (hand), wood level = 1
	// multiplier = 2.0 + (1-0)*0.5 = 2.5
	// breakSpeed = 1.5 / 2.5 = 0.6
	speed := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelWood)
	assert.InDelta(t, 0.6, speed, 1e-6)
}

func TestCalculateBreakSpeed_HigherLevelBonus(t *testing.T) {
	// Stone hardness = 1.5, min level = 0 (hand), iron pickaxe = level 3
	// multiplier = 2.0 + (3-0)*0.5 = 3.5
	// breakSpeed = 1.5 / 3.5
	speed := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelIron)
	assert.InDelta(t, 1.5/3.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_WrongTool(t *testing.T) {
	// Stone hardness = 1.5, axe on stone = no match = multiplier 1.0
	speed := CalculateBreakSpeed(Stone, item.ToolAxe, item.LevelDiamond)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_ShovelOnDirt(t *testing.T) {
	// Dirt hardness = 0.5, shovel matches, min level = 0 (hand), wood level = 1
	// multiplier = 2.0 + (1-0)*0.5 = 2.5
	// breakSpeed = 0.5 / 2.5 = 0.2
	speed := CalculateBreakSpeed(Dirt, item.ToolShovel, item.LevelWood)
	assert.InDelta(t, 0.2, speed, 1e-6)
}

func TestCalculateBreakSpeed_AxeOnWood(t *testing.T) {
	// OakLog hardness = 2.0, axe matches, min level = 0 (hand), wood level = 1
	// multiplier = 2.0 + (1-0)*0.5 = 2.5
	// breakSpeed = 2.0 / 2.5 = 0.8
	speed := CalculateBreakSpeed(OakLog, item.ToolAxe, item.LevelWood)
	assert.InDelta(t, 0.8, speed, 1e-6)
}

func TestCalculateBreakSpeed_UnbreakableBlock(t *testing.T) {
	// Bedrock hardness = -1
	speed := CalculateBreakSpeed(Bedrock, item.ToolPickaxe, item.LevelDiamond)
	assert.Equal(t, float32(0), speed)
}

func TestCalculateBreakSpeed_ZeroHardness(t *testing.T) {
	// Torch hardness = 0
	speed := CalculateBreakSpeed(Torch, item.ToolNone, item.LevelHand)
	assert.Equal(t, float32(0), speed)
}

func TestCalculateBreakSpeed_DiamondOreWithIronPickaxe(t *testing.T) {
	// DiamondOre hardness = 3.0, pickaxe matches, min level = 3 (iron)
	// multiplier = 2.0 + (3-3)*0.5 = 2.0
	// breakSpeed = 3.0 / 2.0 = 1.5
	speed := CalculateBreakSpeed(DiamondOre, item.ToolPickaxe, item.LevelIron)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_DiamondOreWithDiamondPickaxe(t *testing.T) {
	// DiamondOre hardness = 3.0, pickaxe matches, min level = 3 (iron)
	// diamond level = 4, multiplier = 2.0 + (4-3)*0.5 = 2.5
	// breakSpeed = 3.0 / 2.5 = 1.2
	speed := CalculateBreakSpeed(DiamondOre, item.ToolPickaxe, item.LevelDiamond)
	assert.InDelta(t, 1.2, speed, 1e-6)
}
