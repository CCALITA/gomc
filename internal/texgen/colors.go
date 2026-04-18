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
