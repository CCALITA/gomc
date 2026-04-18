package world

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// newTestWorld creates a flat world with stone at Y=0 and air above, suitable for
// fluid tests. The chunk at (0,0) is pre-loaded.
func newTestWorld() *World {
	w := NewWorld(testSeed)
	w.LoadChunk(mcmath.ChunkPos{X: 0, Z: 0})
	return w
}

// prepareFlat clears a region to create a flat stone floor with air above.
func prepareFlat(w *World, minX, maxX, minZ, maxZ int32, floorY int32) {
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			w.SetBlock(mcmath.BlockPos{X: x, Y: floorY, Z: z}, block.Stone)
			for y := floorY + 1; y < floorY+10; y++ {
				w.SetBlock(mcmath.BlockPos{X: x, Y: y, Z: z}, block.Air)
			}
		}
	}
}

func TestWaterSourceSpreadsToAdjacentAir(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	// Prepare a flat area.
	prepareFlat(w, 3, 9, 3, 9, 60)

	// Place a water source.
	sourcePos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(sourcePos, block.Water)
	fs.OnBlockChange(sourcePos, block.Air, block.Water)

	// Tick enough times for water to spread.
	for i := 0; i < waterSpreadInterval+1; i++ {
		fs.Update()
	}

	// Check adjacent horizontal blocks have flowing water.
	neighbors := [4]mcmath.BlockPos{
		{X: 7, Y: 61, Z: 6},
		{X: 5, Y: 61, Z: 6},
		{X: 6, Y: 61, Z: 7},
		{X: 6, Y: 61, Z: 5},
	}
	for _, n := range neighbors {
		id := w.GetBlock(n)
		assert.True(t, block.IsWater(id), "expected water at %v, got %d", n, id)
		assert.True(t, block.IsFluid(id), "expected fluid at %v", n)
	}
}

func TestWaterLevelDecreasesWithDistance(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	// Prepare a flat area.
	prepareFlat(w, 2, 12, 5, 7, 60)

	// Place water source.
	sourcePos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(sourcePos, block.Water)
	fs.OnBlockChange(sourcePos, block.Air, block.Water)

	// Tick enough times for water to spread several blocks.
	for i := 0; i < (waterSpreadInterval+1)*maxFluidLevel; i++ {
		fs.Update()
	}

	// Check levels increase with distance from source.
	prevLevel := -1
	for dx := int32(1); dx <= 3; dx++ {
		pos := mcmath.BlockPos{X: 6 + dx, Y: 61, Z: 6}
		id := w.GetBlock(pos)
		if !block.IsWater(id) {
			break
		}
		level := block.FluidLevel(id)
		if prevLevel >= 0 {
			assert.Greater(t, level, prevLevel, "level should increase at distance %d", dx)
		}
		prevLevel = level
	}
	assert.Greater(t, prevLevel, 0, "should have found flowing water with increasing levels")
}

func TestWaterFlowsDown(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	// Create a ledge: stone platform at Y=62 with a gap (air) at X=6,Z=6.
	prepareFlat(w, 5, 7, 5, 7, 60)

	// Place air at Y=61 to allow downward flow.
	gapPos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(gapPos, block.Air)

	// Place water source at Y=62 directly above the gap.
	sourcePos := mcmath.BlockPos{X: 6, Y: 62, Z: 6}
	w.SetBlock(sourcePos, block.Water)
	fs.OnBlockChange(sourcePos, block.Air, block.Water)

	// Tick for water to spread down.
	for i := 0; i < waterSpreadInterval+1; i++ {
		fs.Update()
	}

	// Check below has flowing water.
	belowID := w.GetBlock(gapPos)
	assert.True(t, block.IsWater(belowID), "water should flow down, got %d", belowID)
}

func TestLavaSpreadsSlower(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	prepareFlat(w, 3, 9, 3, 9, 60)

	// Place lava source.
	lavaPos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(lavaPos, block.Lava)
	fs.OnBlockChange(lavaPos, block.Air, block.Lava)

	// After water spread interval ticks, lava should NOT have spread yet.
	for i := 0; i < waterSpreadInterval+1; i++ {
		fs.Update()
	}

	neighbor := mcmath.BlockPos{X: 7, Y: 61, Z: 6}
	neighborID := w.GetBlock(neighbor)
	assert.False(t, block.IsLava(neighborID),
		"lava should not have spread after %d ticks", waterSpreadInterval+1)

	// After lava spread interval ticks, lava should have spread.
	for i := 0; i < lavaSpreadInterval; i++ {
		fs.Update()
	}

	neighborID = w.GetBlock(neighbor)
	assert.True(t, block.IsLava(neighborID),
		"lava should have spread after %d ticks", lavaSpreadInterval+waterSpreadInterval+1)
}

func TestLavaWaterInteractionCreatesCobblestone(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	prepareFlat(w, 4, 8, 5, 7, 60)

	// Place flowing lava manually.
	lavaPos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(lavaPos, block.FlowingLavaLevel(1))

	// Place water source adjacent.
	waterPos := mcmath.BlockPos{X: 7, Y: 61, Z: 6}
	w.SetBlock(waterPos, block.Water)
	fs.OnBlockChange(waterPos, block.Air, block.Water)

	// The interaction check should turn the flowing lava into cobblestone.
	lavaID := w.GetBlock(lavaPos)
	assert.Equal(t, block.Cobblestone, lavaID,
		"flowing lava + water should create cobblestone")
}

func TestLavaSourceWaterInteractionCreatesObsidian(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	prepareFlat(w, 4, 8, 5, 7, 60)

	// Place lava source.
	lavaPos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(lavaPos, block.Lava)

	// Place water source adjacent.
	waterPos := mcmath.BlockPos{X: 7, Y: 61, Z: 6}
	w.SetBlock(waterPos, block.Water)
	fs.OnBlockChange(waterPos, block.Air, block.Water)

	// The interaction should turn lava source into obsidian.
	lavaID := w.GetBlock(lavaPos)
	assert.Equal(t, block.Obsidian, lavaID,
		"lava source + water should create obsidian")
}

func TestSourceRemovalDrainsFlowingBlocks(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	prepareFlat(w, 3, 9, 3, 9, 60)

	// Place water source and let it spread.
	sourcePos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(sourcePos, block.Water)
	fs.OnBlockChange(sourcePos, block.Air, block.Water)

	for i := 0; i < (waterSpreadInterval+1)*3; i++ {
		fs.Update()
	}

	// Verify water has spread.
	neighbor := mcmath.BlockPos{X: 7, Y: 61, Z: 6}
	assert.True(t, block.IsWater(w.GetBlock(neighbor)), "water should have spread")

	// Remove the source.
	w.SetBlock(sourcePos, block.Air)
	fs.OnBlockChange(sourcePos, block.Water, block.Air)

	// All flowing water connected to the source should be removed.
	assert.Equal(t, block.Air, w.GetBlock(neighbor),
		"flowing water should drain after source removal")
}

func TestFluidDoesNotSpreadThroughSolidBlocks(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)

	prepareFlat(w, 4, 8, 4, 8, 60)

	// Build a wall around the source on all four sides.
	sourcePos := mcmath.BlockPos{X: 6, Y: 61, Z: 6}
	w.SetBlock(mcmath.BlockPos{X: 7, Y: 61, Z: 6}, block.Stone)
	w.SetBlock(mcmath.BlockPos{X: 5, Y: 61, Z: 6}, block.Stone)
	w.SetBlock(mcmath.BlockPos{X: 6, Y: 61, Z: 7}, block.Stone)
	w.SetBlock(mcmath.BlockPos{X: 6, Y: 61, Z: 5}, block.Stone)
	// Also block below.
	w.SetBlock(mcmath.BlockPos{X: 6, Y: 60, Z: 6}, block.Stone)

	// Place water source.
	w.SetBlock(sourcePos, block.Water)
	fs.OnBlockChange(sourcePos, block.Air, block.Water)

	// Tick multiple times.
	for i := 0; i < (waterSpreadInterval+1)*3; i++ {
		fs.Update()
	}

	// Blocks beyond the wall should remain as they were (not water).
	beyondWall := []mcmath.BlockPos{
		{X: 8, Y: 61, Z: 6},
		{X: 4, Y: 61, Z: 6},
		{X: 6, Y: 61, Z: 8},
		{X: 6, Y: 61, Z: 4},
	}
	for _, bp := range beyondWall {
		id := w.GetBlock(bp)
		assert.False(t, block.IsWater(id),
			"water should not spread through solid wall at %v, got %d", bp, id)
	}
}

// --- Block helper tests ---

func TestFlowingWaterLevel(t *testing.T) {
	tests := []struct {
		level    int
		wantBase uint16
	}{
		{0, block.FlowingWater},
		{1, block.FlowingWater},
		{7, block.FlowingWater},
	}
	for _, tc := range tests {
		id := block.FlowingWaterLevel(tc.level)
		assert.Equal(t, tc.wantBase, block.BaseID(id), "base ID for level %d", tc.level)
		assert.Equal(t, tc.level, block.FluidLevel(id), "level for FlowingWaterLevel(%d)", tc.level)
		assert.True(t, block.IsWater(id))
		assert.True(t, block.IsFluid(id))
		assert.False(t, block.IsLava(id))
		assert.False(t, block.IsSolid(id))
	}
}

func TestFlowingLavaLevel(t *testing.T) {
	tests := []struct {
		level    int
		wantBase uint16
	}{
		{0, block.FlowingLava},
		{3, block.FlowingLava},
		{7, block.FlowingLava},
	}
	for _, tc := range tests {
		id := block.FlowingLavaLevel(tc.level)
		assert.Equal(t, tc.wantBase, block.BaseID(id), "base ID for level %d", tc.level)
		assert.Equal(t, tc.level, block.FluidLevel(id), "level for FlowingLavaLevel(%d)", tc.level)
		assert.True(t, block.IsLava(id))
		assert.True(t, block.IsFluid(id))
		assert.False(t, block.IsWater(id))
		assert.False(t, block.IsSolid(id))
	}
}

func TestIsFluidForSources(t *testing.T) {
	assert.True(t, block.IsFluid(block.Water))
	assert.True(t, block.IsFluid(block.Lava))
	assert.False(t, block.IsFluid(block.Air))
	assert.False(t, block.IsFluid(block.Stone))
}

func TestFluidLevelForSources(t *testing.T) {
	assert.Equal(t, 0, block.FluidLevel(block.Water))
	assert.Equal(t, 0, block.FluidLevel(block.Lava))
	assert.Equal(t, 0, block.FluidLevel(block.Stone))
}

func TestNewFluidSimulator(t *testing.T) {
	w := newTestWorld()
	fs := NewFluidSimulator(w)
	assert.NotNil(t, fs)
	assert.NotNil(t, fs.world)
}
