package texgen

import (
	"image/color"
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAtlas_Dimensions(t *testing.T) {
	atlas := GenerateAtlas()
	assert.NotNil(t, atlas)
	assert.Equal(t, AtlasSize, atlas.Bounds().Dx(), "atlas width should be %d", AtlasSize)
	assert.Equal(t, AtlasSize, atlas.Bounds().Dy(), "atlas height should be %d", AtlasSize)
}

func TestGenerateAtlas_NonNil(t *testing.T) {
	atlas := GenerateAtlas()
	assert.NotNil(t, atlas, "GenerateAtlas must return a non-nil image")
	assert.NotEmpty(t, atlas.Pix, "atlas pixel data should not be empty")
}

func TestGenerateAtlas_StoneTileIsGray(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Stone)
	avg := averageTileColor(atlas, col, row)

	// Stone should be gray: R, G, B roughly equal and in the range [100, 160]
	assert.InDelta(t, 128, int(avg.R), 30, "stone R should be gray (~128)")
	assert.InDelta(t, 128, int(avg.G), 30, "stone G should be gray (~128)")
	assert.InDelta(t, 128, int(avg.B), 30, "stone B should be gray (~128)")
	assert.Equal(t, uint8(255), avg.A, "stone should be fully opaque")
}

func TestGenerateAtlas_GrassTileIsGreen(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Grass)
	avg := averageTileColor(atlas, col, row)

	// Grass should be green: G channel dominant
	assert.Greater(t, int(avg.G), int(avg.R), "grass G should exceed R")
	assert.Greater(t, int(avg.G), int(avg.B), "grass G should exceed B")
}

func TestGenerateAtlas_WaterTileIsBlue(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Water)
	avg := averageTileColor(atlas, col, row)

	// Water should be blue: B channel dominant
	assert.Greater(t, int(avg.B), int(avg.R), "water B should exceed R")
	assert.Greater(t, int(avg.B), int(avg.G), "water B should exceed G")
}

func TestGenerateAtlas_WaterTileIsSemiTransparent(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Water)
	avg := averageTileColor(atlas, col, row)

	// Water alpha should be semi-transparent (not 255 and not 0)
	assert.Less(t, avg.A, uint8(255), "water should be semi-transparent")
	assert.Greater(t, avg.A, uint8(0), "water should not be fully transparent")
}

func TestGenerateAtlas_AllBlockIDsHaveUniqueTiles(t *testing.T) {
	atlas := GenerateAtlas()

	// Collect average colors for all block IDs with color definitions.
	type tileInfo struct {
		id  block.BlockID
		avg color.RGBA
	}

	ids := AllBlockIDs()
	tiles := make([]tileInfo, 0, len(ids))
	for _, id := range ids {
		col, row := TilePosition(id)
		avg := averageTileColor(atlas, col, row)
		tiles = append(tiles, tileInfo{id: id, avg: avg})
	}

	assert.GreaterOrEqual(t, len(tiles), 23, "should have at least 23 non-air block tiles")

	// Verify no two blocks produce the exact same average color tile.
	seen := make(map[color.RGBA]block.BlockID)
	for _, ti := range tiles {
		if prevID, dup := seen[ti.avg]; dup {
			t.Errorf("block %d and block %d have identical average tile color %v",
				prevID, ti.id, ti.avg)
		}
		seen[ti.avg] = ti.id
	}
}

func TestGenerateAtlas_AirTileIsTransparent(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Air)
	avg := averageTileColor(atlas, col, row)

	// Air tile should be fully transparent (all zeros).
	assert.Equal(t, uint8(0), avg.R, "air R should be 0")
	assert.Equal(t, uint8(0), avg.G, "air G should be 0")
	assert.Equal(t, uint8(0), avg.B, "air B should be 0")
	assert.Equal(t, uint8(0), avg.A, "air A should be 0")
}

func TestGenerateAtlas_DirtTileIsBrown(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Dirt)
	avg := averageTileColor(atlas, col, row)

	// Dirt should be brown: R > G > B
	assert.Greater(t, int(avg.R), int(avg.G), "dirt R should exceed G")
	assert.Greater(t, int(avg.G), int(avg.B), "dirt G should exceed B")
}

func TestGenerateAtlas_LavaTileIsOrange(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Lava)
	avg := averageTileColor(atlas, col, row)

	// Lava should be orange: R > G > B
	assert.Greater(t, int(avg.R), int(avg.G), "lava R should exceed G")
	assert.Greater(t, int(avg.G), int(avg.B), "lava G should exceed B")
}

func TestGenerateAtlas_BedrockIsDarkGray(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.Bedrock)
	avg := averageTileColor(atlas, col, row)

	// Bedrock should be dark gray: all channels low and roughly equal
	assert.Less(t, int(avg.R), 80, "bedrock R should be dark")
	assert.Less(t, int(avg.G), 80, "bedrock G should be dark")
	assert.Less(t, int(avg.B), 80, "bedrock B should be dark")
}

func TestTilePosition(t *testing.T) {
	tests := []struct {
		id      block.BlockID
		wantCol int
		wantRow int
	}{
		{block.Air, 0, 0},
		{block.Stone, 1, 0},
		{block.Dirt, 2, 0},
		{block.Water, 11, 0},
		{block.Bedrock, 1, 1},
		{block.Sandstone, 7, 1},
	}

	for _, tc := range tests {
		col, row := TilePosition(tc.id)
		assert.Equal(t, tc.wantCol, col, "block %d col", tc.id)
		assert.Equal(t, tc.wantRow, row, "block %d row", tc.id)
	}
}

func TestAtlasToRGBA_Length(t *testing.T) {
	atlas := GenerateAtlas()
	data := AtlasToRGBA(atlas)
	expectedLen := AtlasSize * AtlasSize * 4
	assert.Equal(t, expectedLen, len(data), "RGBA data length should be %d", expectedLen)
}

func TestGetBlockColor_KnownBlock(t *testing.T) {
	bc := GetBlockColor(block.Stone)
	assert.Equal(t, uint8(128), bc.Primary.R)
	assert.Equal(t, uint8(128), bc.Primary.G)
	assert.Equal(t, uint8(128), bc.Primary.B)
}

func TestGetBlockColor_UnknownBlock(t *testing.T) {
	// Block ID 255 is not defined; should return magenta fallback.
	bc := GetBlockColor(255)
	assert.Equal(t, uint8(255), bc.Primary.R)
	assert.Equal(t, uint8(0), bc.Primary.G)
	assert.Equal(t, uint8(255), bc.Primary.B)
}

func TestAllBlockIDs_ReturnsNonEmpty(t *testing.T) {
	ids := AllBlockIDs()
	assert.GreaterOrEqual(t, len(ids), 23, "should have at least 23 color definitions")
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		input int16
		want  uint8
	}{
		{-10, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}
	for _, tc := range tests {
		got := clampByte(tc.input)
		assert.Equal(t, tc.want, got, "clampByte(%d)", tc.input)
	}
}

func TestTileNoise_Deterministic(t *testing.T) {
	// Same inputs should always produce the same output.
	a := tileNoise(3, 7)
	b := tileNoise(3, 7)
	assert.Equal(t, a, b, "tileNoise should be deterministic")
}

func TestTileNoise_Range(t *testing.T) {
	for py := 0; py < TileSize; py++ {
		for px := 0; px < TileSize; px++ {
			n := tileNoise(px, py)
			assert.GreaterOrEqual(t, n, 0, "noise should be >= 0")
			assert.LessOrEqual(t, n, 15, "noise should be <= 15")
		}
	}
}

func TestApplyNoise_PreservesAlpha(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 200}
	result := applyNoise(c, 8)
	assert.Equal(t, uint8(200), result.A, "alpha should be preserved")
}

func TestIsSpeckle_Deterministic(t *testing.T) {
	a := isSpeckle(5, 10)
	b := isSpeckle(5, 10)
	assert.Equal(t, a, b, "isSpeckle should be deterministic")
}

func TestGenerateAtlas_DiamondOreHasSpeckles(t *testing.T) {
	atlas := GenerateAtlas()
	col, row := TilePosition(block.DiamondOre)
	ox := col * TileSize
	oy := row * TileSize

	// Diamond ore should have both stone-gray and diamond-blue pixels.
	hasBlueish := false
	hasGrayish := false
	for py := 0; py < TileSize; py++ {
		for px := 0; px < TileSize; px++ {
			r, g, b, _ := atlas.At(ox+px, oy+py).RGBA()
			// Normalize to 0-255
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			if b8 > 180 && b8 > r8 {
				hasBlueish = true
			}
			if r8 > 100 && g8 > 100 && b8 > 100 && r8 < 170 {
				hasGrayish = true
			}
		}
	}
	assert.True(t, hasBlueish, "diamond ore should have blue speckles")
	assert.True(t, hasGrayish, "diamond ore should have gray base")
}

// averageTileColor computes the average RGBA of the tile at the given grid
// position. Returns the averaged color with integer-rounded channels.
func averageTileColor(atlas interface {
	At(x, y int) color.Color
}, col, row int) color.RGBA {
	ox := col * TileSize
	oy := row * TileSize

	var rSum, gSum, bSum, aSum int
	count := TileSize * TileSize

	for py := 0; py < TileSize; py++ {
		for px := 0; px < TileSize; px++ {
			r, g, b, a := atlas.At(ox+px, oy+py).RGBA()
			rSum += int(r >> 8)
			gSum += int(g >> 8)
			bSum += int(b >> 8)
			aSum += int(a >> 8)
		}
	}

	return color.RGBA{
		R: uint8(rSum / count),
		G: uint8(gSum / count),
		B: uint8(bSum / count),
		A: uint8(aSum / count),
	}
}
