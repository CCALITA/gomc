package world

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// BlockExplosionResult captures the blocks destroyed by a single call to
// ExplodeBlocks. It carries enough context (center, power, blast radius) for
// a higher-level system to apply entity damage without re-computing these
// values.
type BlockExplosionResult struct {
	Center          mcmath.Vec3
	Power           float32
	BlastRadius     float32
	DestroyedBlocks []mcmath.BlockPos
}

// blastResistant reports whether a block type is immune to explosions.
func blastResistant(id uint16) bool {
	base := block.BaseID(id)
	return base == block.Bedrock || base == block.Obsidian
}

// ExplodeBlocks processes the block-destruction phase of an explosion centered
// at center with the given power. Blocks within the blast radius are
// probabilistically destroyed (except for bedrock and obsidian). A
// deterministic *rand.Rand may be supplied for testing; if rng is nil a default
// source is used.
func ExplodeBlocks(center mcmath.Vec3, power float32, w *World, rng *rand.Rand) BlockExplosionResult {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	radius := power * 1.5
	result := BlockExplosionResult{
		Center:      center,
		Power:       power,
		BlastRadius: radius,
	}

	if radius <= 0 {
		return result
	}

	iRadius := int32(math.Ceil(float64(radius)))
	centerBlock := center.Floor()
	blockCenterOffset := mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5}

	for dx := -iRadius; dx <= iRadius; dx++ {
		for dy := -iRadius; dy <= iRadius; dy++ {
			for dz := -iRadius; dz <= iRadius; dz++ {
				bp := mcmath.BlockPos{
					X: centerBlock.X + dx,
					Y: centerBlock.Y + dy,
					Z: centerBlock.Z + dz,
				}

				dist := center.Distance(bp.ToVec3().Add(blockCenterOffset))
				if dist >= radius {
					continue
				}

				id := w.GetBlock(bp)
				if id == block.Air {
					continue
				}
				if blastResistant(id) {
					continue
				}

				attenuation := 1 - dist/radius
				if rng.Float32() < attenuation {
					w.SetBlock(bp, block.Air)
					result.DestroyedBlocks = append(result.DestroyedBlocks, bp)
				}
			}
		}
	}

	return result
}
