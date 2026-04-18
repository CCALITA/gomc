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
	FlowingWater  BlockID = 24
	FlowingLava   BlockID = 25
	Fire          BlockID = 26
)

// fluidBaseMask extracts the base block ID (lower 8 bits) from a block that
// may have a fluid level encoded in the upper bits.
const fluidBaseMask = 0x00FF

// fluidLevelShift is the number of bits to shift to encode/decode a fluid level.
const fluidLevelShift = 8

// FlowingWaterLevel returns a FlowingWater block ID with the given level (0-7)
// encoded in the upper bits. Level 0 is a full block; level 7 is nearly empty.
func FlowingWaterLevel(level int) uint16 {
	return FlowingWater | (uint16(level) << fluidLevelShift)
}

// FlowingLavaLevel returns a FlowingLava block ID with the given level (0-7)
// encoded in the upper bits. Level 0 is a full block; level 7 is nearly empty.
func FlowingLavaLevel(level int) uint16 {
	return FlowingLava | (uint16(level) << fluidLevelShift)
}

// BaseID returns the base block ID, stripping any encoded fluid level from the
// upper bits.
func BaseID(id uint16) uint16 {
	return id & fluidBaseMask
}

// IsFluid reports whether the block is any fluid type (water source, lava source,
// flowing water, or flowing lava).
func IsFluid(id uint16) bool {
	base := BaseID(id)
	return base == Water || base == Lava || base == FlowingWater || base == FlowingLava
}

// FluidLevel returns the flow level encoded in a flowing fluid block.
// Returns 0 for source blocks or non-fluid blocks.
func FluidLevel(id uint16) int {
	base := BaseID(id)
	if base == FlowingWater || base == FlowingLava {
		return int(id >> fluidLevelShift)
	}
	return 0
}

// IsWater reports whether the block is a water source or flowing water.
func IsWater(id uint16) bool {
	base := BaseID(id)
	return base == Water || base == FlowingWater
}

// IsLava reports whether the block is a lava source or flowing lava.
func IsLava(id uint16) bool {
	base := BaseID(id)
	return base == Lava || base == FlowingLava
}

// IsFire reports whether the block is fire.
func IsFire(id uint16) bool {
	return BaseID(id) == Fire
}
