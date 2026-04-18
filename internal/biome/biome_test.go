package biome

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSeed int64 = 42

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
