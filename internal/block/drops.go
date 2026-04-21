package block

import (
	"github.com/fanxiyao/gomc/internal/item"
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
	Jukebox:       item.JukeboxItem,
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
