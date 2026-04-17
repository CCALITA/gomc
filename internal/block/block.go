// Package block defines block types, their properties, and face geometry
// for a Minecraft-like voxel world.
package block

// BlockID is the numeric identifier for a block type.
type BlockID = uint16

// BlockProperties describes the physical and visual attributes of a block type.
type BlockProperties struct {
	Name            string
	Solid           bool
	Transparent     bool
	Hardness        float32
	BlastResistance float32
	LightEmission   uint8
	LightFilter     uint8
}

// Predefined block IDs.
const (
	Air           BlockID = 0
	Stone         BlockID = 1
	Dirt          BlockID = 2
	Grass         BlockID = 3
	Sand          BlockID = 4
	Gravel        BlockID = 5
	OakLog        BlockID = 6
	OakLeaves     BlockID = 7
	OakPlanks     BlockID = 8
	Cobblestone   BlockID = 9
	Glass         BlockID = 10
	Water         BlockID = 11
	Lava          BlockID = 12
	IronOre       BlockID = 13
	CoalOre       BlockID = 14
	DiamondOre    BlockID = 15
	GoldOre       BlockID = 16
	Bedrock       BlockID = 17
	CraftingTable BlockID = 18
	Furnace       BlockID = 19
	Chest         BlockID = 20
	Torch         BlockID = 21
	Obsidian      BlockID = 22
	Sandstone     BlockID = 23
)
