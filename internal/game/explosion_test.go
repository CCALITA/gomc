//go:build !ci

package game

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/world"
)

// newExplosionTestWorld creates a flat world with a single loaded chunk at
// (0,0) filled with stone from Y=0 to Y=fillHeight-1.
func newExplosionTestWorld(fillHeight int) *world.World {
	w := world.NewWorld(0)
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

// --- Entity damage via applyExplosionDamage ---

func TestApplyExplosionDamageScalesWithDistance(t *testing.T) {
	ecsW := ecs.NewWorld()

	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)
	radius := power * 1.5

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

	result := world.BlockExplosionResult{
		Center:      center,
		Power:       power,
		BlastRadius: radius,
	}

	applyExplosionDamage(result, ecsW)

	damageStore := ecs.GetStore[entity.Damage](ecsW)

	closeDmg, closeOk := damageStore.Get(closeEntity)
	farDmg, farOk := damageStore.Get(farEntity)
	require.True(t, closeOk, "close entity should have damage")
	require.True(t, farOk, "far entity should have damage")

	assert.Greater(t, closeDmg.Amount, farDmg.Amount,
		"closer entity should take more damage (%.2f) than farther (%.2f)", closeDmg.Amount, farDmg.Amount)
}

func TestApplyExplosionDamageKnockbackDirection(t *testing.T) {
	ecsW := ecs.NewWorld()

	center := mcmath.Vec3{X: 8, Y: 5, Z: 8}
	power := float32(4.0)
	radius := power * 1.5

	// Entity to the positive-X side.
	e := ecsW.NewEntity()
	ecs.GetStore[entity.Transform](ecsW).Set(e, entity.Transform{
		Position: mcmath.Vec3{X: 10, Y: 5, Z: 8},
	})

	result := world.BlockExplosionResult{
		Center:      center,
		Power:       power,
		BlastRadius: radius,
	}

	applyExplosionDamage(result, ecsW)

	damageStore := ecs.GetStore[entity.Damage](ecsW)
	dmg, ok := damageStore.Get(e)
	require.True(t, ok)

	assert.Greater(t, dmg.Knockback.X, float32(0), "knockback should push entity in +X direction")
	assert.InDelta(t, 0, dmg.Knockback.Z, 0.01, "knockback Z should be near zero for axis-aligned blast")
}

func TestApplyExplosionDamageSkipsNilWorld(t *testing.T) {
	result := world.BlockExplosionResult{
		Center:      mcmath.Vec3{X: 8, Y: 5, Z: 8},
		Power:       4.0,
		BlastRadius: 6.0,
	}

	// Must not panic.
	applyExplosionDamage(result, nil)
}

func TestApplyExplosionDamageSkipsZeroRadius(t *testing.T) {
	ecsW := ecs.NewWorld()
	e := ecsW.NewEntity()
	ecs.GetStore[entity.Transform](ecsW).Set(e, entity.Transform{
		Position: mcmath.Vec3{X: 8, Y: 5, Z: 8},
	})

	result := world.BlockExplosionResult{
		Center:      mcmath.Vec3{X: 8, Y: 5, Z: 8},
		Power:       0,
		BlastRadius: 0,
	}

	applyExplosionDamage(result, ecsW)

	damageStore := ecs.GetStore[entity.Damage](ecsW)
	_, ok := damageStore.Get(e)
	assert.False(t, ok, "zero radius should not damage any entity")
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
