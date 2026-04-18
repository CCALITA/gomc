// Package texgen procedurally generates a block texture atlas at runtime.
package texgen

import (
	"image/color"

	"github.com/fanxiyao/gomc/internal/block"
)

// BlockColor defines the color palette for a single block type.
// Primary is the dominant color. Secondary is used for patterns or side faces.
type BlockColor struct {
	Primary   color.RGBA
	Secondary color.RGBA
	Pattern   PatternType
}

// PatternType selects the noise/pattern applied to a tile.
type PatternType int

const (
	// PatternSolid fills the tile with a single color plus subtle noise.
	PatternSolid PatternType = iota
	// PatternSpeckled draws random speckles in the secondary color over a
	// primary background (used for ores).
	PatternSpeckled
	// PatternMottled blends primary and secondary in a mottled/checkerboard
	// pattern (used for cobblestone, gravel).
	PatternMottled
	// PatternStriped draws vertical stripes alternating primary and secondary
	// (used for oak log bark).
	PatternStriped
	// PatternBorder draws a border of the secondary color around a primary
	// interior (used for crafting table, chest, furnace).
	PatternBorder
	// PatternCross draws a cross/plus shape in the secondary color over
	// primary (used for torch, glass).
	PatternCross
)

// blockColors maps each known block ID to its color definition.
// Air (block ID 0) is intentionally absent: it has no visible texture.
var blockColors = map[block.BlockID]BlockColor{
	block.Stone: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 112, G: 112, B: 112, A: 255},
		Pattern:   PatternSolid,
	},
	block.Dirt: {
		Primary:   color.RGBA{R: 134, G: 96, B: 67, A: 255},
		Secondary: color.RGBA{R: 120, G: 84, B: 58, A: 255},
		Pattern:   PatternSolid,
	},
	block.Grass: {
		Primary:   color.RGBA{R: 95, G: 159, B: 53, A: 255},
		Secondary: color.RGBA{R: 134, G: 96, B: 67, A: 255},
		Pattern:   PatternSolid,
	},
	block.Sand: {
		Primary:   color.RGBA{R: 219, G: 211, B: 160, A: 255},
		Secondary: color.RGBA{R: 210, G: 200, B: 148, A: 255},
		Pattern:   PatternSolid,
	},
	block.Gravel: {
		Primary:   color.RGBA{R: 136, G: 126, B: 126, A: 255},
		Secondary: color.RGBA{R: 110, G: 100, B: 100, A: 255},
		Pattern:   PatternMottled,
	},
	block.OakLog: {
		Primary:   color.RGBA{R: 109, G: 85, B: 51, A: 255},
		Secondary: color.RGBA{R: 177, G: 144, B: 86, A: 255},
		Pattern:   PatternStriped,
	},
	block.OakLeaves: {
		Primary:   color.RGBA{R: 60, G: 140, B: 30, A: 255},
		Secondary: color.RGBA{R: 48, G: 120, B: 24, A: 255},
		Pattern:   PatternSolid,
	},
	block.OakPlanks: {
		Primary:   color.RGBA{R: 177, G: 144, B: 86, A: 255},
		Secondary: color.RGBA{R: 160, G: 130, B: 76, A: 255},
		Pattern:   PatternStriped,
	},
	block.Cobblestone: {
		Primary:   color.RGBA{R: 122, G: 122, B: 122, A: 255},
		Secondary: color.RGBA{R: 90, G: 90, B: 90, A: 255},
		Pattern:   PatternMottled,
	},
	block.Glass: {
		Primary:   color.RGBA{R: 175, G: 210, B: 230, A: 128},
		Secondary: color.RGBA{R: 200, G: 230, B: 245, A: 160},
		Pattern:   PatternCross,
	},
	block.Water: {
		Primary:   color.RGBA{R: 64, G: 64, B: 255, A: 180},
		Secondary: color.RGBA{R: 48, G: 48, B: 220, A: 160},
		Pattern:   PatternSolid,
	},
	block.Lava: {
		Primary:   color.RGBA{R: 207, G: 91, B: 33, A: 255},
		Secondary: color.RGBA{R: 240, G: 140, B: 20, A: 255},
		Pattern:   PatternMottled,
	},
	block.IronOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 200, G: 180, B: 160, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.CoalOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 32, G: 32, B: 32, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.DiamondOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 80, G: 220, B: 230, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.GoldOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 240, G: 200, B: 40, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.Bedrock: {
		Primary:   color.RGBA{R: 48, G: 48, B: 48, A: 255},
		Secondary: color.RGBA{R: 32, G: 32, B: 32, A: 255},
		Pattern:   PatternMottled,
	},
	block.CraftingTable: {
		Primary:   color.RGBA{R: 177, G: 144, B: 86, A: 255},
		Secondary: color.RGBA{R: 120, G: 80, B: 40, A: 255},
		Pattern:   PatternBorder,
	},
	block.Furnace: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 64, G: 64, B: 64, A: 255},
		Pattern:   PatternBorder,
	},
	block.Chest: {
		Primary:   color.RGBA{R: 160, G: 120, B: 60, A: 255},
		Secondary: color.RGBA{R: 100, G: 70, B: 30, A: 255},
		Pattern:   PatternBorder,
	},
	block.Torch: {
		Primary:   color.RGBA{R: 200, G: 180, B: 100, A: 255},
		Secondary: color.RGBA{R: 255, G: 200, B: 50, A: 255},
		Pattern:   PatternCross,
	},
	block.Obsidian: {
		Primary:   color.RGBA{R: 20, G: 18, B: 30, A: 255},
		Secondary: color.RGBA{R: 40, G: 30, B: 50, A: 255},
		Pattern:   PatternSolid,
	},
	block.Sandstone: {
		Primary:   color.RGBA{R: 216, G: 199, B: 143, A: 255},
		Secondary: color.RGBA{R: 190, G: 175, B: 120, A: 255},
		Pattern:   PatternStriped,
	},

	// Wood variants
	block.BirchLog: {
		Primary:   color.RGBA{R: 206, G: 206, B: 201, A: 255},
		Secondary: color.RGBA{R: 180, G: 170, B: 130, A: 255},
		Pattern:   PatternStriped,
	},
	block.BirchPlanks: {
		Primary:   color.RGBA{R: 216, G: 200, B: 150, A: 255},
		Secondary: color.RGBA{R: 196, G: 180, B: 135, A: 255},
		Pattern:   PatternStriped,
	},
	block.BirchLeaves: {
		Primary:   color.RGBA{R: 80, G: 152, B: 48, A: 255},
		Secondary: color.RGBA{R: 64, G: 130, B: 36, A: 255},
		Pattern:   PatternSolid,
	},
	block.SpruceLog: {
		Primary:   color.RGBA{R: 58, G: 37, B: 16, A: 255},
		Secondary: color.RGBA{R: 110, G: 85, B: 51, A: 255},
		Pattern:   PatternStriped,
	},
	block.SprucePlanks: {
		Primary:   color.RGBA{R: 115, G: 85, B: 49, A: 255},
		Secondary: color.RGBA{R: 100, G: 72, B: 40, A: 255},
		Pattern:   PatternStriped,
	},
	block.SpruceLeaves: {
		Primary:   color.RGBA{R: 42, G: 90, B: 42, A: 255},
		Secondary: color.RGBA{R: 32, G: 72, B: 32, A: 255},
		Pattern:   PatternSolid,
	},
	block.JungleLog: {
		Primary:   color.RGBA{R: 86, G: 68, B: 25, A: 255},
		Secondary: color.RGBA{R: 170, G: 137, B: 78, A: 255},
		Pattern:   PatternStriped,
	},
	block.JunglePlanks: {
		Primary:   color.RGBA{R: 170, G: 130, B: 80, A: 255},
		Secondary: color.RGBA{R: 150, G: 112, B: 68, A: 255},
		Pattern:   PatternStriped,
	},
	block.JungleLeaves: {
		Primary:   color.RGBA{R: 36, G: 120, B: 24, A: 255},
		Secondary: color.RGBA{R: 28, G: 100, B: 18, A: 255},
		Pattern:   PatternSolid,
	},
	block.DarkOakLog: {
		Primary:   color.RGBA{R: 60, G: 46, B: 26, A: 255},
		Secondary: color.RGBA{R: 78, G: 56, B: 28, A: 255},
		Pattern:   PatternStriped,
	},
	block.DarkOakPlanks: {
		Primary:   color.RGBA{R: 67, G: 43, B: 20, A: 255},
		Secondary: color.RGBA{R: 54, G: 34, B: 16, A: 255},
		Pattern:   PatternStriped,
	},
	block.DarkOakLeaves: {
		Primary:   color.RGBA{R: 40, G: 110, B: 20, A: 255},
		Secondary: color.RGBA{R: 30, G: 90, B: 14, A: 255},
		Pattern:   PatternSolid,
	},
	block.AcaciaLog: {
		Primary:   color.RGBA{R: 103, G: 96, B: 86, A: 255},
		Secondary: color.RGBA{R: 174, G: 92, B: 40, A: 255},
		Pattern:   PatternStriped,
	},
	block.AcaciaPlanks: {
		Primary:   color.RGBA{R: 168, G: 90, B: 50, A: 255},
		Secondary: color.RGBA{R: 148, G: 76, B: 40, A: 255},
		Pattern:   PatternStriped,
	},
	block.AcaciaLeaves: {
		Primary:   color.RGBA{R: 52, G: 128, B: 28, A: 255},
		Secondary: color.RGBA{R: 40, G: 106, B: 20, A: 255},
		Pattern:   PatternSolid,
	},

	// Wool colors
	block.WhiteWool: {
		Primary:   color.RGBA{R: 234, G: 236, B: 236, A: 255},
		Secondary: color.RGBA{R: 220, G: 222, B: 222, A: 255},
		Pattern:   PatternSolid,
	},
	block.OrangeWool: {
		Primary:   color.RGBA{R: 241, G: 118, B: 20, A: 255},
		Secondary: color.RGBA{R: 220, G: 100, B: 14, A: 255},
		Pattern:   PatternSolid,
	},
	block.MagentaWool: {
		Primary:   color.RGBA{R: 189, G: 68, B: 179, A: 255},
		Secondary: color.RGBA{R: 168, G: 54, B: 158, A: 255},
		Pattern:   PatternSolid,
	},
	block.LightBlueWool: {
		Primary:   color.RGBA{R: 58, G: 175, B: 217, A: 255},
		Secondary: color.RGBA{R: 44, G: 155, B: 196, A: 255},
		Pattern:   PatternSolid,
	},
	block.YellowWool: {
		Primary:   color.RGBA{R: 254, G: 204, B: 15, A: 255},
		Secondary: color.RGBA{R: 236, G: 186, B: 10, A: 255},
		Pattern:   PatternSolid,
	},
	block.LimeWool: {
		Primary:   color.RGBA{R: 112, G: 185, B: 26, A: 255},
		Secondary: color.RGBA{R: 94, G: 164, B: 18, A: 255},
		Pattern:   PatternSolid,
	},
	block.PinkWool: {
		Primary:   color.RGBA{R: 237, G: 141, B: 172, A: 255},
		Secondary: color.RGBA{R: 218, G: 120, B: 152, A: 255},
		Pattern:   PatternSolid,
	},
	block.GrayWool: {
		Primary:   color.RGBA{R: 63, G: 68, B: 72, A: 255},
		Secondary: color.RGBA{R: 50, G: 54, B: 58, A: 255},
		Pattern:   PatternSolid,
	},
	block.LightGrayWool: {
		Primary:   color.RGBA{R: 142, G: 142, B: 135, A: 255},
		Secondary: color.RGBA{R: 126, G: 126, B: 120, A: 255},
		Pattern:   PatternSolid,
	},
	block.CyanWool: {
		Primary:   color.RGBA{R: 21, G: 138, B: 145, A: 255},
		Secondary: color.RGBA{R: 14, G: 118, B: 126, A: 255},
		Pattern:   PatternSolid,
	},
	block.PurpleWool: {
		Primary:   color.RGBA{R: 122, G: 42, B: 173, A: 255},
		Secondary: color.RGBA{R: 104, G: 32, B: 152, A: 255},
		Pattern:   PatternSolid,
	},
	block.BlueWool: {
		Primary:   color.RGBA{R: 53, G: 57, B: 157, A: 255},
		Secondary: color.RGBA{R: 40, G: 44, B: 138, A: 255},
		Pattern:   PatternSolid,
	},
	block.BrownWool: {
		Primary:   color.RGBA{R: 114, G: 72, B: 40, A: 255},
		Secondary: color.RGBA{R: 96, G: 58, B: 30, A: 255},
		Pattern:   PatternSolid,
	},
	block.GreenWool: {
		Primary:   color.RGBA{R: 84, G: 109, B: 27, A: 255},
		Secondary: color.RGBA{R: 70, G: 92, B: 20, A: 255},
		Pattern:   PatternSolid,
	},
	block.RedWool: {
		Primary:   color.RGBA{R: 161, G: 39, B: 35, A: 255},
		Secondary: color.RGBA{R: 140, G: 30, B: 26, A: 255},
		Pattern:   PatternSolid,
	},
	block.BlackWool: {
		Primary:   color.RGBA{R: 20, G: 21, B: 26, A: 255},
		Secondary: color.RGBA{R: 12, G: 12, B: 16, A: 255},
		Pattern:   PatternSolid,
	},

	// Functional and nature blocks
	block.IronBlock: {
		Primary:   color.RGBA{R: 220, G: 220, B: 220, A: 255},
		Secondary: color.RGBA{R: 196, G: 196, B: 196, A: 255},
		Pattern:   PatternSolid,
	},
	block.GoldBlock: {
		Primary:   color.RGBA{R: 246, G: 208, B: 62, A: 255},
		Secondary: color.RGBA{R: 220, G: 180, B: 40, A: 255},
		Pattern:   PatternSolid,
	},
	block.DiamondBlock: {
		Primary:   color.RGBA{R: 98, G: 237, B: 228, A: 255},
		Secondary: color.RGBA{R: 76, G: 210, B: 202, A: 255},
		Pattern:   PatternSolid,
	},
	block.Bookshelf: {
		Primary:   color.RGBA{R: 177, G: 144, B: 86, A: 255},
		Secondary: color.RGBA{R: 90, G: 60, B: 30, A: 255},
		Pattern:   PatternBorder,
	},
	block.TNT: {
		Primary:   color.RGBA{R: 200, G: 30, B: 20, A: 255},
		Secondary: color.RGBA{R: 60, G: 60, B: 60, A: 255},
		Pattern:   PatternStriped,
	},
	block.Rail: {
		Primary:   color.RGBA{R: 130, G: 102, B: 70, A: 255},
		Secondary: color.RGBA{R: 160, G: 160, B: 160, A: 255},
		Pattern:   PatternCross,
	},
	block.Ladder: {
		Primary:   color.RGBA{R: 160, G: 130, B: 76, A: 255},
		Secondary: color.RGBA{R: 120, G: 90, B: 50, A: 255},
		Pattern:   PatternCross,
	},
	block.Farmland: {
		Primary:   color.RGBA{R: 110, G: 72, B: 40, A: 255},
		Secondary: color.RGBA{R: 90, G: 58, B: 30, A: 255},
		Pattern:   PatternMottled,
	},
	block.WheatCrop: {
		Primary:   color.RGBA{R: 180, G: 170, B: 40, A: 255},
		Secondary: color.RGBA{R: 100, G: 140, B: 20, A: 255},
		Pattern:   PatternCross,
	},
	block.Sugarcane: {
		Primary:   color.RGBA{R: 100, G: 180, B: 60, A: 255},
		Secondary: color.RGBA{R: 80, G: 160, B: 44, A: 255},
		Pattern:   PatternCross,
	},
	block.Cactus: {
		Primary:   color.RGBA{R: 14, G: 112, B: 20, A: 255},
		Secondary: color.RGBA{R: 8, G: 90, B: 14, A: 255},
		Pattern:   PatternSolid,
	},
	block.Clay: {
		Primary:   color.RGBA{R: 160, G: 166, B: 180, A: 255},
		Secondary: color.RGBA{R: 140, G: 146, B: 160, A: 255},
		Pattern:   PatternSolid,
	},
	block.Bricks: {
		Primary:   color.RGBA{R: 150, G: 74, B: 56, A: 255},
		Secondary: color.RGBA{R: 190, G: 186, B: 170, A: 255},
		Pattern:   PatternMottled,
	},
	block.NetherRack: {
		Primary:   color.RGBA{R: 111, G: 54, B: 53, A: 255},
		Secondary: color.RGBA{R: 90, G: 40, B: 40, A: 255},
		Pattern:   PatternMottled,
	},
	block.SoulSand: {
		Primary:   color.RGBA{R: 81, G: 62, B: 50, A: 255},
		Secondary: color.RGBA{R: 64, G: 48, B: 38, A: 255},
		Pattern:   PatternSolid,
	},
	block.Glowstone: {
		Primary:   color.RGBA{R: 249, G: 212, B: 113, A: 255},
		Secondary: color.RGBA{R: 230, G: 190, B: 80, A: 255},
		Pattern:   PatternMottled,
	},
	block.EndStone: {
		Primary:   color.RGBA{R: 219, G: 223, B: 158, A: 255},
		Secondary: color.RGBA{R: 200, G: 204, B: 140, A: 255},
		Pattern:   PatternSolid,
	},
	block.Snow: {
		Primary:   color.RGBA{R: 249, G: 254, B: 254, A: 255},
		Secondary: color.RGBA{R: 230, G: 236, B: 236, A: 255},
		Pattern:   PatternSolid,
	},
	block.Ice: {
		Primary:   color.RGBA{R: 145, G: 183, B: 253, A: 200},
		Secondary: color.RGBA{R: 120, G: 160, B: 230, A: 180},
		Pattern:   PatternSolid,
	},
	block.Sponge: {
		Primary:   color.RGBA{R: 196, G: 192, B: 74, A: 255},
		Secondary: color.RGBA{R: 176, G: 172, B: 58, A: 255},
		Pattern:   PatternMottled,
	},
	block.RedstoneOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 200, G: 30, B: 20, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.LapisOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 38, G: 68, B: 160, A: 255},
		Pattern:   PatternSpeckled,
	},
	block.EmeraldOre: {
		Primary:   color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Secondary: color.RGBA{R: 40, G: 200, B: 60, A: 255},
		Pattern:   PatternSpeckled,
	},
}

// GetBlockColor returns the color definition for the given block ID.
// If the block ID is not mapped, a fallback magenta tile is returned so
// unmapped blocks are visually obvious.
func GetBlockColor(id block.BlockID) BlockColor {
	if bc, ok := blockColors[id]; ok {
		return bc
	}
	return BlockColor{
		Primary:   color.RGBA{R: 255, G: 0, B: 255, A: 255},
		Secondary: color.RGBA{R: 200, G: 0, B: 200, A: 255},
		Pattern:   PatternSolid,
	}
}

// AllBlockIDs returns a slice of every block ID that has a color definition.
// The order is non-deterministic (map iteration order).
func AllBlockIDs() []block.BlockID {
	ids := make([]block.BlockID, 0, len(blockColors))
	for id := range blockColors {
		ids = append(ids, id)
	}
	return ids
}
