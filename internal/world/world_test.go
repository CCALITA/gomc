package world

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const testSeed int64 = 42

// --- TerrainGenerator tests ---

func TestTerrainHeightInExpectedRange(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)
	c := gen.GenerateChunk(mcmath.ChunkPos{X: 0, Z: 0})

	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			highest := c.HighestBlock(lx, lz)
			// Height should be somewhere between baseHeight-heightAmp and baseHeight+heightAmp
			// plus some margin for trees and leaves.
			assert.GreaterOrEqual(t, highest, 0, "highest block should be >= 0 at (%d,%d)", lx, lz)
			assert.Less(t, highest, mcmath.ChunkHeight, "highest block should be < ChunkHeight at (%d,%d)", lx, lz)
		}
	}
}

func TestBedrockAtY0(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)
	c := gen.GenerateChunk(mcmath.ChunkPos{X: 0, Z: 0})

	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			id := c.GetBlock(lx, 0, lz)
			assert.Equal(t, block.Bedrock, id, "Y=0 should always be bedrock at (%d,%d)", lx, lz)
		}
	}
}

func TestBedrockInLowerLayers(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)
	c := gen.GenerateChunk(mcmath.ChunkPos{X: 1, Z: 1})

	hasBedrock := false
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 1; y <= bedrockRandMax; y++ {
				if c.GetBlock(lx, y, lz) == block.Bedrock {
					hasBedrock = true
				}
			}
		}
	}
	assert.True(t, hasBedrock, "there should be some random bedrock in Y=1..4")
}

func TestWaterAtSeaLevel(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)

	// Test several chunks to increase the chance of finding low terrain.
	hasWater := false
	for cx := int32(-5); cx <= 5; cx++ {
		for cz := int32(-5); cz <= 5; cz++ {
			c := gen.GenerateChunk(mcmath.ChunkPos{X: cx, Z: cz})
			for lx := 0; lx < mcmath.ChunkSize; lx++ {
				for lz := 0; lz < mcmath.ChunkSize; lz++ {
					if c.GetBlock(lx, seaLevel, lz) == block.Water {
						hasWater = true
					}
				}
			}
		}
	}
	assert.True(t, hasWater, "there should be water at sea level in some chunks")
}

func TestCavesCarved(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)

	hasAirBelowSurface := false
	for cx := int32(-3); cx <= 3; cx++ {
		for cz := int32(-3); cz <= 3; cz++ {
			c := gen.GenerateChunk(mcmath.ChunkPos{X: cx, Z: cz})
			for lx := 0; lx < mcmath.ChunkSize; lx++ {
				for lz := 0; lz < mcmath.ChunkSize; lz++ {
					h := c.HighestBlock(lx, lz)
					for y := caveMinY; y < h-4; y++ {
						if c.GetBlock(lx, y, lz) == block.Air {
							hasAirBelowSurface = true
						}
					}
				}
			}
		}
	}
	assert.True(t, hasAirBelowSurface, "caves should carve air pockets below the surface")
}

func TestOresGenerated(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)

	oreFound := map[block.BlockID]bool{}
	for cx := int32(-3); cx <= 3; cx++ {
		for cz := int32(-3); cz <= 3; cz++ {
			c := gen.GenerateChunk(mcmath.ChunkPos{X: cx, Z: cz})
			for lx := 0; lx < mcmath.ChunkSize; lx++ {
				for lz := 0; lz < mcmath.ChunkSize; lz++ {
					for y := 1; y < 80; y++ {
						id := c.GetBlock(lx, y, lz)
						if id == block.CoalOre || id == block.IronOre || id == block.GoldOre || id == block.DiamondOre {
							oreFound[id] = true
						}
					}
				}
			}
		}
	}
	assert.True(t, oreFound[block.CoalOre], "coal ore should be generated")
	assert.True(t, oreFound[block.IronOre], "iron ore should be generated")
}

func TestTreesGenerated(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)

	hasLog := false
	hasLeaves := false
	for cx := int32(-5); cx <= 5; cx++ {
		for cz := int32(-5); cz <= 5; cz++ {
			c := gen.GenerateChunk(mcmath.ChunkPos{X: cx, Z: cz})
			for lx := 0; lx < mcmath.ChunkSize; lx++ {
				for lz := 0; lz < mcmath.ChunkSize; lz++ {
					for y := seaLevel; y < mcmath.ChunkHeight; y++ {
						id := c.GetBlock(lx, y, lz)
						if id == block.OakLog {
							hasLog = true
						}
						if id == block.OakLeaves {
							hasLeaves = true
						}
					}
				}
			}
		}
	}
	assert.True(t, hasLog, "there should be oak log blocks from trees")
	assert.True(t, hasLeaves, "there should be oak leaves blocks from trees")
}

// --- World tests ---

func TestWorldGetSetBlock(t *testing.T) {
	w := NewWorld(testSeed)

	pos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	cp := pos.ToChunkPos()
	w.LoadChunk(cp)

	// SetBlock should modify the loaded chunk.
	w.SetBlock(pos, block.Stone)
	id := w.GetBlock(pos)
	assert.Equal(t, block.Stone, id)

	// Overwrite with a different block.
	w.SetBlock(pos, block.Dirt)
	id = w.GetBlock(pos)
	assert.Equal(t, block.Dirt, id)
}

func TestWorldGetBlockUnloaded(t *testing.T) {
	w := NewWorld(testSeed)

	pos := mcmath.BlockPos{X: 5000, Y: 64, Z: 5000}
	id := w.GetBlock(pos)
	assert.Equal(t, block.Air, id, "unloaded chunk should return Air")
}

func TestWorldSetBlockUnloaded(t *testing.T) {
	w := NewWorld(testSeed)

	// SetBlock on unloaded chunk should be a no-op.
	pos := mcmath.BlockPos{X: 5000, Y: 64, Z: 5000}
	w.SetBlock(pos, block.Stone)
	id := w.GetBlock(pos)
	assert.Equal(t, block.Air, id)
}

func TestWorldGetBlockOutOfRange(t *testing.T) {
	w := NewWorld(testSeed)

	below := mcmath.BlockPos{X: 0, Y: -1, Z: 0}
	assert.Equal(t, block.Air, w.GetBlock(below))

	above := mcmath.BlockPos{X: 0, Y: 256, Z: 0}
	assert.Equal(t, block.Air, w.GetBlock(above))
}

func TestChunkLoadUnload(t *testing.T) {
	w := NewWorld(testSeed)
	pos := mcmath.ChunkPos{X: 3, Z: 3}

	assert.Nil(t, w.GetChunk(pos), "chunk should not be loaded initially")

	c := w.LoadChunk(pos)
	assert.NotNil(t, c)
	assert.Equal(t, pos, c.Pos)
	assert.Equal(t, 1, w.LoadedChunkCount())

	// Loading again should return the same chunk.
	c2 := w.LoadChunk(pos)
	assert.Equal(t, c, c2)

	w.UnloadChunk(pos)
	assert.Nil(t, w.GetChunk(pos))
	assert.Equal(t, 0, w.LoadedChunkCount())
}

func TestGetBlockAABBsReturnsSolidBlocks(t *testing.T) {
	w := NewWorld(testSeed)
	cp := mcmath.ChunkPos{X: 0, Z: 0}
	w.LoadChunk(cp)

	// Query a small region around Y=0 where bedrock exists.
	region := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 2, Y: 2, Z: 2},
	}
	aabbs := w.GetBlockAABBs(region)
	assert.NotEmpty(t, aabbs, "should find solid blocks (bedrock) near Y=0")

	// Each AABB should be a unit cube.
	for _, ab := range aabbs {
		size := ab.Size()
		assert.InDelta(t, 1.0, float64(size.X), 0.001)
		assert.InDelta(t, 1.0, float64(size.Y), 0.001)
		assert.InDelta(t, 1.0, float64(size.Z), 0.001)
	}
}

func TestGetBlockAABBsEmptyRegion(t *testing.T) {
	w := NewWorld(testSeed)
	cp := mcmath.ChunkPos{X: 0, Z: 0}
	w.LoadChunk(cp)

	// Query a region high above terrain where there are only air blocks.
	region := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 250, Z: 0},
		Max: mcmath.Vec3{X: 2, Y: 252, Z: 2},
	}
	aabbs := w.GetBlockAABBs(region)
	assert.Empty(t, aabbs, "should find no solid blocks at Y=250")
}

// --- ChunkLoader tests ---

func TestLoaderLoadsChunksWithinRadius(t *testing.T) {
	w := NewWorld(testSeed)
	loader := NewChunkLoader(w)
	loader.LoadRadius = 2
	defer loader.Stop()

	center := mcmath.ChunkPos{X: 0, Z: 0}
	loader.Update(center)

	// Wait for async loading to complete.
	assert.Eventually(t, func() bool {
		return loader.PendingLoads() == 0
	}, 10*time.Second, 50*time.Millisecond, "all pending loads should finish")

	// Check that chunks within radius are loaded.
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			cp := mcmath.ChunkPos{X: dx, Z: dz}
			if center.Distance(cp) <= 2.0 {
				assert.NotNil(t, w.GetChunk(cp), "chunk at %v should be loaded", cp)
			}
		}
	}
}

func TestLoaderUnloadsDistantChunks(t *testing.T) {
	w := NewWorld(testSeed)
	loader := NewChunkLoader(w)
	loader.LoadRadius = 2
	defer loader.Stop()

	// Load chunks around origin.
	loader.Update(mcmath.ChunkPos{X: 0, Z: 0})
	assert.Eventually(t, func() bool {
		return loader.PendingLoads() == 0
	}, 10*time.Second, 50*time.Millisecond)

	// Move center far away — old chunks should be unloaded.
	loader.Update(mcmath.ChunkPos{X: 100, Z: 100})
	assert.Eventually(t, func() bool {
		return loader.PendingLoads() == 0
	}, 10*time.Second, 50*time.Millisecond)

	// The origin chunk should be unloaded since it is far from the new center.
	assert.Nil(t, w.GetChunk(mcmath.ChunkPos{X: 0, Z: 0}), "chunk at origin should be unloaded after moving away")
}

func TestLoaderLoadedCount(t *testing.T) {
	w := NewWorld(testSeed)
	loader := NewChunkLoader(w)
	loader.LoadRadius = 1
	defer loader.Stop()

	loader.Update(mcmath.ChunkPos{X: 0, Z: 0})
	assert.Eventually(t, func() bool {
		return loader.PendingLoads() == 0
	}, 10*time.Second, 50*time.Millisecond)

	assert.Greater(t, loader.LoadedCount(), 0, "loaded count should be > 0 after loading")
}

func TestLoaderStopIdempotent(t *testing.T) {
	w := NewWorld(testSeed)
	loader := NewChunkLoader(w)

	// Calling Stop multiple times should not panic.
	loader.Stop()
	loader.Stop()
}

// --- abs helper test ---

func TestAbsHelper(t *testing.T) {
	assert.Equal(t, 5, abs(5))
	assert.Equal(t, 5, abs(-5))
	assert.Equal(t, 0, abs(0))
}

// --- Terrain determinism test ---

func TestTerrainDeterministic(t *testing.T) {
	gen1 := NewTerrainGenerator(testSeed)
	gen2 := NewTerrainGenerator(testSeed)

	pos := mcmath.ChunkPos{X: 7, Z: -3}
	c1 := gen1.GenerateChunk(pos)
	c2 := gen2.GenerateChunk(pos)

	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y < mcmath.ChunkHeight; y++ {
				assert.Equal(t, c1.GetBlock(lx, y, lz), c2.GetBlock(lx, y, lz),
					"blocks should be identical at (%d,%d,%d)", lx, y, lz)
			}
		}
	}
}

// --- Sand at shoreline test ---

func TestSandAtShoreline(t *testing.T) {
	gen := NewTerrainGenerator(testSeed)

	hasSand := false
	for cx := int32(-5); cx <= 5; cx++ {
		for cz := int32(-5); cz <= 5; cz++ {
			c := gen.GenerateChunk(mcmath.ChunkPos{X: cx, Z: cz})
			for lx := 0; lx < mcmath.ChunkSize; lx++ {
				for lz := 0; lz < mcmath.ChunkSize; lz++ {
					for y := seaLevel - 3; y <= seaLevel; y++ {
						if y >= 0 && c.GetBlock(lx, y, lz) == block.Sand {
							hasSand = true
						}
					}
				}
			}
		}
	}
	assert.True(t, hasSand, "sand should appear at shorelines near sea level")
}
