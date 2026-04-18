package tick

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// stubWorld: in-memory block storage for testing
// ---------------------------------------------------------------------------

type stubWorld struct {
	blocks map[mcmath.BlockPos]uint16
}

func newStubWorld() *stubWorld {
	return &stubWorld{blocks: make(map[mcmath.BlockPos]uint16)}
}

func (w *stubWorld) GetBlock(pos mcmath.BlockPos) uint16 {
	return w.blocks[pos] // zero value is block.Air
}

func (w *stubWorld) SetBlock(pos mcmath.BlockPos, id uint16) {
	w.blocks[pos] = id
}

// ---------------------------------------------------------------------------
// RandomTick tests
// ---------------------------------------------------------------------------

func TestRandomTick_CallsHandlerForKnownBlock(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	called := false
	var calledPos mcmath.BlockPos
	reg.Register(block.Stone, func(_ BlockAccess, pos mcmath.BlockPos) {
		called = true
		calledPos = pos
	})

	// Fill chunk (0,0) entirely with stone so every random pick hits it.
	for x := int32(0); x < mcmath.ChunkSize; x++ {
		for z := int32(0); z < mcmath.ChunkSize; z++ {
			for y := int32(0); y < mcmath.ChunkHeight; y++ {
				w.SetBlock(mcmath.BlockPos{X: x, Y: y, Z: z}, block.Stone)
			}
		}
	}

	ticker := NewTickerWithSeed(w, reg, 1)
	ticker.RandomTick(0, 0)

	assert.True(t, called, "handler should be called for a known block")
	assert.GreaterOrEqual(t, calledPos.X, int32(0))
	assert.Less(t, calledPos.X, int32(mcmath.ChunkSize))
}

func TestRandomTick_SkipsBlocksWithoutHandler(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	called := false
	reg.Register(block.DiamondOre, func(_ BlockAccess, _ mcmath.BlockPos) {
		called = true
	})

	// Fill with stone, but only diamond ore has a handler.
	for x := int32(0); x < mcmath.ChunkSize; x++ {
		for z := int32(0); z < mcmath.ChunkSize; z++ {
			for y := int32(0); y < mcmath.ChunkHeight; y++ {
				w.SetBlock(mcmath.BlockPos{X: x, Y: y, Z: z}, block.Stone)
			}
		}
	}

	ticker := NewTickerWithSeed(w, reg, 2)
	ticker.RandomTick(0, 0)

	assert.False(t, called, "handler should not be called for blocks without a handler")
}

func TestRandomTick_PicksCorrectCount(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	callCount := 0
	reg.Register(block.Stone, func(_ BlockAccess, _ mcmath.BlockPos) {
		callCount++
	})

	// Fill chunk entirely with stone.
	for x := int32(0); x < mcmath.ChunkSize; x++ {
		for z := int32(0); z < mcmath.ChunkSize; z++ {
			for y := int32(0); y < mcmath.ChunkHeight; y++ {
				w.SetBlock(mcmath.BlockPos{X: x, Y: y, Z: z}, block.Stone)
			}
		}
	}

	ticker := NewTickerWithSeed(w, reg, 3)
	ticker.RandomTick(0, 0)

	numSections := mcmath.ChunkHeight / mcmath.SectionHeight
	expected := numSections * randomTicksPerSection
	assert.Equal(t, expected, callCount, "should pick 3 blocks per section")
}

// ---------------------------------------------------------------------------
// Grass spread tests
// ---------------------------------------------------------------------------

func TestGrassSpread_DirtBecomesGrassWhenAdjacentGrassAndAirAbove(t *testing.T) {
	w := newStubWorld()

	dirtPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	w.SetBlock(dirtPos, block.Dirt)
	w.SetBlock(dirtPos.Above(), block.Air)
	w.SetBlock(dirtPos.Offset(1, 0, 0), block.Grass) // adjacent grass

	grassSpreadHandler(w, dirtPos)

	assert.Equal(t, block.Grass, w.GetBlock(dirtPos), "dirt should become grass")
}

func TestGrassSpread_DirtStaysWhenNoAdjacentGrass(t *testing.T) {
	w := newStubWorld()

	dirtPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	w.SetBlock(dirtPos, block.Dirt)
	w.SetBlock(dirtPos.Above(), block.Air)

	grassSpreadHandler(w, dirtPos)

	assert.Equal(t, block.Dirt, w.GetBlock(dirtPos), "dirt should stay without adjacent grass")
}

func TestGrassSpread_DirtStaysWhenOpaqueBlockAbove(t *testing.T) {
	w := newStubWorld()

	dirtPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	w.SetBlock(dirtPos, block.Dirt)
	w.SetBlock(dirtPos.Above(), block.Stone) // opaque block above
	w.SetBlock(dirtPos.Offset(1, 0, 0), block.Grass)

	grassSpreadHandler(w, dirtPos)

	assert.Equal(t, block.Dirt, w.GetBlock(dirtPos), "dirt should stay with opaque block above")
}

func TestGrassSpread_DirtBecomesGrassWithTransparentBlockAbove(t *testing.T) {
	w := newStubWorld()

	dirtPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	w.SetBlock(dirtPos, block.Dirt)
	w.SetBlock(dirtPos.Above(), block.Glass) // transparent, LightFilter = 0
	w.SetBlock(dirtPos.Offset(1, 0, 0), block.Grass)

	grassSpreadHandler(w, dirtPos)

	assert.Equal(t, block.Grass, w.GetBlock(dirtPos), "dirt should become grass with glass above")
}

// ---------------------------------------------------------------------------
// Leaf decay tests
// ---------------------------------------------------------------------------

func TestLeafDecay_RemovesLeafWithoutNearbyLog(t *testing.T) {
	w := newStubWorld()

	leafPos := mcmath.BlockPos{X: 10, Y: 70, Z: 10}
	w.SetBlock(leafPos, block.OakLeaves)

	leafDecayHandler(w, leafPos)

	assert.Equal(t, block.Air, w.GetBlock(leafPos), "leaf should decay without nearby log")
}

func TestLeafDecay_PreservesLeafNearLog(t *testing.T) {
	w := newStubWorld()

	leafPos := mcmath.BlockPos{X: 10, Y: 70, Z: 10}
	w.SetBlock(leafPos, block.OakLeaves)
	w.SetBlock(leafPos.Offset(3, 0, 0), block.OakLog) // within 6 taxicab distance

	leafDecayHandler(w, leafPos)

	assert.Equal(t, block.OakLeaves, w.GetBlock(leafPos), "leaf should persist near log")
}

func TestLeafDecay_DecaysWhenLogTooFar(t *testing.T) {
	w := newStubWorld()

	leafPos := mcmath.BlockPos{X: 10, Y: 70, Z: 10}
	w.SetBlock(leafPos, block.OakLeaves)
	w.SetBlock(leafPos.Offset(7, 0, 0), block.OakLog) // taxicab 7 > 6

	leafDecayHandler(w, leafPos)

	assert.Equal(t, block.Air, w.GetBlock(leafPos), "leaf should decay when log is too far")
}

func TestLeafDecay_PreservesLeafAtMaxDistance(t *testing.T) {
	w := newStubWorld()

	leafPos := mcmath.BlockPos{X: 10, Y: 70, Z: 10}
	w.SetBlock(leafPos, block.OakLeaves)
	w.SetBlock(leafPos.Offset(3, 2, 1), block.OakLog) // taxicab 6, exactly at boundary

	leafDecayHandler(w, leafPos)

	assert.Equal(t, block.OakLeaves, w.GetBlock(leafPos), "leaf should persist at max distance")
}

// ---------------------------------------------------------------------------
// Scheduled tick tests
// ---------------------------------------------------------------------------

func TestScheduledTick_FiresAtCorrectTime(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	fired := false
	reg.Register(block.Water, func(_ BlockAccess, _ mcmath.BlockPos) {
		fired = true
	})

	ticker := NewTickerWithSeed(w, reg, 4)
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	ticker.Schedule(pos, block.Water, 10, 100)

	// At tick 109 the scheduled tick should not have fired.
	ticker.ProcessScheduledTicks(109)
	assert.False(t, fired, "should not fire before delay expires")

	// At tick 110 it should fire.
	ticker.ProcessScheduledTicks(110)
	assert.True(t, fired, "should fire at the scheduled tick")
}

func TestScheduledTick_Ordering(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	var order []uint16
	reg.Register(block.Water, func(_ BlockAccess, _ mcmath.BlockPos) {
		order = append(order, block.Water)
	})
	reg.Register(block.Lava, func(_ BlockAccess, _ mcmath.BlockPos) {
		order = append(order, block.Lava)
	})
	reg.Register(block.Sand, func(_ BlockAccess, _ mcmath.BlockPos) {
		order = append(order, block.Sand)
	})

	ticker := NewTickerWithSeed(w, reg, 5)
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}

	// Schedule in non-chronological order.
	ticker.Schedule(pos, block.Sand, 30, 0)  // fires at 30
	ticker.Schedule(pos, block.Water, 10, 0) // fires at 10
	ticker.Schedule(pos, block.Lava, 20, 0)  // fires at 20

	ticker.ProcessScheduledTicks(100)

	assert.Equal(t, []uint16{block.Water, block.Lava, block.Sand}, order,
		"scheduled ticks should fire in chronological order")
}

func TestScheduledTick_NoHandlerDoesNotPanic(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	ticker := NewTickerWithSeed(w, reg, 6)
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}

	ticker.Schedule(pos, block.Stone, 5, 0) // no handler for stone
	assert.NotPanics(t, func() {
		ticker.ProcessScheduledTicks(10)
	})
}

func TestScheduledTick_MultipleAtSameTick(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()

	count := 0
	reg.Register(block.Water, func(_ BlockAccess, _ mcmath.BlockPos) {
		count++
	})

	ticker := NewTickerWithSeed(w, reg, 7)
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}

	ticker.Schedule(pos, block.Water, 5, 0)
	ticker.Schedule(pos, block.Water, 5, 0)
	ticker.Schedule(pos, block.Water, 5, 0)

	ticker.ProcessScheduledTicks(5)
	assert.Equal(t, 3, count, "all three ticks at the same time should fire")
}

func TestScheduledLen(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()
	ticker := NewTickerWithSeed(w, reg, 8)

	assert.Equal(t, 0, ticker.ScheduledLen())

	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	ticker.Schedule(pos, block.Water, 10, 0)
	ticker.Schedule(pos, block.Lava, 20, 0)
	assert.Equal(t, 2, ticker.ScheduledLen())

	ticker.ProcessScheduledTicks(10)
	assert.Equal(t, 1, ticker.ScheduledLen())
}

// ---------------------------------------------------------------------------
// HandlerRegistry tests
// ---------------------------------------------------------------------------

func TestHandlerRegistry_RegisterAndGet(t *testing.T) {
	reg := NewHandlerRegistry()

	reg.Register(block.Stone, func(_ BlockAccess, _ mcmath.BlockPos) {})

	_, ok := reg.Get(block.Stone)
	assert.True(t, ok, "should find registered handler")

	_, ok = reg.Get(block.Dirt)
	assert.False(t, ok, "should not find unregistered handler")
}

func TestRegisterDefaults_RegistersDirtAndLeaves(t *testing.T) {
	reg := NewHandlerRegistry()
	RegisterDefaults(reg)

	_, ok := reg.Get(block.Dirt)
	assert.True(t, ok, "dirt handler should be registered")

	_, ok = reg.Get(block.OakLeaves)
	assert.True(t, ok, "oak leaves handler should be registered")
}

// ---------------------------------------------------------------------------
// abs32 helper
// ---------------------------------------------------------------------------

func TestAbs32(t *testing.T) {
	assert.Equal(t, int32(5), mcmath.Abs(int32(5)))
	assert.Equal(t, int32(5), mcmath.Abs(int32(-5)))
	assert.Equal(t, int32(0), mcmath.Abs(int32(0)))
}

// ---------------------------------------------------------------------------
// hasAdjacentGrass / hasNearbyLog unit tests
// ---------------------------------------------------------------------------

func TestHasAdjacentGrass_AllDirections(t *testing.T) {
	center := mcmath.BlockPos{X: 10, Y: 64, Z: 10}

	for _, neighbor := range center.Neighbors() {
		w := newStubWorld()
		w.SetBlock(neighbor, block.Grass)

		assert.True(t, hasAdjacentGrass(w, center),
			"should detect grass at %v", neighbor)
	}
}

func TestHasAdjacentGrass_NoGrass(t *testing.T) {
	w := newStubWorld()
	center := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	assert.False(t, hasAdjacentGrass(w, center))
}

func TestHasNearbyLog_AtBoundary(t *testing.T) {
	w := newStubWorld()
	center := mcmath.BlockPos{X: 10, Y: 70, Z: 10}

	w.SetBlock(center.Offset(6, 0, 0), block.OakLog)
	assert.True(t, hasNearbyLog(w, center, 6), "log at taxicab 6 should be found")

	w2 := newStubWorld()
	w2.SetBlock(center.Offset(7, 0, 0), block.OakLog)
	assert.False(t, hasNearbyLog(w2, center, 6), "log at taxicab 7 should not be found")
}
