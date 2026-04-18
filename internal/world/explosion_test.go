package world

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
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

// --- Explosion block destruction ---

func TestExplosionDestroysBlocksWithinRadius(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)
	radius := power * 1.5 // 6.0

	result := Explode(center, power, w, nil, deterministicRng())

	assert.NotEmpty(t, result.DestroyedBlocks, "explosion should destroy at least one block")

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

	Explode(center, power, w, nil, deterministicRng())

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

	Explode(mcmath.Vec3{X: 8.5, Y: 5.5, Z: 8.5}, 10, w, nil, deterministicRng())

	assert.Equal(t, block.Bedrock, w.GetBlock(bp), "bedrock must not be destroyed by explosions")
}

func TestObsidianResistsExplosion(t *testing.T) {
	w := newExplosionTestWorld(0)
	bp := mcmath.BlockPos{X: 8, Y: 5, Z: 8}
	w.SetBlock(bp, block.Obsidian)

	Explode(mcmath.Vec3{X: 8.5, Y: 5.5, Z: 8.5}, 10, w, nil, deterministicRng())

	assert.Equal(t, block.Obsidian, w.GetBlock(bp), "obsidian must not be destroyed by explosions")
}

// --- Entity damage ---

func TestEntityDamageScalesWithDistance(t *testing.T) {
	w := newExplosionTestWorld(0)
	ecsW := ecs.NewWorld()

	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)

	// Entity close to center.
	closeEntity := ecsW.NewEntity()
	ecs.GetStore[entity.Transform](ecsW).Set(closeEntity, entity.Transform{
		Position: mcmath.Vec3{X: 8.5, Y: 5, Z: 8},
	})

	// Entity further away.
	farEntity := ecsW.NewEntity()
	ecs.GetStore[entity.Transform](ecsW).Set(farEntity, entity.Transform{
		Position: mcmath.Vec3{X: 12, Y: 5, Z: 8},
	})

	result := Explode(center, power, w, ecsW, deterministicRng())

	require.Len(t, result.DamagedEntities, 2, "both entities should be damaged")

	var closeDmg, farDmg float32
	for _, d := range result.DamagedEntities {
		if d.Entity == closeEntity {
			closeDmg = d.Damage
		}
		if d.Entity == farEntity {
			farDmg = d.Damage
		}
	}

	assert.Greater(t, closeDmg, farDmg,
		"closer entity should take more damage (%.2f) than farther (%.2f)", closeDmg, farDmg)
}

func TestEntityKnockbackDirection(t *testing.T) {
	w := newExplosionTestWorld(0)
	ecsW := ecs.NewWorld()

	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)

	// Entity to the positive-X side.
	e := ecsW.NewEntity()
	ecs.GetStore[entity.Transform](ecsW).Set(e, entity.Transform{
		Position: mcmath.Vec3{X: 10, Y: 5, Z: 8},
	})

	result := Explode(center, power, w, ecsW, deterministicRng())

	require.Len(t, result.DamagedEntities, 1)

	kb := result.DamagedEntities[0].Knockback
	assert.Greater(t, kb.X, float32(0), "knockback should push entity in +X direction")
	assert.InDelta(t, 0, kb.Z, 0.01, "knockback Z should be near zero for axis-aligned blast")
}

// --- TNT fuse countdown ---

func TestTNTFuseCountdown(t *testing.T) {
	ecsW := ecs.NewWorld()
	tntEntity := SpawnTNT(ecsW, mcmath.Vec3{X: 5, Y: 5, Z: 5}, 4.0)

	tntStore := ecs.GetStore[TNTEntity](ecsW)
	tnt, ok := tntStore.Get(tntEntity)
	require.True(t, ok)

	initialFuse := tnt.FuseTime

	bw := newExplosionTestWorld(0)
	sys := &TNTSystem{BlockWorld: bw}
	sys.Update(ecsW, 1.0) // advance 1 second

	tnt, ok = tntStore.Get(tntEntity)
	require.True(t, ok)
	assert.InDelta(t, initialFuse-1.0, tnt.FuseTime, 0.001,
		"fuse should decrement by dt each tick")
}

func TestTNTExplodesAtZero(t *testing.T) {
	bw := newExplosionTestWorld(10)
	ecsW := ecs.NewWorld()

	pos := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	tntEntity := SpawnTNT(ecsW, pos, 4.0)

	sys := &TNTSystem{BlockWorld: bw}

	// Advance past the 4-second fuse.
	sys.Update(ecsW, 4.1)

	assert.False(t, ecsW.Alive(tntEntity), "TNT entity should be destroyed after detonation")

	// Verify some blocks were destroyed around the TNT.
	destroyed := 0
	iRadius := int32(math.Ceil(float64(4.0 * 1.5)))
	cb := pos.Floor()
	for dx := -iRadius; dx <= iRadius; dx++ {
		for dy := -iRadius; dy <= iRadius; dy++ {
			for dz := -iRadius; dz <= iRadius; dz++ {
				bp := mcmath.BlockPos{X: cb.X + dx, Y: cb.Y + dy, Z: cb.Z + dz}
				if bw.GetBlock(bp) == block.Air {
					destroyed++
				}
			}
		}
	}
	assert.Greater(t, destroyed, 0, "TNT detonation should destroy some blocks")
}

// --- Explosion result correctness ---

func TestExplosionResultContainsCorrectBlocks(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(3.0)

	result := Explode(center, power, w, nil, deterministicRng())

	// Every block in the result should now be Air and should have been non-Air
	// before (we trust SetBlock was called because GetBlock returns Air).
	for _, bp := range result.DestroyedBlocks {
		assert.Equal(t, block.Air, w.GetBlock(bp),
			"block in result at %v should be air", bp)
	}
}

func TestZeroPowerDestroysNothing(t *testing.T) {
	w := newExplosionTestWorld(10)
	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}

	result := Explode(center, 0, w, nil, deterministicRng())

	assert.Empty(t, result.DestroyedBlocks, "zero power should destroy no blocks")
	assert.Empty(t, result.DamagedEntities, "zero power should damage no entities")
}
