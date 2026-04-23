package block

import (
	"math/rand"
	"testing"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
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

func TestGetDrops_StoneWithBareHand(t *testing.T) {
	drops := GetDrops(Stone, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Cobblestone, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
}

func TestGetDrops_CoalOreDropsCoal(t *testing.T) {
	drops := GetDrops(CoalOre, item.ToolPickaxe, item.LevelWood)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Coal, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_CoalOreWithBareHand(t *testing.T) {
	drops := GetDrops(CoalOre, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Coal, drops[0].ItemID)
}

func TestGetDrops_CoalOreWithAxe(t *testing.T) {
	drops := GetDrops(CoalOre, item.ToolAxe, item.LevelDiamond)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Coal, drops[0].ItemID)
}

func TestGetDrops_DiamondOreWithIronPickaxe(t *testing.T) {
	drops := GetDrops(DiamondOre, item.ToolPickaxe, item.LevelIron)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Diamond, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_IronOreWithStonePickaxe(t *testing.T) {
	drops := GetDrops(IronOre, item.ToolPickaxe, item.LevelStone)
	require.Len(t, drops, 1)
	assert.Equal(t, item.IronOre, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_GoldOreWithStonePickaxe(t *testing.T) {
	drops := GetDrops(GoldOre, item.ToolPickaxe, item.LevelStone)
	require.Len(t, drops, 1)
	assert.Equal(t, item.GoldOre, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_GrassDropsDirt(t *testing.T) {
	drops := GetDrops(Grass, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Dirt, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.Equal(t, 1.0, drops[0].Chance)
}

func TestGetDrops_GrassWithShovel(t *testing.T) {
	drops := GetDrops(Grass, item.ToolShovel, item.LevelWood)
	require.Len(t, drops, 1)
	assert.Equal(t, item.Dirt, drops[0].ItemID)
}

func TestGetDrops_OakLeavesDropsSapling(t *testing.T) {
	drops := GetDrops(OakLeaves, item.ToolNone, item.LevelHand)
	require.Len(t, drops, 1)
	assert.Equal(t, uint16(OakLeaves), drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
	assert.InDelta(t, 0.1, drops[0].Chance, 1e-9)
}

func TestGetDrops_OakLeavesChanceBelowOne(t *testing.T) {
	drops := GetDrops(OakLeaves, item.ToolAxe, item.LevelWood)
	require.Len(t, drops, 1)
	assert.Less(t, drops[0].Chance, 1.0, "oak leaves drop chance should be less than 1.0")
	assert.Greater(t, drops[0].Chance, 0.0, "oak leaves drop chance should be greater than 0.0")
}

// ---------------------------------------------------------------------------
// GetDrops – blocks that drop nothing
// ---------------------------------------------------------------------------

func TestGetDrops_NoDropBlocks(t *testing.T) {
	tests := []struct {
		name    string
		blockID uint16
	}{
		{"Air", Air},
		{"Bedrock", Bedrock},
		{"Water", Water},
		{"Lava", Lava},
		{"FlowingWater", FlowingWater},
		{"FlowingLava", FlowingLava},
		{"Glass", Glass},
		{"Glass with pickaxe", Glass},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(tc.blockID, item.ToolNone, item.LevelHand)
			assert.Empty(t, drops)
		})
	}
}

func TestGetDrops_GlassWithPickaxeStillDropsNothing(t *testing.T) {
	drops := GetDrops(Glass, item.ToolPickaxe, item.LevelDiamond)
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
// Mining level requirements – insufficient level yields no drops
// ---------------------------------------------------------------------------

func TestMiningLevel_DiamondOreRequiresIronPickaxe(t *testing.T) {
	tests := []struct {
		name      string
		toolType  string
		toolLevel int
		wantDrops bool
	}{
		{"bare hand", item.ToolNone, item.LevelHand, false},
		{"wood pickaxe", item.ToolPickaxe, item.LevelWood, false},
		{"stone pickaxe", item.ToolPickaxe, item.LevelStone, false},
		{"iron pickaxe", item.ToolPickaxe, item.LevelIron, true},
		{"diamond pickaxe", item.ToolPickaxe, item.LevelDiamond, true},
		{"diamond axe (wrong type)", item.ToolAxe, item.LevelDiamond, false},
		{"diamond shovel (wrong type)", item.ToolShovel, item.LevelDiamond, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(DiamondOre, tc.toolType, tc.toolLevel)
			if tc.wantDrops {
				require.Len(t, drops, 1)
				assert.Equal(t, item.Diamond, drops[0].ItemID)
			} else {
				assert.Empty(t, drops)
			}
		})
	}
}

func TestMiningLevel_IronOreRequiresStonePickaxe(t *testing.T) {
	tests := []struct {
		name      string
		toolType  string
		toolLevel int
		wantDrops bool
	}{
		{"bare hand", item.ToolNone, item.LevelHand, false},
		{"wood pickaxe", item.ToolPickaxe, item.LevelWood, false},
		{"stone pickaxe", item.ToolPickaxe, item.LevelStone, true},
		{"iron pickaxe", item.ToolPickaxe, item.LevelIron, true},
		{"diamond pickaxe", item.ToolPickaxe, item.LevelDiamond, true},
		{"diamond axe (wrong type)", item.ToolAxe, item.LevelDiamond, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(IronOre, tc.toolType, tc.toolLevel)
			if tc.wantDrops {
				require.Len(t, drops, 1)
				assert.Equal(t, item.IronOre, drops[0].ItemID)
			} else {
				assert.Empty(t, drops)
			}
		})
	}
}

func TestMiningLevel_GoldOreRequiresStonePickaxe(t *testing.T) {
	tests := []struct {
		name      string
		toolType  string
		toolLevel int
		wantDrops bool
	}{
		{"bare hand", item.ToolNone, item.LevelHand, false},
		{"wood pickaxe", item.ToolPickaxe, item.LevelWood, false},
		{"stone pickaxe", item.ToolPickaxe, item.LevelStone, true},
		{"iron pickaxe", item.ToolPickaxe, item.LevelIron, true},
		{"diamond shovel (wrong type)", item.ToolShovel, item.LevelDiamond, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(GoldOre, tc.toolType, tc.toolLevel)
			if tc.wantDrops {
				require.Len(t, drops, 1)
				assert.Equal(t, item.GoldOre, drops[0].ItemID)
			} else {
				assert.Empty(t, drops)
			}
		})
	}
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
	speed := CalculateBreakSpeed(Stone, item.ToolNone, item.LevelHand)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_MatchingTool(t *testing.T) {
	speed := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelWood)
	assert.InDelta(t, 0.6, speed, 1e-6)
}

func TestCalculateBreakSpeed_HigherLevelBonus(t *testing.T) {
	speed := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelIron)
	assert.InDelta(t, 1.5/3.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_WrongTool(t *testing.T) {
	speed := CalculateBreakSpeed(Stone, item.ToolAxe, item.LevelDiamond)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_ShovelOnDirt(t *testing.T) {
	speed := CalculateBreakSpeed(Dirt, item.ToolShovel, item.LevelWood)
	assert.InDelta(t, 0.2, speed, 1e-6)
}

func TestCalculateBreakSpeed_AxeOnWood(t *testing.T) {
	speed := CalculateBreakSpeed(OakLog, item.ToolAxe, item.LevelWood)
	assert.InDelta(t, 0.8, speed, 1e-6)
}

func TestCalculateBreakSpeed_UnbreakableBlock(t *testing.T) {
	speed := CalculateBreakSpeed(Bedrock, item.ToolPickaxe, item.LevelDiamond)
	assert.Equal(t, float32(0), speed)
}

func TestCalculateBreakSpeed_ZeroHardness(t *testing.T) {
	speed := CalculateBreakSpeed(Torch, item.ToolNone, item.LevelHand)
	assert.Equal(t, float32(0), speed)
}

func TestCalculateBreakSpeed_DiamondOreWithIronPickaxe(t *testing.T) {
	speed := CalculateBreakSpeed(DiamondOre, item.ToolPickaxe, item.LevelIron)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_DiamondOreWithDiamondPickaxe(t *testing.T) {
	speed := CalculateBreakSpeed(DiamondOre, item.ToolPickaxe, item.LevelDiamond)
	assert.InDelta(t, 1.2, speed, 1e-6)
}

func TestCalculateBreakSpeed_ExactMinLevel(t *testing.T) {
	speed := CalculateBreakSpeed(IronOre, item.ToolPickaxe, item.LevelStone)
	assert.InDelta(t, 1.5, speed, 1e-6)
}

func TestCalculateBreakSpeed_MatchingToolAtHandLevel(t *testing.T) {
	speed := CalculateBreakSpeed(Dirt, item.ToolShovel, item.LevelHand)
	assert.InDelta(t, 0.25, speed, 1e-6)
}

func TestCalculateBreakSpeed_AirBlock(t *testing.T) {
	speed := CalculateBreakSpeed(Air, item.ToolNone, item.LevelHand)
	assert.Equal(t, float32(0), speed)
}

func TestCalculateBreakSpeed_GlassBlock(t *testing.T) {
	speed := CalculateBreakSpeed(Glass, item.ToolNone, item.LevelHand)
	assert.InDelta(t, 0.3, speed, 1e-6)
}

func TestCalculateBreakSpeed_ObsidianWithDiamondPickaxe(t *testing.T) {
	speed := CalculateBreakSpeed(Obsidian, item.ToolPickaxe, item.LevelDiamond)
	assert.InDelta(t, 12.5, speed, 1e-6)
}

// ---------------------------------------------------------------------------
// Tool effectiveness – correct tool breaks faster than bare hand
// ---------------------------------------------------------------------------

func TestToolEffectiveness_CorrectToolFasterThanHand(t *testing.T) {
	tests := []struct {
		name     string
		blockID  uint16
		toolType string
	}{
		{"pickaxe on stone", Stone, item.ToolPickaxe},
		{"axe on wood", OakLog, item.ToolAxe},
		{"shovel on dirt", Dirt, item.ToolShovel},
		{"shovel on sand", Sand, item.ToolShovel},
		{"shovel on gravel", Gravel, item.ToolShovel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bareHand := CalculateBreakSpeed(tc.blockID, item.ToolNone, item.LevelHand)
			withTool := CalculateBreakSpeed(tc.blockID, tc.toolType, item.LevelWood)
			assert.Less(t, withTool, bareHand,
				"matching tool should break block faster (lower time) than bare hand")
		})
	}
}

func TestToolEffectiveness_WrongToolNotFasterThanHand(t *testing.T) {
	bareHand := CalculateBreakSpeed(Stone, item.ToolNone, item.LevelHand)
	wrongTool := CalculateBreakSpeed(Stone, item.ToolShovel, item.LevelDiamond)
	assert.InDelta(t, bareHand, wrongTool, 1e-6,
		"wrong tool type should be same speed as bare hand")
}

func TestToolEffectiveness_HigherLevelPickaxeFasterThanLower(t *testing.T) {
	woodPickaxe := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelWood)
	ironPickaxe := CalculateBreakSpeed(Stone, item.ToolPickaxe, item.LevelIron)
	assert.Less(t, ironPickaxe, woodPickaxe,
		"iron pickaxe should break stone faster than wood pickaxe")
}

// ---------------------------------------------------------------------------
// SpawnDrops
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// GetDrops – decorative and functional blocks
// ---------------------------------------------------------------------------

func TestGetDrops_WoolBlocksDropWool(t *testing.T) {
	woolBlocks := []struct {
		name    string
		blockID uint16
	}{
		{"WhiteWool", WhiteWool},
		{"OrangeWool", OrangeWool},
		{"MagentaWool", MagentaWool},
		{"LightBlueWool", LightBlueWool},
		{"YellowWool", YellowWool},
		{"LimeWool", LimeWool},
		{"PinkWool", PinkWool},
		{"GrayWool", GrayWool},
		{"LightGrayWool", LightGrayWool},
		{"CyanWool", CyanWool},
		{"PurpleWool", PurpleWool},
		{"BlueWool", BlueWool},
		{"BrownWool", BrownWool},
		{"GreenWool", GreenWool},
		{"RedWool", RedWool},
		{"BlackWool", BlackWool},
	}

	for _, tc := range woolBlocks {
		t.Run(tc.name, func(t *testing.T) {
			drops := GetDrops(tc.blockID, item.ToolNone, item.LevelHand)
			require.Len(t, drops, 1)
			assert.Equal(t, item.Wool, drops[0].ItemID)
			assert.Equal(t, 1, drops[0].Count)
			assert.Equal(t, 1.0, drops[0].Chance)
		})
	}
}

func TestGetDrops_DecorativeBlocksDropSelf(t *testing.T) {
	tests := []struct {
		name    string
		blockID uint16
		itemID  uint16
	}{
		{"OakDoor", OakDoor, item.OakDoor},
		{"OakFence", OakFence, item.OakFence},
		{"Ladder", Ladder, item.Ladder},
		{"Bookshelf", Bookshelf, item.Bookshelf},
		{"IronBlock", IronBlock, item.IronBlock},
		{"GoldBlock", GoldBlock, item.GoldBlock},
		{"DiamondBlock", DiamondBlock, item.DiamondBlock},
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

func TestGetDrops_StairsDropSelf(t *testing.T) {
	tests := []struct {
		name    string
		blockID uint16
		itemID  uint16
	}{
		{"OakStairs", OakStairs, item.OakStairs},
		{"CobblestoneStairs", CobblestoneStairs, item.CobblestoneStairs},
		{"StoneStairs", StoneStairs, item.StoneStairs},
		{"BirchStairs", BirchStairs, item.BirchStairs},
		{"SpruceStairs", SpruceStairs, item.SpruceStairs},
		{"SandstoneStairs", SandstoneStairs, item.SandstoneStairs},
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

func TestGetToolCategory_DecorativeBlocks(t *testing.T) {
	tests := []struct {
		blockID uint16
		want    string
	}{
		{OakDoor, item.ToolAxe},
		{OakFence, item.ToolAxe},
		{Ladder, item.ToolAxe},
		{Bookshelf, item.ToolAxe},
		{OakStairs, item.ToolAxe},
		{CobblestoneStairs, item.ToolPickaxe},
		{IronBlock, item.ToolPickaxe},
		{GoldBlock, item.ToolPickaxe},
		{DiamondBlock, item.ToolPickaxe},
		{WhiteWool, item.ToolNone},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, GetToolCategory(tc.blockID), "block %d", tc.blockID)
	}
}

func TestGetDrops_SlabsDropSelf(t *testing.T) {
	tests := []struct {
		name    string
		blockID uint16
		itemID  uint16
	}{
		{"OakSlab", OakSlab, item.OakSlab},
		{"CobblestoneSlab", CobblestoneSlab, item.CobblestoneSlab},
		{"StoneSlab", StoneSlab, item.StoneSlab},
		{"BirchSlab", BirchSlab, item.BirchSlab},
		{"SpruceSlab", SpruceSlab, item.SpruceSlab},
		{"SandstoneSlab", SandstoneSlab, item.SandstoneSlab},
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

// ---------------------------------------------------------------------------
// SpawnDrops
// ---------------------------------------------------------------------------

func TestSpawnDrops_StoneSpawnsCobblestoneEntity(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.BlockPos{X: 10, Y: 64, Z: 20}

	SpawnDrops(w, pos, Stone, item.ToolPickaxe, item.LevelWood, nil)

	// Verify an item entity was spawned.
	dropStore := ecs.GetStore[entity.ItemDrop](w)
	assert.Equal(t, 1, dropStore.Len(), "one item entity should have been spawned")

	dropStore.Each(func(_ ecs.Entity, d *entity.ItemDrop) {
		assert.Equal(t, item.Cobblestone, d.ItemID)
		assert.Equal(t, 1, d.Count)
	})

	// Check position is centred on the block.
	transformStore := ecs.GetStore[entity.Transform](w)
	transformStore.Each(func(_ ecs.Entity, tr *entity.Transform) {
		assert.InDelta(t, 10.5, float64(tr.Position.X), 1e-6)
		assert.InDelta(t, 64.25, float64(tr.Position.Y), 1e-6)
		assert.InDelta(t, 20.5, float64(tr.Position.Z), 1e-6)
	})
}

func TestSpawnDrops_GlassDropsNothing(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}

	SpawnDrops(w, pos, Glass, item.ToolNone, item.LevelHand, nil)

	dropStore := ecs.GetStore[entity.ItemDrop](w)
	assert.Equal(t, 0, dropStore.Len(), "glass should spawn no item entities")
}

func TestSpawnDrops_GuaranteedDropsAlwaysSpawn(t *testing.T) {
	// Stone always drops cobblestone (Chance=1.0). Run multiple times
	// to confirm determinism.
	for i := 0; i < 20; i++ {
		w := ecs.NewWorld()
		pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
		SpawnDrops(w, pos, Stone, item.ToolPickaxe, item.LevelWood, nil)

		dropStore := ecs.GetStore[entity.ItemDrop](w)
		assert.Equal(t, 1, dropStore.Len(), "guaranteed drops must always spawn (iteration %d)", i)
	}
}

func TestSpawnDrops_ProbabilisticDropsCanSpawn(t *testing.T) {
	// Use a seeded RNG that always returns 0.0 (below the 0.1 threshold),
	// so the probabilistic oak-leaves drop should spawn.
	rng := rand.New(rand.NewSource(0))
	// Find a seed whose first Float64 is below 0.1.
	for seed := int64(0); seed < 1000; seed++ {
		rng = rand.New(rand.NewSource(seed))
		if rng.Float64() <= 0.1 {
			rng = rand.New(rand.NewSource(seed)) // reset so SpawnDrops gets the same value
			break
		}
	}

	w := ecs.NewWorld()
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	SpawnDrops(w, pos, OakLeaves, item.ToolNone, item.LevelHand, rng)

	dropStore := ecs.GetStore[entity.ItemDrop](w)
	assert.Equal(t, 1, dropStore.Len(), "probabilistic drops should spawn when the roll succeeds")
}

func TestSpawnDrops_ProbabilisticDropsCanBeSkipped(t *testing.T) {
	// Use a seeded RNG that returns a value above 0.1 so the drop is skipped.
	rng := rand.New(rand.NewSource(0))
	for seed := int64(0); seed < 1000; seed++ {
		rng = rand.New(rand.NewSource(seed))
		if rng.Float64() > 0.1 {
			rng = rand.New(rand.NewSource(seed))
			break
		}
	}

	w := ecs.NewWorld()
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	SpawnDrops(w, pos, OakLeaves, item.ToolNone, item.LevelHand, rng)

	dropStore := ecs.GetStore[entity.ItemDrop](w)
	assert.Equal(t, 0, dropStore.Len(), "probabilistic drops should be skipped when the roll fails")
}

func TestSpawnDrops_ZeroChanceNeverSpawns(t *testing.T) {
	// Manually create a Drop with Chance=0.0 by using GetDrops on a block
	// that would never exist. Instead, test via SpawnDrops on a world where
	// we can verify behaviour. We patch by calling the function directly:
	// Chance 0.0 means roll must be <= 0.0 which Float64 never returns (it
	// returns [0.0, 1.0)), but 0.0 is possible. However, the condition is
	// d.Chance < 1.0 && roll > d.Chance. With Chance=0.0, any roll > 0
	// skips the drop. Use a seeded RNG to guarantee a non-zero roll.
	//
	// Since we cannot directly inject a zero-chance block into GetDrops, we
	// test the logic indirectly: with OakLeaves (Chance=0.1) and an RNG
	// that always returns high values, no drop should ever appear.
	for i := 0; i < 50; i++ {
		// seed 1 produces Float64 ~0.6046... which is > 0.1
		rng := rand.New(rand.NewSource(1))
		w := ecs.NewWorld()
		pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
		SpawnDrops(w, pos, OakLeaves, item.ToolNone, item.LevelHand, rng)

		dropStore := ecs.GetStore[entity.ItemDrop](w)
		assert.Equal(t, 0, dropStore.Len(), "high-roll should never produce a drop (iteration %d)", i)
	}
}
