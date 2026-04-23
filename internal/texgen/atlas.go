package texgen

import (
	"image"
	"image/color"

	"github.com/fanxiyao/gomc/internal/block"
)

const (
	// TileSize is the width and height of each tile in pixels.
	TileSize = 16
	// GridSize is the number of tiles along each axis of the atlas.
	GridSize = 16
	// AtlasSize is the total width (and height) of the atlas in pixels.
	AtlasSize = TileSize * GridSize // 256
)

// GenerateAtlas creates a 256x256 RGBA image containing a 16x16 grid of
// procedurally generated block textures. Block IDs map to tile positions
// with col = blockID % 16 and row = blockID / 16.
func GenerateAtlas() *image.RGBA {
	atlas := image.NewRGBA(image.Rect(0, 0, AtlasSize, AtlasSize))

	// Generate a tile for every block ID that has a color definition.
	for _, id := range AllBlockIDs() {
		col, row := TilePosition(id)
		drawTile(atlas, col, row, id)
	}

	return atlas
}

// AtlasToRGBA extracts the raw RGBA byte slice from the atlas image.
// The returned slice has length AtlasSize * AtlasSize * 4.
func AtlasToRGBA(atlas *image.RGBA) []byte {
	return atlas.Pix
}

// TilePosition returns the (col, row) position in the atlas grid for the
// given block ID. col = id % 16, row = id / 16.
func TilePosition(id block.BlockID) (col, row int) {
	col = int(id) % GridSize
	row = int(id) / GridSize
	return col, row
}

// drawTile renders a single 16x16 tile into the atlas at the given grid
// position, using the block's color palette and pattern.
func drawTile(atlas *image.RGBA, col, row int, id block.BlockID) {
	bc := GetBlockColor(id)
	originX := col * TileSize
	originY := row * TileSize

	switch bc.Pattern {
	case PatternSpeckled:
		drawSpeckled(atlas, originX, originY, bc)
	case PatternMottled:
		drawMottled(atlas, originX, originY, bc)
	case PatternStriped:
		drawStriped(atlas, originX, originY, bc)
	case PatternBorder:
		drawBorder(atlas, originX, originY, bc)
	case PatternCross:
		drawCross(atlas, originX, originY, bc)
	default:
		drawSolid(atlas, originX, originY, bc)
	}
}

// pixelSelector returns true when the pixel at (px, py) should use the
// secondary color instead of the primary.
type pixelSelector func(px, py, noise int) bool

// drawTilePixels iterates over every pixel in a tile, selecting primary or
// secondary color via the useSecondary callback, applying noise, and writing
// the result. All pattern-specific draw functions delegate to this.
func drawTilePixels(atlas *image.RGBA, ox, oy int, bc BlockColor, useSecondary pixelSelector) {
	for py := 0; py < TileSize; py++ {
		for px := 0; px < TileSize; px++ {
			noise := tileNoise(px, py)
			base := bc.Primary
			if useSecondary(px, py, noise) {
				base = bc.Secondary
			}
			atlas.SetRGBA(ox+px, oy+py, applyNoise(base, noise))
		}
	}
}

// drawSolid fills the tile with the primary color plus subtle per-pixel noise.
func drawSolid(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	drawTilePixels(atlas, ox, oy, bc, func(_, _, _ int) bool { return false })
}

// drawSpeckled draws a stone-like base with colored speckles for ore blocks.
func drawSpeckled(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	drawTilePixels(atlas, ox, oy, bc, func(px, py, _ int) bool { return isSpeckle(px, py) })
}

// drawMottled creates a mottled pattern by alternating between primary and
// secondary based on a checkerboard with noise.
func drawMottled(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	drawTilePixels(atlas, ox, oy, bc, func(px, py, noise int) bool {
		usePrimary := ((px/3)+(py/3))%2 == 0
		if noise > 4 {
			usePrimary = !usePrimary
		}
		return !usePrimary
	})
}

// drawStriped draws vertical stripes alternating primary and secondary.
func drawStriped(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	drawTilePixels(atlas, ox, oy, bc, func(px, _, _ int) bool { return (px/4)%2 != 0 })
}

// drawBorder draws a 2-pixel border of the secondary color around a primary
// interior.
func drawBorder(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	const borderWidth = 2
	drawTilePixels(atlas, ox, oy, bc, func(px, py, _ int) bool {
		return px < borderWidth || px >= TileSize-borderWidth ||
			py < borderWidth || py >= TileSize-borderWidth
	})
}

// drawCross draws a plus/cross shape in the secondary color over a primary
// background.
func drawCross(atlas *image.RGBA, ox, oy int, bc BlockColor) {
	drawTilePixels(atlas, ox, oy, bc, func(px, py, _ int) bool {
		return (px >= 6 && px <= 9) || (py >= 6 && py <= 9)
	})
}

// tileNoise returns a deterministic pseudo-random noise value in [0, 15]
// for the given pixel position within a tile. This is a simple hash-based
// approach that produces repeatable, tileable noise.
func tileNoise(px, py int) int {
	h := uint32(px*7 + py*13)
	h ^= h << 5
	h = h * 2654435761 // Knuth multiplicative hash constant
	return int(h & 0x0F)
}

// isSpeckle returns true if the given pixel position should show a speckle.
// Uses a deterministic pattern that places roughly 15-20% of pixels as
// speckles.
func isSpeckle(px, py int) bool {
	h := uint32(px*31 + py*37)
	h ^= h << 7
	h = h * 2654435761
	return (h % 6) == 0
}

// applyNoise adjusts a color's RGB channels by a small noise offset for
// visual variety. The alpha channel is preserved.
func applyNoise(c color.RGBA, noise int) color.RGBA {
	offset := int16(noise) - 8
	scale := int16(3)
	return color.RGBA{
		R: clampByte(int16(c.R) + offset*scale),
		G: clampByte(int16(c.G) + offset*scale),
		B: clampByte(int16(c.B) + offset*scale),
		A: c.A,
	}
}

// clampByte clamps an int16 to the [0, 255] range and returns it as a byte.
func clampByte(v int16) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
