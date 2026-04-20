package block

import (
	"github.com/fanxiyao/gomc/internal/registry"
)

// Blocks is the global block registry mapping block names to BlockProperties.
var Blocks = registry.New[BlockProperties]()

// InitRegistry registers every known block type in the global Blocks
// registry and freezes it. It must be called exactly once during
// application startup. It panics if any registration fails.
func InitRegistry() {
	for _, id := range blockOrder {
		props := properties[id]
		assigned, err := Blocks.Register(props.Name, props)
		if err != nil {
			panic("block.InitRegistry: " + err.Error())
		}
		if assigned != id {
			panic("block.InitRegistry: unexpected ID assignment for " + props.Name)
		}
	}
	Blocks.Freeze()
}

// blockOrder defines the registration order so that each block receives
// the expected BlockID (sequential from 0).
var blockOrder = []BlockID{
	Air,
	Stone,
	Dirt,
	Grass,
	Sand,
	Gravel,
	OakLog,
	OakLeaves,
	OakPlanks,
	Cobblestone,
	Glass,
	Water,
	Lava,
	IronOre,
	CoalOre,
	DiamondOre,
	GoldOre,
	Bedrock,
	CraftingTable,
	Furnace,
	Chest,
	Torch,
	Obsidian,
	Sandstone,
	FlowingWater,
	FlowingLava,
	Fire,
	BirchLog,
	BirchPlanks,
	BirchLeaves,
	SpruceLog,
	SprucePlanks,
	SpruceLeaves,
	JungleLog,
	JunglePlanks,
	JungleLeaves,
	DarkOakLog,
	DarkOakPlanks,
	DarkOakLeaves,
	AcaciaLog,
	AcaciaPlanks,
	AcaciaLeaves,
	WhiteWool,
	OrangeWool,
	MagentaWool,
	LightBlueWool,
	YellowWool,
	LimeWool,
	PinkWool,
	GrayWool,
	LightGrayWool,
	CyanWool,
	PurpleWool,
	BlueWool,
	BrownWool,
	GreenWool,
	RedWool,
	BlackWool,
	IronBlock,
	GoldBlock,
	DiamondBlock,
	Bookshelf,
	TNT,
	Rail,
	Ladder,
	Farmland,
	WheatCrop,
	Sugarcane,
	Cactus,
	Clay,
	Bricks,
	NetherRack,
	SoulSand,
	Glowstone,
	EndStone,
	Snow,
	Ice,
	Sponge,
	RedstoneOre,
	LapisOre,
	EmeraldOre,
	OakStairs,
	CobblestoneStairs,
	StoneStairs,
	BirchStairs,
	SpruceStairs,
	SandstoneStairs,
	OakSlab,
	CobblestoneSlab,
	StoneSlab,
	BirchSlab,
	SpruceSlab,
	SandstoneSlab,
}
