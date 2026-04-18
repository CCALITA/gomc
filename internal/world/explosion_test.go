package world

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// newExplosionTestWorld creates a flat world with a single loaded chunk at
// (0,0) filled with stone from Y=0 to Y=fillHeight-1.
func newExplosionTestWorld(fillHeight int) *World {
	w := NewWorld(0)
	// Load chunk at origin so SetBlock/GetBlock work.
	w.LoadChunk(mcmath.ChunkPos{X: 0, Z: 0})

	for x := int32(0); x < mcmath.ChunkSize; x++ {
		for z := int32(0); z < mcmath.ChunkSize; z++ {
			for y := int32(0); y < int32(fillHeight); y++ {
				w.SetBlock(mcmath.BlockPos{X: x, Y: y, Z: z}, block.Stone)
			}
		}
	}
	return w
}

// deterministicRng returns a seeded RNG for reproducible tests.
func deterministicRng() *rand.Rand {
	return rand.New(rand.NewSource(1))
}

// --- Block destruction ---

func TestExplodeBlocksDestroysBlocksWithinRadius(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)
	radius := power * 1.5 // 6.0

	result := ExplodeBlocks(center, power, w, deterministicRng())

	assert.NotEmpty(t, result.DestroyedBlocks, "explosion should destroy at least one block")
	assert.InDelta(t, radius, result.BlastRadius, 0.001, "result should carry the blast radius")
	assert.Equal(t, center, result.Center, "result should carry the explosion center")
	assert.Equal(t, power, result.Power, "result should carry the explosion power")

	for _, bp := range result.DestroyedBlocks {
		dist := center.Distance(bp.ToVec3().Add(mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5}))
		assert.Less(t, dist, radius, "destroyed block %v should be within radius", bp)
		assert.Equal(t, block.Air, w.GetBlock(bp), "destroyed block %v should now be air", bp)
	}
}

func TestBlocksOutsideRadiusSurvive(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(2.0)
	radius := power * 1.5 // 3.0

	ExplodeBlocks(center, power, w, deterministicRng())

	// Check blocks well outside the radius.
	farPositions := []mcmath.BlockPos{
		{X: 0, Y: 5, Z: 0},
		{X: 15, Y: 5, Z: 15},
		{X: 0, Y: 0, Z: 0},
	}
	for _, bp := range farPositions {
		dist := center.Distance(bp.ToVec3().Add(mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5}))
		if dist >= radius {
			assert.NotEqual(t, block.Air, w.GetBlock(bp),
				"block at %v (dist %.1f) should survive explosion (radius %.1f)", bp, dist, radius)
		}
	}
}

func TestBedrockResistsExplosion(t *testing.T) {
	w := newExplosionTestWorld(0)
	bp := mcmath.BlockPos{X: 8, Y: 5, Z: 8}
	w.SetBlock(bp, block.Bedrock)

	ExplodeBlocks(mcmath.Vec3{X: 8.5, Y: 5.5, Z: 8.5}, 10, w, deterministicRng())

	assert.Equal(t, block.Bedrock, w.GetBlock(bp), "bedrock must not be destroyed by explosions")
}

func TestObsidianResistsExplosion(t *testing.T) {
	w := newExplosionTestWorld(0)
	bp := mcmath.BlockPos{X: 8, Y: 5, Z: 8}
	w.SetBlock(bp, block.Obsidian)

	ExplodeBlocks(mcmath.Vec3{X: 8.5, Y: 5.5, Z: 8.5}, 10, w, deterministicRng())

	assert.Equal(t, block.Obsidian, w.GetBlock(bp), "obsidian must not be destroyed by explosions")
}

// --- Result correctness ---

func TestExplodeBlocksResultContainsCorrectBlocks(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(3.0)

	result := ExplodeBlocks(center, power, w, deterministicRng())

	// Every block in the result should now be Air.
	for _, bp := range result.DestroyedBlocks {
		assert.Equal(t, block.Air, w.GetBlock(bp),
			"block in result at %v should be air", bp)
	}
}

func TestZeroPowerDestroysNothing(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}

	result := ExplodeBlocks(center, 0, w, deterministicRng())

	assert.Empty(t, result.DestroyedBlocks, "zero power should destroy no blocks")
	assert.Equal(t, float32(0), result.BlastRadius, "zero power should have zero blast radius")
}
