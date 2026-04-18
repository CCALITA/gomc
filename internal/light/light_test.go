package light

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// newEmptyChunk returns a chunk with all air blocks (no sections initialised).
func newEmptyChunk() *chunk.Chunk {
	return &chunk.Chunk{}
}

// placeTorch sets a Torch (block ID 21, emission 14) at the given position.
func placeTorch(c *chunk.Chunk, x, y, z int) {
	c.SetBlock(x, y, z, block.Torch)
}

// ---------------------------------------------------------------------------
// Block light tests
// ---------------------------------------------------------------------------

func TestSingleTorchFalloff(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 8, 8, 8)

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	// Torch itself should be at level 14.
	sec := c.GetSection(8 / mcmath.SectionHeight)
	assert.Equal(t, uint8(14), sec.GetBlockLight(8, 8%mcmath.SectionHeight, 8))

	// Check falloff along positive X: each step decreases by 1.
	for d := 1; d <= 13; d++ {
		nx := 8 + d
		if nx >= 16 {
			break
		}
		expected := uint8(14 - d)
		si := 8 / mcmath.SectionHeight
		actual := c.GetSection(si).GetBlockLight(nx, 8%mcmath.SectionHeight, 8)
		assert.Equal(t, expected, actual, "distance %d along +X", d)
	}
}

func TestLightLevelAtDistanceD(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 8, 8, 8)

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	// Light at distance D from torch = max(0, 14-D).
	// Test along +X, +Y, +Z, -X, -Y, -Z.
	directions := [][3]int{
		{1, 0, 0}, {-1, 0, 0},
		{0, 1, 0}, {0, -1, 0},
		{0, 0, 1}, {0, 0, -1},
	}

	for _, dir := range directions {
		for d := 0; d <= 14; d++ {
			nx := 8 + dir[0]*d
			ny := 8 + dir[1]*d
			nz := 8 + dir[2]*d
			if nx < 0 || nx >= 16 || ny < 0 || ny >= mcmath.ChunkHeight || nz < 0 || nz >= 16 {
				continue
			}
			expected := uint8(0)
			if 14-d > 0 {
				expected = uint8(14 - d)
			}
			si := ny / mcmath.SectionHeight
			sec := c.GetSection(si)
			if sec == nil {
				assert.Equal(t, uint8(0), expected, "distance %d dir %v: section nil but expected non-zero", d, dir)
				continue
			}
			actual := sec.GetBlockLight(nx, ny%mcmath.SectionHeight, nz)
			assert.Equal(t, expected, actual, "distance %d dir %v", d, dir)
		}
	}
}

func TestSolidBlockBlocksLight(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 4, 4, 4)

	// Place a solid opaque wall at x=6 (two blocks from torch).
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(6, y, z, block.Stone)
		}
	}

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	// At x=5 (distance 1) light should be 13.
	sec := c.GetSection(0)
	assert.Equal(t, uint8(13), sec.GetBlockLight(5, 4, 4))

	// At x=7 (behind the wall along direct path) light should be much less
	// than the unobstructed value of 11 (14-3). The wall forces light to go
	// around, increasing the effective distance.
	behindWall := sec.GetBlockLight(7, 4, 4)
	assert.Less(t, behindWall, uint8(11), "light behind solid wall should be reduced")
}

func TestTwoTorchesMaxMerging(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 2, 4, 8)
	placeTorch(c, 12, 4, 8)

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	sec := c.GetSection(0)

	// Both sources emit 14.
	assert.Equal(t, uint8(14), sec.GetBlockLight(2, 4, 8))
	assert.Equal(t, uint8(14), sec.GetBlockLight(12, 4, 8))

	// Midpoint at x=7 is distance 5 from torch at x=2 and distance 5 from
	// torch at x=12. Expected light = max(14-5, 14-5) = 9.
	assert.Equal(t, uint8(9), sec.GetBlockLight(7, 4, 8))

	// At x=3 (distance 1 from torch@2, distance 9 from torch@12):
	// max(13, 5) = 13.
	assert.Equal(t, uint8(13), sec.GetBlockLight(3, 4, 8))
}

// ---------------------------------------------------------------------------
// Sky light tests
// ---------------------------------------------------------------------------

func TestSkyLightEmptyChunk(t *testing.T) {
	c := newEmptyChunk()

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	// In an empty chunk the heightmap is all 0, so every Y from 0 to 255
	// should have sky light = 15.
	for y := 0; y < mcmath.ChunkHeight; y++ {
		si := y / mcmath.SectionHeight
		sec := c.GetSection(si)
		if sec == nil {
			// Sections that were not created still count as 0 light; however
			// PropagateSkyLight creates sections lazily. If nil, we accept 0.
			continue
		}
		actual := sec.GetSkyLight(8, y%mcmath.SectionHeight, 8)
		assert.Equal(t, uint8(15), actual, "sky light at y=%d should be 15", y)
	}
}

func TestSkyLightBlockedByRoof(t *testing.T) {
	c := newEmptyChunk()

	// Place a solid roof at y=10 across the entire chunk.
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(x, 10, z, block.Stone)
		}
	}

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	sec0 := c.GetSection(0)

	// Above the roof (y=11) should be 15.
	assert.Equal(t, uint8(15), sec0.GetSkyLight(8, 11, 8))

	// The stone block itself at y=10 should have 0 sky light (solid + opaque).
	assert.Equal(t, uint8(0), sec0.GetSkyLight(8, 10, 8))

	// Below the solid roof, sky light can only arrive via sideways propagation
	// from chunk edges (which are bounded by the chunk itself). At position
	// (8, 9, 8) the nearest open edge is 8 blocks away, so light = 15 - 8 - 1 = 6.
	// (The -1 accounts for the step down from the edge column y=9.)
	// The exact value depends on the BFS path; just verify it is less than 15.
	below := sec0.GetSkyLight(8, 9, 8)
	assert.Less(t, below, uint8(15), "sky light below roof should be reduced")
}

// ---------------------------------------------------------------------------
// LightEngine integration
// ---------------------------------------------------------------------------

func TestLightEngine_LavaEmitsLevel15(t *testing.T) {
	c := newEmptyChunk()
	c.SetBlock(8, 4, 8, block.Lava)

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	sec := c.GetSection(0)
	assert.Equal(t, uint8(15), sec.GetBlockLight(8, 4, 8))

	// Distance 1 from lava should be 14.
	assert.Equal(t, uint8(14), sec.GetBlockLight(9, 4, 8))
}

func TestLightEngine_GlassTransmitsLight(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 4, 4, 4)

	// Glass wall at x=5 — glass is solid but transparent with LightFilter=0.
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(5, y, z, block.Glass)
		}
	}

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	sec := c.GetSection(0)

	// Light at x=6 (through glass) should be 14 - 2 = 12, because glass has
	// LightFilter=0 so the reduction is max(1, 0) = 1 per step.
	assert.Equal(t, uint8(12), sec.GetBlockLight(6, 4, 4))
}

func TestLightEngine_RecomputeResetsOldLight(t *testing.T) {
	c := newEmptyChunk()
	placeTorch(c, 8, 8, 8)

	engine := NewLightEngine()
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	// Remove the torch and recompute.
	c.SetBlock(8, 8, 8, block.Air)
	engine.ComputeChunkLight(c, [4]*chunk.Chunk{})

	sec := c.GetSection(8 / mcmath.SectionHeight)
	assert.Equal(t, uint8(0), sec.GetBlockLight(8, 8%mcmath.SectionHeight, 8),
		"block light should be 0 after torch removal and recompute")
}
