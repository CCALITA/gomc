package world

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/biome"
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// --- ShouldGenerateVillage ---

func TestShouldGenerateVillage_DeterministicWithSeed(t *testing.T) {
	vg1 := &VillageGenerator{Seed: 12345}
	vg2 := &VillageGenerator{Seed: 12345}

	// Scan a grid of chunks and verify both generators agree.
	for cx := int32(-64); cx < 64; cx++ {
		for cz := int32(-64); cz < 64; cz++ {
			pos := mcmath.ChunkPos{X: cx, Z: cz}
			r1 := vg1.ShouldGenerateVillage(pos, uint8(biome.Plains))
			r2 := vg2.ShouldGenerateVillage(pos, uint8(biome.Plains))
			assert.Equal(t, r1, r2, "determinism broken at %v", pos)
		}
	}
}

func TestShouldGenerateVillage_OnlyPlainsAndDesert(t *testing.T) {
	vg := &VillageGenerator{Seed: 42}

	rejectedBiomes := []biome.BiomeID{
		biome.Forest,
		biome.Taiga,
		biome.Jungle,
		biome.Swamp,
		biome.Mountains,
		biome.Ocean,
	}

	for _, bid := range rejectedBiomes {
		for cx := int32(-64); cx < 64; cx++ {
			for cz := int32(-64); cz < 64; cz++ {
				pos := mcmath.ChunkPos{X: cx, Z: cz}
				assert.False(t, vg.ShouldGenerateVillage(pos, uint8(bid)),
					"village should not generate in biome %d at %v", bid, pos)
			}
		}
	}
}

func TestShouldGenerateVillage_AtMostOnePerCell(t *testing.T) {
	vg := &VillageGenerator{Seed: 99}

	// Check several cells.
	for cellX := -2; cellX <= 2; cellX++ {
		for cellZ := -2; cellZ <= 2; cellZ++ {
			count := 0
			baseX := int32(cellX * villageCellSize)
			baseZ := int32(cellZ * villageCellSize)
			for dx := int32(0); dx < villageCellSize; dx++ {
				for dz := int32(0); dz < villageCellSize; dz++ {
					pos := mcmath.ChunkPos{X: baseX + dx, Z: baseZ + dz}
					if vg.ShouldGenerateVillage(pos, uint8(biome.Plains)) {
						count++
					}
				}
			}
			assert.LessOrEqual(t, count, 1,
				"cell (%d,%d) has %d villages, expected at most 1", cellX, cellZ, count)
		}
	}
}

func TestShouldGenerateVillage_PlainsAccepted(t *testing.T) {
	vg := &VillageGenerator{Seed: 42}
	found := false
	for cx := int32(0); cx < villageCellSize; cx++ {
		for cz := int32(0); cz < villageCellSize; cz++ {
			if vg.ShouldGenerateVillage(mcmath.ChunkPos{X: cx, Z: cz}, uint8(biome.Plains)) {
				found = true
			}
		}
	}
	assert.True(t, found, "should find a village in plains within one cell")
}

func TestShouldGenerateVillage_DesertAccepted(t *testing.T) {
	vg := &VillageGenerator{Seed: 42}
	found := false
	for cx := int32(0); cx < villageCellSize; cx++ {
		for cz := int32(0); cz < villageCellSize; cz++ {
			if vg.ShouldGenerateVillage(mcmath.ChunkPos{X: cx, Z: cz}, uint8(biome.Desert)) {
				found = true
			}
		}
	}
	assert.True(t, found, "should find a village in desert within one cell")
}

// --- Building dimensions ---

func TestSmallHouseDimensions(t *testing.T) {
	c := &chunk.Chunk{}
	// Prepare a flat surface at Y=64.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y <= 64; y++ {
				c.SetBlock(lx, y, lz, block.Stone)
			}
		}
	}

	placeSmallHouse(c, 2, 64, 2, block.OakPlanks)

	// Foundation: 5x5 cobblestone at Y=64.
	for dx := 0; dx < 5; dx++ {
		for dz := 0; dz < 5; dz++ {
			assert.Equal(t, block.Cobblestone, c.GetBlock(2+dx, 64, 2+dz),
				"foundation at (%d, 64, %d)", 2+dx, 2+dz)
		}
	}

	// Roof: 5x5 oak planks at Y=68.
	for dx := 0; dx < 5; dx++ {
		for dz := 0; dz < 5; dz++ {
			assert.Equal(t, block.OakPlanks, c.GetBlock(2+dx, 68, 2+dz),
				"roof at (%d, 68, %d)", 2+dx, 2+dz)
		}
	}

	// Door at front wall center.
	assert.Equal(t, block.OakDoor, c.GetBlock(4, 65, 2))
	assert.Equal(t, block.OakDoor, c.GetBlock(4, 66, 2))

	// Torch at center, Y=66 (baseY+2).
	assert.Equal(t, block.Torch, c.GetBlock(4, 66, 4))
}

func TestWellDimensions(t *testing.T) {
	c := &chunk.Chunk{}
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y <= 64; y++ {
				c.SetBlock(lx, y, lz, block.Stone)
			}
		}
	}

	placeWell(c, 3, 64, 3)

	// 3x3 cobblestone ring at Y=64.
	for dx := 0; dx < 3; dx++ {
		for dz := 0; dz < 3; dz++ {
			assert.Equal(t, block.Cobblestone, c.GetBlock(3+dx, 64, 3+dz),
				"well ring at (%d, 64, %d)", 3+dx, 3+dz)
		}
	}

	// Water inside at Y=63.
	assert.Equal(t, block.Water, c.GetBlock(4, 63, 4))

	// Fence posts on corners at Y=65.
	assert.Equal(t, block.OakFence, c.GetBlock(3, 65, 3))
	assert.Equal(t, block.OakFence, c.GetBlock(5, 65, 3))
	assert.Equal(t, block.OakFence, c.GetBlock(3, 65, 5))
	assert.Equal(t, block.OakFence, c.GetBlock(5, 65, 5))
}

func TestFarmDimensions(t *testing.T) {
	c := &chunk.Chunk{}
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y <= 64; y++ {
				c.SetBlock(lx, y, lz, block.Stone)
			}
		}
	}

	rng := rand.New(rand.NewSource(7))
	placeFarm(c, 2, 64, 2, rng)

	// Width 7 (dx 0..6), depth 3 (dz 0..2).
	for dx := 0; dx < 7; dx++ {
		lx := 2 + dx
		// Center row (dz=1) is water.
		assert.Equal(t, block.Water, c.GetBlock(lx, 64, 3),
			"water channel at (%d, 64, 3)", lx)
		// Outer rows are farmland.
		assert.Equal(t, block.Farmland, c.GetBlock(lx, 64, 2),
			"farmland at (%d, 64, 2)", lx)
		assert.Equal(t, block.Farmland, c.GetBlock(lx, 64, 4),
			"farmland at (%d, 64, 4)", lx)
		// Wheat above outer rows.
		assert.Equal(t, block.WheatCrop, c.GetBlock(lx, 65, 2),
			"wheat at (%d, 65, 2)", lx)
		assert.Equal(t, block.WheatCrop, c.GetBlock(lx, 65, 4),
			"wheat at (%d, 65, 4)", lx)
	}
}

// --- GenerateVillage building count ---

func TestGenerateVillage_BuildingCount(t *testing.T) {
	for trial := int64(0); trial < 20; trial++ {
		rng := rand.New(rand.NewSource(trial))
		count := villageMinBuildings + rng.Intn(villageMaxBuildings-villageMinBuildings+1)
		assert.GreaterOrEqual(t, count, villageMinBuildings)
		assert.LessOrEqual(t, count, villageMaxBuildings)
	}
}

// --- Integration: village appears in terrain ---

func TestVillageIntegration(t *testing.T) {
	vg := &VillageGenerator{Seed: testSeed}

	// Find the village chunk in the first cell for plains.
	var villagePos mcmath.ChunkPos
	found := false
	for cx := int32(0); cx < villageCellSize; cx++ {
		for cz := int32(0); cz < villageCellSize; cz++ {
			pos := mcmath.ChunkPos{X: cx, Z: cz}
			if vg.ShouldGenerateVillage(pos, uint8(biome.Plains)) {
				villagePos = pos
				found = true
			}
		}
	}
	require.True(t, found, "need a village position for integration test")

	// Generate a chunk with a village on it by creating a flat chunk and
	// placing the village manually.
	c := &chunk.Chunk{Pos: villagePos}
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y <= 64; y++ {
				c.SetBlock(lx, y, lz, block.Stone)
			}
			c.SetBlock(lx, 64, lz, block.Grass)
		}
	}

	rng := rand.New(rand.NewSource(testSeed))
	vg.GenerateVillage(c, uint8(biome.Plains), rng)

	// The village should have placed cobblestone (foundations/wells) and
	// gravel (paths).
	hasCobblestone := false
	hasGravel := false
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 60; y < 80; y++ {
				id := c.GetBlock(lx, y, lz)
				if id == block.Cobblestone {
					hasCobblestone = true
				}
				if id == block.Gravel {
					hasGravel = true
				}
			}
		}
	}
	assert.True(t, hasCobblestone, "village should contain cobblestone")
	assert.True(t, hasGravel, "village should contain gravel paths")
}

// --- Desert variant uses sandstone ---

func TestVillageDesertUsesSandstone(t *testing.T) {
	c := &chunk.Chunk{}
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y <= 64; y++ {
				c.SetBlock(lx, y, lz, block.Sand)
			}
		}
	}

	rng := rand.New(rand.NewSource(testSeed))
	vg := &VillageGenerator{Seed: testSeed}
	vg.GenerateVillage(c, uint8(biome.Desert), rng)

	hasSandstone := false
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 60; y < 80; y++ {
				if c.GetBlock(lx, y, lz) == block.Sandstone {
					hasSandstone = true
				}
			}
		}
	}
	assert.True(t, hasSandstone, "desert village should use sandstone walls")
}

// --- Layout helpers ---

func TestPickBuildings_AlwaysStartsWithWell(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		types := pickBuildings(5, rng)
		assert.Equal(t, buildingWell, types[0], "first building should be a well")
	}
}

func TestLayoutBuildings_FitsInChunk(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		count := villageMinBuildings + rng.Intn(villageMaxBuildings-villageMinBuildings+1)
		types := pickBuildings(count, rng)
		placements := layoutBuildings(types)
		for _, p := range placements {
			fp := footprints[p.bt]
			assert.Less(t, p.lx+fp.w, mcmath.ChunkSize,
				"building exceeds chunk X bound")
			assert.Less(t, p.lz+fp.d, mcmath.ChunkSize,
				"building exceeds chunk Z bound")
		}
	}
}

func TestCellCoords_NegativeValues(t *testing.T) {
	tests := []struct {
		pos        mcmath.ChunkPos
		wantCellX  int
		wantCellZ  int
	}{
		{mcmath.ChunkPos{X: 0, Z: 0}, 0, 0},
		{mcmath.ChunkPos{X: 31, Z: 31}, 0, 0},
		{mcmath.ChunkPos{X: 32, Z: 32}, 1, 1},
		{mcmath.ChunkPos{X: -1, Z: -1}, -1, -1},
		{mcmath.ChunkPos{X: -32, Z: -32}, -1, -1},
		{mcmath.ChunkPos{X: -33, Z: -33}, -2, -2},
	}
	for _, tt := range tests {
		cx, cz := cellCoords(tt.pos)
		assert.Equal(t, tt.wantCellX, cx, "cellX for %v", tt.pos)
		assert.Equal(t, tt.wantCellZ, cz, "cellZ for %v", tt.pos)
	}
}
