package tick

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestPistonExtend_PushSingleBlock(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	blockPos := mcmath.BlockPos{X: 0, Y: 64, Z: 4} // North is -Z

	w.SetBlock(pistonPos, block.Piston)
	w.SetBlock(blockPos, block.Stone)

	PistonExtend(w, pistonPos, mcmath.North)

	// Stone should have moved one step north.
	assert.Equal(t, block.Stone, w.GetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 3}))
	// Piston head should be at the block's old position.
	assert.Equal(t, block.PistonHead, w.GetBlock(blockPos))
	// Piston should be marked extended.
	assert.True(t, block.IsPistonExtended(w.GetBlock(pistonPos)))
}

func TestPistonExtend_PushChainOfThree(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 10}
	w.SetBlock(pistonPos, block.Piston)
	w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 9}, block.Stone)
	w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 8}, block.Dirt)
	w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 7}, block.Sand)

	PistonExtend(w, pistonPos, mcmath.North)

	assert.Equal(t, block.PistonHead, w.GetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 9}))
	assert.Equal(t, block.Stone, w.GetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 8}))
	assert.Equal(t, block.Dirt, w.GetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 7}))
	assert.Equal(t, block.Sand, w.GetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 6}))
}

func TestCanPistonPush_CannotPushMoreThan12(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 20}
	w.SetBlock(pistonPos, block.Piston)

	// Place 13 blocks in front of the piston.
	for i := int32(1); i <= 13; i++ {
		w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 20 - i}, block.Stone)
	}

	assert.False(t, CanPistonPush(w, pistonPos, mcmath.North, 12))
}

func TestCanPistonPush_CannotPushObsidian(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	w.SetBlock(pistonPos, block.Piston)
	w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 4}, block.Obsidian)

	assert.False(t, CanPistonPush(w, pistonPos, mcmath.North, 12))
}

func TestCanPistonPush_CannotPushBedrock(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	w.SetBlock(pistonPos, block.Piston)
	w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 4}, block.Bedrock)

	assert.False(t, CanPistonPush(w, pistonPos, mcmath.North, 12))
}

func TestStickyPistonRetract_PullsBlock(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	headPos := mcmath.BlockPos{X: 0, Y: 64, Z: 4}
	pullPos := mcmath.BlockPos{X: 0, Y: 64, Z: 3}

	w.SetBlock(pistonPos, block.WithPistonExtended(block.StickyPiston, true))
	w.SetBlock(headPos, block.PistonHead)
	w.SetBlock(pullPos, block.Stone)

	PistonRetract(w, pistonPos, mcmath.North, true)

	// Stone should be pulled to head position.
	assert.Equal(t, block.Stone, w.GetBlock(headPos))
	// Pull position should be air.
	assert.Equal(t, block.Air, w.GetBlock(pullPos))
	// Piston should be retracted.
	assert.False(t, block.IsPistonExtended(w.GetBlock(pistonPos)))
}

func TestPistonExtendRetract_StateToggle(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	headPos := mcmath.BlockPos{X: 0, Y: 64, Z: 4}

	w.SetBlock(pistonPos, block.Piston)
	// Nothing in front, extend into air.
	PistonExtend(w, pistonPos, mcmath.North)

	assert.True(t, block.IsPistonExtended(w.GetBlock(pistonPos)))
	assert.Equal(t, block.PistonHead, w.GetBlock(headPos))

	// Retract.
	PistonRetract(w, pistonPos, mcmath.North, false)

	assert.False(t, block.IsPistonExtended(w.GetBlock(pistonPos)))
	assert.Equal(t, block.Air, w.GetBlock(headPos))
}

func TestCanPistonPush_EmptySpace(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	w.SetBlock(pistonPos, block.Piston)
	// Nothing in front.
	assert.True(t, CanPistonPush(w, pistonPos, mcmath.North, 12))
}

func TestCanPistonPush_Exactly12Blocks(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 20}
	w.SetBlock(pistonPos, block.Piston)

	// Place exactly 12 blocks.
	for i := int32(1); i <= 12; i++ {
		w.SetBlock(mcmath.BlockPos{X: 0, Y: 64, Z: 20 - i}, block.Stone)
	}

	assert.True(t, CanPistonPush(w, pistonPos, mcmath.North, 12))
}

func TestPistonRetract_NormalDoesNotPull(t *testing.T) {
	w := newStubWorld()
	pistonPos := mcmath.BlockPos{X: 0, Y: 64, Z: 5}
	headPos := mcmath.BlockPos{X: 0, Y: 64, Z: 4}
	beyondPos := mcmath.BlockPos{X: 0, Y: 64, Z: 3}

	w.SetBlock(pistonPos, block.WithPistonExtended(block.Piston, true))
	w.SetBlock(headPos, block.PistonHead)
	w.SetBlock(beyondPos, block.Stone)

	PistonRetract(w, pistonPos, mcmath.North, false)

	// Head position should be air (not pulled stone).
	assert.Equal(t, block.Air, w.GetBlock(headPos))
	// Stone should remain.
	assert.Equal(t, block.Stone, w.GetBlock(beyondPos))
}
