package block

import (
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// properties maps every known BlockID to its BlockProperties.
var properties = map[BlockID]BlockProperties{
	Air: {
		Name: "air", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 0, LightFilter: 0,
	},
	Stone: {
		Name: "stone", Solid: true, Transparent: false,
		Hardness: 1.5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	Dirt: {
		Name: "dirt", Solid: true, Transparent: false,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 15,
	},
	Grass: {
		Name: "grass", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	Sand: {
		Name: "sand", Solid: true, Transparent: false,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 15,
	},
	Gravel: {
		Name: "gravel", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	OakLog: {
		Name: "oak_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	OakLeaves: {
		Name: "oak_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	OakPlanks: {
		Name: "oak_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	Cobblestone: {
		Name: "cobblestone", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	Glass: {
		Name: "glass", Solid: true, Transparent: true,
		Hardness: 0.3, BlastResistance: 0.3,
		LightEmission: 0, LightFilter: 0,
	},
	Water: {
		Name: "water", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 0, LightFilter: 2,
	},
	Lava: {
		Name: "lava", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 15, LightFilter: 0,
	},
	IronOre: {
		Name: "iron_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	CoalOre: {
		Name: "coal_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	DiamondOre: {
		Name: "diamond_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	GoldOre: {
		Name: "gold_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	Bedrock: {
		Name: "bedrock", Solid: true, Transparent: false,
		Hardness: -1, BlastResistance: 3600000,
		LightEmission: 0, LightFilter: 15,
	},
	CraftingTable: {
		Name: "crafting_table", Solid: true, Transparent: false,
		Hardness: 2.5, BlastResistance: 2.5,
		LightEmission: 0, LightFilter: 15,
	},
	Furnace: {
		Name: "furnace", Solid: true, Transparent: false,
		Hardness: 3.5, BlastResistance: 3.5,
		LightEmission: 0, LightFilter: 15,
	},
	Chest: {
		Name: "chest", Solid: true, Transparent: false,
		Hardness: 2.5, BlastResistance: 2.5,
		LightEmission: 0, LightFilter: 15,
	},
	Torch: {
		Name: "torch", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 14, LightFilter: 0,
	},
	Obsidian: {
		Name: "obsidian", Solid: true, Transparent: false,
		Hardness: 50, BlastResistance: 1200,
		LightEmission: 0, LightFilter: 15,
	},
	Sandstone: {
		Name: "sandstone", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	FlowingWater: {
		Name: "flowing_water", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 0, LightFilter: 2,
	},
	FlowingLava: {
		Name: "flowing_lava", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 15, LightFilter: 0,
	},
	Fire: {
		Name: "fire", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 15, LightFilter: 0,
	},

	// Wood variants
	BirchLog: {
		Name: "birch_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	BirchPlanks: {
		Name: "birch_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	BirchLeaves: {
		Name: "birch_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	SpruceLog: {
		Name: "spruce_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	SprucePlanks: {
		Name: "spruce_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	SpruceLeaves: {
		Name: "spruce_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	JungleLog: {
		Name: "jungle_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	JunglePlanks: {
		Name: "jungle_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	JungleLeaves: {
		Name: "jungle_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	DarkOakLog: {
		Name: "dark_oak_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	DarkOakPlanks: {
		Name: "dark_oak_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	DarkOakLeaves: {
		Name: "dark_oak_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	AcaciaLog: {
		Name: "acacia_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	AcaciaPlanks: {
		Name: "acacia_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	AcaciaLeaves: {
		Name: "acacia_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},

	// Wool colors
	WhiteWool: {
		Name: "white_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	OrangeWool: {
		Name: "orange_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	MagentaWool: {
		Name: "magenta_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	LightBlueWool: {
		Name: "light_blue_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	YellowWool: {
		Name: "yellow_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	LimeWool: {
		Name: "lime_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	PinkWool: {
		Name: "pink_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	GrayWool: {
		Name: "gray_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	LightGrayWool: {
		Name: "light_gray_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	CyanWool: {
		Name: "cyan_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	PurpleWool: {
		Name: "purple_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	BlueWool: {
		Name: "blue_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	BrownWool: {
		Name: "brown_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	GreenWool: {
		Name: "green_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	RedWool: {
		Name: "red_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	BlackWool: {
		Name: "black_wool", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},

	// Functional and nature blocks
	IronBlock: {
		Name: "iron_block", Solid: true, Transparent: false,
		Hardness: 5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	GoldBlock: {
		Name: "gold_block", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	DiamondBlock: {
		Name: "diamond_block", Solid: true, Transparent: false,
		Hardness: 5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	Bookshelf: {
		Name: "bookshelf", Solid: true, Transparent: false,
		Hardness: 1.5, BlastResistance: 1.5,
		LightEmission: 0, LightFilter: 15,
	},
	TNT: {
		Name: "tnt", Solid: true, Transparent: false,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 0, LightFilter: 15,
	},
	Rail: {
		Name: "rail", Solid: false, Transparent: true,
		Hardness: 0.7, BlastResistance: 0.7,
		LightEmission: 0, LightFilter: 0,
	},
	Ladder: {
		Name: "ladder", Solid: false, Transparent: true,
		Hardness: 0.4, BlastResistance: 0.4,
		LightEmission: 0, LightFilter: 0,
	},
	Farmland: {
		Name: "farmland", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	WheatCrop: {
		Name: "wheat_crop", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 0, LightFilter: 0,
	},
	Sugarcane: {
		Name: "sugarcane", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 0, LightFilter: 0,
	},
	Cactus: {
		Name: "cactus", Solid: true, Transparent: false,
		Hardness: 0.4, BlastResistance: 0.4,
		LightEmission: 0, LightFilter: 15,
	},
	Clay: {
		Name: "clay", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	Bricks: {
		Name: "bricks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	NetherRack: {
		Name: "netherrack", Solid: true, Transparent: false,
		Hardness: 0.4, BlastResistance: 0.4,
		LightEmission: 0, LightFilter: 15,
	},
	SoulSand: {
		Name: "soul_sand", Solid: true, Transparent: false,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 15,
	},
	Glowstone: {
		Name: "glowstone", Solid: true, Transparent: false,
		Hardness: 0.3, BlastResistance: 0.3,
		LightEmission: 15, LightFilter: 15,
	},
	EndStone: {
		Name: "end_stone", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 9,
		LightEmission: 0, LightFilter: 15,
	},
	Snow: {
		Name: "snow", Solid: true, Transparent: false,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 15,
	},
	Ice: {
		Name: "ice", Solid: true, Transparent: true,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 1,
	},
	Sponge: {
		Name: "sponge", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	RedstoneOre: {
		Name: "redstone_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	LapisOre: {
		Name: "lapis_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	EmeraldOre: {
		Name: "emerald_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},

	// Stairs — hardness matches base material
	OakStairs: {
		Name: "oak_stairs", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},
	CobblestoneStairs: {
		Name: "cobblestone_stairs", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},
	StoneStairs: {
		Name: "stone_stairs", Solid: true, Transparent: false,
		Hardness: 1.5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},
	BirchStairs: {
		Name: "birch_stairs", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},
	SpruceStairs: {
		Name: "spruce_stairs", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},
	SandstoneStairs: {
		Name: "sandstone_stairs", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
		IsStair: true,
	},

	// Slabs — hardness matches base material
	OakSlab: {
		Name: "oak_slab", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
	CobblestoneSlab: {
		Name: "cobblestone_slab", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
	StoneSlab: {
		Name: "stone_slab", Solid: true, Transparent: false,
		Hardness: 1.5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
	BirchSlab: {
		Name: "birch_slab", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
	SpruceSlab: {
		Name: "spruce_slab", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
	SandstoneSlab: {
		Name: "sandstone_slab", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
		IsSlab: true,
	},
}

// GetBlockAABBs returns the collision AABBs for a block based on its properties
// and state. For stair blocks, it returns two AABBs based on the orientation.
// For slab blocks, it returns a half-height AABB. For regular solid blocks,
// it returns a single full-block AABB. For non-solid blocks, it returns nil.
func GetBlockAABBs(id BlockID, pos mcmath.BlockPos, orientation Orientation) []mcmath.AABB {
	props := GetProperties(id)
	if !props.Solid {
		return nil
	}
	if props.IsStair {
		return mcmath.StairAABBs(pos, int(orientation))
	}
	if props.IsSlab {
		// OrientUp means top slab; everything else is bottom slab.
		top := orientation == OrientUp
		return []mcmath.AABB{mcmath.SlabAABB(pos, top)}
	}
	return []mcmath.AABB{mcmath.BlockAABB(pos)}
}

// GetProperties returns the BlockProperties for the given block ID.
// For flowing fluid blocks with encoded levels, the base ID is used for lookup.
// If the ID is unknown, a zero-value BlockProperties is returned.
func GetProperties(id BlockID) BlockProperties {
	if p, ok := properties[id]; ok {
		return p
	}
	// Only fall back to base ID for blocks that encode a fluid level.
	base := BaseID(id)
	if base == FlowingWater || base == FlowingLava {
		return properties[base]
	}
	return BlockProperties{}
}

// IsSolid reports whether the block with the given ID is solid.
func IsSolid(id BlockID) bool {
	return GetProperties(id).Solid
}

// IsTransparent reports whether the block with the given ID is transparent.
func IsTransparent(id BlockID) bool {
	return GetProperties(id).Transparent
}

// IsLightSource reports whether the block with the given ID emits light.
func IsLightSource(id BlockID) bool {
	return GetProperties(id).LightEmission > 0
}
