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
	IsStair         bool
	IsSlab          bool
	IsDoor          bool
	IsTrapdoor      bool
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

	// Wood variants (27-41)
	BirchLog      BlockID = 27
	BirchPlanks   BlockID = 28
	BirchLeaves   BlockID = 29
	SpruceLog     BlockID = 30
	SprucePlanks  BlockID = 31
	SpruceLeaves  BlockID = 32
	JungleLog     BlockID = 33
	JunglePlanks  BlockID = 34
	JungleLeaves  BlockID = 35
	DarkOakLog    BlockID = 36
	DarkOakPlanks BlockID = 37
	DarkOakLeaves BlockID = 38
	AcaciaLog     BlockID = 39
	AcaciaPlanks  BlockID = 40
	AcaciaLeaves  BlockID = 41

	// Wool colors (42-57)
	WhiteWool     BlockID = 42
	OrangeWool    BlockID = 43
	MagentaWool   BlockID = 44
	LightBlueWool BlockID = 45
	YellowWool    BlockID = 46
	LimeWool      BlockID = 47
	PinkWool      BlockID = 48
	GrayWool      BlockID = 49
	LightGrayWool BlockID = 50
	CyanWool      BlockID = 51
	PurpleWool    BlockID = 52
	BlueWool      BlockID = 53
	BrownWool     BlockID = 54
	GreenWool     BlockID = 55
	RedWool       BlockID = 56
	BlackWool     BlockID = 57

	// Functional and nature blocks (58-80)
	IronBlock    BlockID = 58
	GoldBlock    BlockID = 59
	DiamondBlock BlockID = 60
	Bookshelf    BlockID = 61
	TNT          BlockID = 62
	Rail         BlockID = 63
	Ladder       BlockID = 64
	Farmland     BlockID = 65
	WheatCrop    BlockID = 66
	Sugarcane    BlockID = 67
	Cactus       BlockID = 68
	Clay         BlockID = 69
	Bricks       BlockID = 70
	NetherRack   BlockID = 71
	SoulSand     BlockID = 72
	Glowstone    BlockID = 73
	EndStone     BlockID = 74
	Snow         BlockID = 75
	Ice          BlockID = 76
	Sponge       BlockID = 77
	RedstoneOre  BlockID = 78
	LapisOre     BlockID = 79
	EmeraldOre   BlockID = 80

	// Stairs (81-86)
	OakStairs         BlockID = 81
	CobblestoneStairs BlockID = 82
	StoneStairs       BlockID = 83
	BirchStairs       BlockID = 84
	SpruceStairs      BlockID = 85
	SandstoneStairs   BlockID = 86

	// Slabs (87-92)
	OakSlab         BlockID = 87
	CobblestoneSlab BlockID = 88
	StoneSlab       BlockID = 89
	BirchSlab       BlockID = 90
	SpruceSlab      BlockID = 91
	SandstoneSlab   BlockID = 92
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

// IsStairBlock reports whether the block is a stair type.
func IsStairBlock(id uint16) bool {
	return GetProperties(id).IsStair
}

// IsSlabBlock reports whether the block is a slab type.
func IsSlabBlock(id uint16) bool {
	return GetProperties(id).IsSlab
}

// Additional block IDs added in batch 12.
const (
	OakDoor       BlockID = 93
	IronDoor      BlockID = 94
	OakTrapdoor   BlockID = 95
	Bed           BlockID = 96
	MossyCobblestone BlockID = 97
	MobSpawner    BlockID = 98
	OakSign       BlockID = 99
	OakWallSign   BlockID = 100
	OakFence      BlockID = 101
	Anvil         BlockID = 110
	BrewingStand  BlockID = 111

	// Nether blocks (114-115)
	NetherPortal    BlockID = 114
	NetherQuartzOre BlockID = 115
)

const cropStageShift = 8
const cropStageMask = 0x07
const CropMatureStage = 7

func WheatCropStage(stage int) uint16 {
	return WheatCrop | (uint16(stage&cropStageMask) << cropStageShift)
}

func CropGrowthStage(id uint16) int {
	if BaseID(id) != WheatCrop { return 0 }
	return int((id >> cropStageShift) & cropStageMask)
}

func IsCrop(id uint16) bool { return BaseID(id) == WheatCrop }

func IsBed(id uint16) bool { return BaseID(id) == Bed }

func IsDoorBlock(id uint16) bool {
	base := BaseID(id)
	return base == OakDoor || base == IronDoor
}

func IsDoorOrTrapdoor(id uint16) bool {
	base := BaseID(id)
	return base == OakDoor || base == IronDoor || base == OakTrapdoor
}

func IsDoorOpen(id uint16) bool {
	return id&(1<<10) != 0
}

func ToggleDoorOpen(id uint16) uint16 {
	return id ^ (1 << 10)
}

func IsSign(id uint16) bool {
	base := BaseID(id)
	return base == OakSign || base == OakWallSign
}

const doorOpenBit uint16 = 1 << fluidLevelShift

func WithDoorOpen(id uint16, open bool) uint16 {
	if open { return id | doorOpenBit }
	return id &^ doorOpenBit
}
func IsTrapdoorBlock(id uint16) bool { return BaseID(id) == OakTrapdoor }

// IsNetherPortal reports whether the block is a nether portal.
func IsNetherPortal(id uint16) bool { return BaseID(id) == NetherPortal }
