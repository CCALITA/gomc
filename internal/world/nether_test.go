package world

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNetherWorld_DimensionID(t *testing.T) {
	w := NewNetherWorld(42)
	assert.Equal(t, Nether, w.Dimension)
}

func TestNewWorld_DimensionID(t *testing.T) {
	w := NewWorld(42)
	assert.Equal(t, Overworld, w.Dimension)
}

func TestNetherChunk_BedrockFloorAndCeiling(t *testing.T) {
	w := NewNetherWorld(12345)
	c := w.LoadChunk(mcmath.ChunkPos{X: 0, Z: 0})
	require.NotNil(t, c)

	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			assert.Equal(t, block.Bedrock, c.GetBlock(lx, 0, lz),
				"bedrock floor at (%d, 0, %d)", lx, lz)
			assert.Equal(t, block.Bedrock, c.GetBlock(lx, 127, lz),
				"bedrock ceiling at (%d, 127, %d)", lx, lz)
		}
	}
}

func TestNetherChunk_ContainsNetherrack(t *testing.T) {
	w := NewNetherWorld(12345)
	c := w.LoadChunk(mcmath.ChunkPos{X: 0, Z: 0})
	require.NotNil(t, c)

	netherrackCount := 0
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 1; y < 127; y++ {
				if c.GetBlock(lx, y, lz) == block.NetherRack {
					netherrackCount++
				}
			}
		}
	}
	assert.Greater(t, netherrackCount, 0, "nether chunk should contain netherrack")
}

func TestNetherChunk_LavaAtLevel31(t *testing.T) {
	w := NewNetherWorld(12345)
	// Check multiple chunks to find lava (caves must exist below lava level).
	lavaFound := false
	for cx := int32(0); cx < 4; cx++ {
		c := w.LoadChunk(mcmath.ChunkPos{X: cx, Z: 0})
		for lx := 0; lx < mcmath.ChunkSize; lx++ {
			for lz := 0; lz < mcmath.ChunkSize; lz++ {
				for y := 1; y <= 31; y++ {
					if c.GetBlock(lx, y, lz) == block.Lava {
						lavaFound = true
					}
				}
			}
		}
	}
	assert.True(t, lavaFound, "should find lava at or below Y=31 in nether")
}

func TestNetherChunk_GlowstoneClusters(t *testing.T) {
	w := NewNetherWorld(12345)
	glowFound := false
	for cx := int32(0); cx < 4; cx++ {
		c := w.LoadChunk(mcmath.ChunkPos{X: cx, Z: 0})
		for lx := 0; lx < mcmath.ChunkSize; lx++ {
			for lz := 0; lz < mcmath.ChunkSize; lz++ {
				for y := 120; y < 127; y++ {
					if c.GetBlock(lx, y, lz) == block.Glowstone {
						glowFound = true
					}
				}
			}
		}
	}
	assert.True(t, glowFound, "should find glowstone clusters near ceiling")
}

func TestNetherChunk_SoulSandPatches(t *testing.T) {
	w := NewNetherWorld(12345)
	soulFound := false
	for cx := int32(0); cx < 4; cx++ {
		c := w.LoadChunk(mcmath.ChunkPos{X: cx, Z: 0})
		for lx := 0; lx < mcmath.ChunkSize; lx++ {
			for lz := 0; lz < mcmath.ChunkSize; lz++ {
				for y := 32; y <= 40; y++ {
					if c.GetBlock(lx, y, lz) == block.SoulSand {
						soulFound = true
					}
				}
			}
		}
	}
	assert.True(t, soulFound, "should find soul sand patches at Y 32-40")
}

func TestNetherChunk_QuartzOre(t *testing.T) {
	w := NewNetherWorld(12345)
	quartzFound := false
	for cx := int32(0); cx < 4; cx++ {
		c := w.LoadChunk(mcmath.ChunkPos{X: cx, Z: 0})
		for lx := 0; lx < mcmath.ChunkSize; lx++ {
			for lz := 0; lz < mcmath.ChunkSize; lz++ {
				for y := 1; y < 120; y++ {
					if c.GetBlock(lx, y, lz) == block.NetherQuartzOre {
						quartzFound = true
					}
				}
			}
		}
	}
	assert.True(t, quartzFound, "should find nether quartz ore")
}

func TestNetherPortalBlockProperties(t *testing.T) {
	props := block.GetProperties(block.NetherPortal)
	assert.False(t, props.Solid)
	assert.True(t, props.Transparent)
	assert.Equal(t, uint8(11), props.LightEmission)
}

func TestNetherQuartzOreBlockProperties(t *testing.T) {
	props := block.GetProperties(block.NetherQuartzOre)
	assert.True(t, props.Solid)
	assert.False(t, props.Transparent)
	assert.Equal(t, uint8(0), props.LightEmission)
}

func TestDimensionConstants(t *testing.T) {
	assert.Equal(t, DimensionID(0), Overworld)
	assert.Equal(t, DimensionID(1), Nether)
	assert.Equal(t, 8, NetherOverworldScale)
}
