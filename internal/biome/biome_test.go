package biome

import (
	"math"
	"math/rand"
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSeed int64 = 42

// allBiomeIDs enumerates every predefined BiomeID for table-driven tests.
var allBiomeIDs = []BiomeID{Plains, Forest, Desert, Taiga, Jungle, Swamp, Mountains, Ocean}

func TestBiomeAtDeterministic(t *testing.T) {
	bm1 := NewBiomeMap(testSeed)
	bm2 := NewBiomeMap(testSeed)

	coords := [][2]int{
		{0, 0}, {100, 200}, {-500, 300}, {9999, -9999},
	}
	for _, c := range coords {
		b1 := bm1.BiomeAt(c[0], c[1])
		b2 := bm2.BiomeAt(c[0], c[1])
		assert.Equal(t, b1.ID, b2.ID,
			"biome should be deterministic at (%d, %d)", c[0], c[1])
	}
}

func TestBiomeVarietyAcrossLargeArea(t *testing.T) {
	bm := NewBiomeMap(testSeed)

	seen := make(map[BiomeID]bool)
	for x := -500; x <= 500; x += 10 {
		for z := -500; z <= 500; z += 10 {
			b := bm.BiomeAt(x, z)
			seen[b.ID] = true
		}
	}

	assert.GreaterOrEqual(t, len(seen), 5,
		"expected at least 5 biome types in a 1000x1000 area, got %d: %v", len(seen), seen)
}

func TestAllBiomesHaveValidBlocks(t *testing.T) {
	validBlocks := map[uint16]bool{
		block.Grass: true,
		block.Dirt:  true,
		block.Sand:  true,
		block.Stone: true,
		block.Snow:  true,
	}

	for id := BiomeID(0); id < 8; id++ {
		b := ByID(id)
		assert.True(t, validBlocks[b.SurfaceBlock],
			"biome %q surface block %d should be a valid surface block", b.Name, b.SurfaceBlock)
		assert.True(t, validBlocks[b.SubsurfaceBlock] || b.SubsurfaceBlock == block.Sand || b.SubsurfaceBlock == block.Stone,
			"biome %q subsurface block %d should be a valid subsurface block", b.Name, b.SubsurfaceBlock)
	}
}

func TestDesertIsHotAndDry(t *testing.T) {
	b := ByID(Desert)
	assert.Greater(t, b.Temperature, 1.5,
		"desert temperature should be high")
	assert.LessOrEqual(t, b.Rainfall, 0.1,
		"desert rainfall should be very low")
	assert.Equal(t, uint16(block.Sand), b.SurfaceBlock,
		"desert surface should be sand")
	assert.Equal(t, uint16(block.Sand), b.SubsurfaceBlock,
		"desert subsurface should be sand")
	assert.Equal(t, 0.0, b.TreeDensity,
		"desert should have no trees")
}

func TestTaigaIsCold(t *testing.T) {
	b := ByID(Taiga)
	assert.Less(t, b.Temperature, 0.3,
		"taiga temperature should be cold")
	assert.Equal(t, TreeSpruce, b.Tree,
		"taiga trees should be spruce")
}

func TestOceanHasLowBaseHeight(t *testing.T) {
	b := ByID(Ocean)
	assert.LessOrEqual(t, b.BaseHeight, 45.0,
		"ocean base height should be low (<=45)")
}

func TestMountainsHaveHighAmplitude(t *testing.T) {
	b := ByID(Mountains)
	assert.GreaterOrEqual(t, b.HeightAmplitude, 50.0,
		"mountains should have high height amplitude (>=50)")
}

func TestByIDOutOfRange(t *testing.T) {
	b := ByID(BiomeID(255))
	assert.Equal(t, Plains, b.ID,
		"out-of-range ID should fall back to Plains")
}

func TestClassifyWhittakerCoverage(t *testing.T) {
	seen := make(map[BiomeID]bool)
	steps := 20
	for ti := 0; ti <= steps; ti++ {
		for ri := 0; ri <= steps; ri++ {
			temp := float64(ti) / float64(steps)
			rain := float64(ri) / float64(steps)
			b := classify(temp, rain)
			seen[b.ID] = true
		}
	}
	require.GreaterOrEqual(t, len(seen), 7,
		"Whittaker classification should produce at least 7 biome types across the full temp/rain range, got %d", len(seen))
}

func TestAllBiomesHaveDistinctSurfaceBlockCombinations(t *testing.T) {
	// At least 3 distinct surface blocks must appear across all biomes.
	surfaceBlocks := make(map[uint16][]string)
	for _, id := range allBiomeIDs {
		b := ByID(id)
		surfaceBlocks[b.SurfaceBlock] = append(surfaceBlocks[b.SurfaceBlock], b.Name)
	}
	assert.GreaterOrEqual(t, len(surfaceBlocks), 3,
		"expected at least 3 distinct surface blocks across 8 biomes, got %d: %v",
		len(surfaceBlocks), surfaceBlocks)

	type blockPair struct{ surface, subsurface uint16 }
	pairs := make(map[blockPair]int)
	for _, id := range allBiomeIDs {
		b := ByID(id)
		pairs[blockPair{b.SurfaceBlock, b.SubsurfaceBlock}]++
	}
	assert.GreaterOrEqual(t, len(pairs), 3,
		"expected at least 3 distinct (surface, subsurface) block pairs, got %d", len(pairs))
}

func TestAllBiomesHaveNonEmptyName(t *testing.T) {
	for _, id := range allBiomeIDs {
		b := ByID(id)
		assert.NotEmpty(t, b.Name, "biome ID %d should have a non-empty Name", id)
	}
}

func TestBiomeAtExtremeCoordinatesDoesNotPanic(t *testing.T) {
	bm := NewBiomeMap(testSeed)

	extremes := [][2]int{
		{math.MaxInt32, math.MaxInt32},
		{math.MinInt32, math.MinInt32},
		{math.MaxInt32, math.MinInt32},
		{math.MinInt32, math.MaxInt32},
		{0, 0},
		{1<<30 - 1, -(1 << 30)},
	}

	for _, c := range extremes {
		b := bm.BiomeAt(c[0], c[1])
		assert.NotEmpty(t, b.Name,
			"BiomeAt(%d, %d) should return a valid biome", c[0], c[1])
	}
}

func TestBiomeAtReturnsNonNilFor100RandomCoords(t *testing.T) {
	bm := NewBiomeMap(testSeed)
	rng := rand.New(rand.NewSource(12345))

	for i := 0; i < 100; i++ {
		x := rng.Intn(200001) - 100000 // range [-100000, 100000]
		z := rng.Intn(200001) - 100000
		b := bm.BiomeAt(x, z)
		assert.NotEmpty(t, b.Name,
			"BiomeAt(%d, %d) must return a biome with a non-empty Name (iteration %d)", x, z, i)
		assert.True(t, b.ID <= 7,
			"BiomeAt(%d, %d) returned unknown biome ID %d", x, z, b.ID)
	}
}

func TestDifferentSeedsProduceDifferentBiomeMaps(t *testing.T) {
	bm1 := NewBiomeMap(1)
	bm2 := NewBiomeMap(999999)

	diffCount := 0
	total := 0
	for x := -200; x <= 200; x += 20 {
		for z := -200; z <= 200; z += 20 {
			total++
			if bm1.BiomeAt(x, z).ID != bm2.BiomeAt(x, z).ID {
				diffCount++
			}
		}
	}

	assert.Greater(t, diffCount, 0,
		"two different seeds should produce at least one differing biome in a 400x400 area (%d coords sampled)", total)
}

func TestEachBiomeHasReasonableHeightParameters(t *testing.T) {
	for _, id := range allBiomeIDs {
		b := ByID(id)
		assert.Greater(t, b.HeightAmplitude, 0.0,
			"biome %q HeightAmplitude should be positive", b.Name)
		assert.LessOrEqual(t, b.HeightAmplitude, 128.0,
			"biome %q HeightAmplitude should be at most 128", b.Name)
		assert.GreaterOrEqual(t, b.BaseHeight, 0.0,
			"biome %q BaseHeight should be non-negative", b.Name)
		assert.LessOrEqual(t, b.BaseHeight, 256.0,
			"biome %q BaseHeight should be at most 256", b.Name)
	}
}

func TestOceanHasLowestBaseHeight(t *testing.T) {
	ocean := ByID(Ocean)
	for _, id := range allBiomeIDs {
		if id == Ocean {
			continue
		}
		b := ByID(id)
		assert.Less(t, ocean.BaseHeight, b.BaseHeight,
			"ocean BaseHeight (%.1f) should be strictly less than %s BaseHeight (%.1f)",
			ocean.BaseHeight, b.Name, b.BaseHeight)
	}
}

func TestMountainsHasHighestHeightAmplitude(t *testing.T) {
	mountains := ByID(Mountains)
	for _, id := range allBiomeIDs {
		if id == Mountains {
			continue
		}
		b := ByID(id)
		assert.Greater(t, mountains.HeightAmplitude, b.HeightAmplitude,
			"mountains HeightAmplitude (%.1f) should be strictly greater than %s HeightAmplitude (%.1f)",
			mountains.HeightAmplitude, b.Name, b.HeightAmplitude)
	}
}
