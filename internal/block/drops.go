package block

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Drop represents a single item drop from a broken block.
type Drop struct {
	ItemID uint16
	Count  int
	Chance float64 // 0.0 to 1.0; 1.0 = always drops
}

// blockToItemID maps block IDs to their corresponding item IDs.
// Block and item constants use separate numbering so an explicit
// mapping is necessary.
var blockToItemID = map[BlockID]item.ItemID{
	Stone:         item.Stone,
	Dirt:          item.Dirt,
	Grass:         item.Grass,
	Sand:          item.Sand,
	Gravel:        item.Gravel,
	OakLog:        item.OakLog,
	OakLeaves:     item.OakLeaves,
	OakPlanks:     item.OakPlanks,
	Cobblestone:   item.Cobblestone,
	Glass:         item.Glass,
	IronOre:       item.IronOre,
	CoalOre:       item.CoalOre,
	DiamondOre:    item.DiamondOre,
	GoldOre:       item.GoldOre,
	CraftingTable: item.CraftingTable,
	Furnace:       item.Furnace,
	Chest:         item.Chest,
	Torch:         item.Torch,
	Obsidian:      item.Obsidian,
	Sandstone:     Sandstone,
	Bedrock:       item.Bedrock,
	Bed:           item.BedItem,

	// Wool colors
	WhiteWool:     item.Wool,
	OrangeWool:    item.Wool,
	MagentaWool:   item.Wool,
	LightBlueWool: item.Wool,
	YellowWool:    item.Wool,
	LimeWool:      item.Wool,
	PinkWool:      item.Wool,
	GrayWool:      item.Wool,
	LightGrayWool: item.Wool,
	CyanWool:      item.Wool,
	PurpleWool:    item.Wool,
	BlueWool:      item.Wool,
	BrownWool:     item.Wool,
	GreenWool:     item.Wool,
	RedWool:       item.Wool,
	BlackWool:     item.Wool,

	// Decorative / functional blocks
	OakDoor:      item.OakDoor,
	OakFence:     item.OakFence,
	Ladder:       item.Ladder,
	Bookshelf:    item.Bookshelf,
	IronBlock:    item.IronBlock,
	GoldBlock:    item.GoldBlock,
	DiamondBlock: item.DiamondBlock,

	// Stairs
	OakStairs:         item.OakStairs,
	CobblestoneStairs: item.CobblestoneStairs,
	StoneStairs:       item.StoneStairs,
	BirchStairs:       item.BirchStairs,
	SpruceStairs:      item.SpruceStairs,
	SandstoneStairs:   item.SandstoneStairs,

	// Slabs
	OakSlab:         item.OakSlab,
	CobblestoneSlab: item.CobblestoneSlab,
	StoneSlab:       item.StoneSlab,
	BirchSlab:       item.BirchSlab,
	SpruceSlab:      item.SpruceSlab,
	SandstoneSlab:   item.SandstoneSlab,
}

// toolCategory maps block IDs to the tool type that is effective
// against them.
var toolCategory = map[BlockID]string{
	Stone:       item.ToolPickaxe,
	Cobblestone: item.ToolPickaxe,
	IronOre:     item.ToolPickaxe,
	CoalOre:     item.ToolPickaxe,
	DiamondOre:  item.ToolPickaxe,
	GoldOre:     item.ToolPickaxe,
	Obsidian:    item.ToolPickaxe,
	Sandstone:   item.ToolPickaxe,
	OakLog:      item.ToolAxe,
	OakPlanks:   item.ToolAxe,
	OakLeaves:   item.ToolAxe,
	Dirt:        item.ToolShovel,
	Grass:       item.ToolShovel,
	Sand:        item.ToolShovel,
	Gravel:      item.ToolShovel,

	// Wood blocks – axe
	OakDoor:           item.ToolAxe,
	OakFence:          item.ToolAxe,
	Ladder:            item.ToolAxe,
	Bookshelf:         item.ToolAxe,
	OakStairs:         item.ToolAxe,
	BirchStairs:       item.ToolAxe,
	SpruceStairs:      item.ToolAxe,
	OakSlab:           item.ToolAxe,
	BirchSlab:         item.ToolAxe,
	SpruceSlab:        item.ToolAxe,

	// Stone stairs/slabs – pickaxe
	CobblestoneStairs: item.ToolPickaxe,
	StoneStairs:       item.ToolPickaxe,
	SandstoneStairs:   item.ToolPickaxe,
	CobblestoneSlab:   item.ToolPickaxe,
	StoneSlab:         item.ToolPickaxe,
	SandstoneSlab:     item.ToolPickaxe,

	// Metal / mineral blocks – pickaxe
	IronBlock:    item.ToolPickaxe,
	GoldBlock:    item.ToolPickaxe,
	DiamondBlock: item.ToolPickaxe,
}

// minToolLevel specifies the minimum tool level required to obtain
// drops from certain blocks. Blocks not listed have no requirement.
var minToolLevel = map[BlockID]int{
	DiamondOre: item.LevelIron,
	IronOre:    item.LevelStone,
	GoldOre:    item.LevelStone,
}

// GetDrops returns the items that drop when the given block is broken
// with the specified tool type and tool level. The returned slice may
// be empty (e.g. glass drops nothing, or diamond ore mined without a
// sufficient pickaxe).
func GetDrops(blockID uint16, toolType string, toolLevel int) []Drop {
	switch blockID {
	case Air, Water, Lava, FlowingWater, FlowingLava, Bedrock:
		return nil

	case Glass:
		return nil

	case Stone:
		return []Drop{{ItemID: item.Cobblestone, Count: 1, Chance: 1.0}}

	case Grass:
		return []Drop{{ItemID: item.Dirt, Count: 1, Chance: 1.0}}

	case CoalOre:
		return []Drop{{ItemID: item.Coal, Count: 1, Chance: 1.0}}

	case DiamondOre:
		if toolType != item.ToolPickaxe || toolLevel < item.LevelIron {
			return nil
		}
		return []Drop{{ItemID: item.Diamond, Count: 1, Chance: 1.0}}

	case IronOre:
		if toolType != item.ToolPickaxe || toolLevel < item.LevelStone {
			return nil
		}
		return []Drop{{ItemID: item.IronOre, Count: 1, Chance: 1.0}}

	case GoldOre:
		if toolType != item.ToolPickaxe || toolLevel < item.LevelStone {
			return nil
		}
		return []Drop{{ItemID: item.GoldOre, Count: 1, Chance: 1.0}}

	case OakLeaves:
		return []Drop{{ItemID: OakLeaves, Count: 1, Chance: 0.1}}

	default:
		// Bed blocks encode state in the upper bits; use the base ID.
		if IsBed(blockID) {
			return []Drop{{ItemID: item.BedItem, Count: 1, Chance: 1.0}}
		}
		// Default: the block drops its corresponding item.
		itemID, ok := blockToItemID[blockID]
		if !ok {
			return nil
		}
		return []Drop{{ItemID: itemID, Count: 1, Chance: 1.0}}
	}
}

// GetToolCategory returns the tool type that is effective against the
// given block. Returns item.ToolNone if no specific tool is preferred.
func GetToolCategory(blockID uint16) string {
	if cat, ok := toolCategory[blockID]; ok {
		return cat
	}
	return item.ToolNone
}

// GetMinToolLevel returns the minimum tool level required to receive
// drops from the given block. Returns item.LevelHand (0) when any
// tool (or bare hand) is sufficient.
func GetMinToolLevel(blockID uint16) int {
	if lvl, ok := minToolLevel[blockID]; ok {
		return lvl
	}
	return item.LevelHand
}

// CalculateBreakSpeed returns the break-time multiplier for mining a
// block with the given tool. The formula is:
//
//	baseTime * hardness / toolMultiplier
//
// where toolMultiplier is 2.0 for a matching tool category, plus 0.5
// for each tool level above the block's minimum requirement. Bare-hand
// (or wrong tool) gives a multiplier of 1.0.
func CalculateBreakSpeed(blockID uint16, toolType string, toolLevel int) float32 {
	props := GetProperties(blockID)
	if props.Hardness <= 0 {
		return 0
	}

	multiplier := float32(1.0)
	cat := GetToolCategory(blockID)
	if cat != item.ToolNone && toolType == cat {
		multiplier = 2.0
		minLevel := GetMinToolLevel(blockID)
		if toolLevel > minLevel {
			multiplier += float32(toolLevel-minLevel) * 0.5
		}
	}

	return props.Hardness / multiplier
}

// SpawnDrops evaluates the drops for the given block and spawns item
// entities at the centre of the block position. toolType and toolLevel
// describe the tool used to break the block.
func SpawnDrops(w *ecs.World, pos mcmath.BlockPos, blockID uint16, toolType string, toolLevel int) {
	drops := GetDrops(blockID, toolType, toolLevel)
	// Offset to centre of block (+0.5 on X and Z, +0.25 above floor).
	spawnPos := mcmath.Vec3{
		X: float32(pos.X) + 0.5,
		Y: float32(pos.Y) + 0.25,
		Z: float32(pos.Z) + 0.5,
	}
	for _, d := range drops {
		if d.Chance < 1.0 {
			// Probabilistic drops are skipped here; caller should roll
			// the dice and filter before calling SpawnDrops, or use this
			// for guaranteed drops only.
			continue
		}
		entity.SpawnItemDrop(w, spawnPos, d.ItemID, d.Count)
	}
}
