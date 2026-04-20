package tick

import (
	"math/rand/v2"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// farmlandHydrationRange is the maximum manhattan distance to search for
// water when checking farmland hydration (vanilla value is 4).
const farmlandHydrationRange int32 = 4

// cropGrowthChance is the probability (1 in N) that a wheat crop advances
// one growth stage per random tick.
const cropGrowthChance = 3

// RegisterCropHandlers registers tick handlers for WheatCrop and Farmland.
func RegisterCropHandlers(registry *HandlerRegistry) {
	registry.Register(block.WheatCrop, wheatCropHandler)
	registry.Register(block.Farmland, farmlandHandler)
}

// wheatCropHandler advances wheat crop growth by one stage with a 1/3
// chance per tick. If the block below is not farmland, the crop pops off
// (is replaced with air).
func wheatCropHandler(w BlockAccess, pos mcmath.BlockPos) {
	below := pos.Below()
	if block.BaseID(w.GetBlock(below)) != block.Farmland {
		w.SetBlock(pos, block.Air)
		return
	}

	stage := block.CropGrowthStage(w.GetBlock(pos))
	if stage >= block.CropMatureStage {
		return
	}

	if rand.IntN(cropGrowthChance) == 0 {
		w.SetBlock(pos, block.WheatCropStage(stage+1))
	}
}

// farmlandHandler checks whether farmland is hydrated by nearby water.
// If no water source is found within farmlandHydrationRange manhattan
// distance, the farmland reverts to dirt.
func farmlandHandler(w BlockAccess, pos mcmath.BlockPos) {
	if hasNearbyWater(w, pos, farmlandHydrationRange) {
		return
	}
	w.SetBlock(pos, block.Dirt)
}

// hasNearbyWater checks for a water block within the given manhattan
// distance from pos. Only the same Y level and one above/below are
// checked, matching vanilla behavior for farmland hydration.
func hasNearbyWater(w BlockAccess, center mcmath.BlockPos, maxDist int32) bool {
	for dx := -maxDist; dx <= maxDist; dx++ {
		remAfterX := maxDist - mcmath.Abs(dx)
		for dy := int32(-1); dy <= 1; dy++ {
			remAfterXY := remAfterX - mcmath.Abs(dy)
			for dz := -remAfterXY; dz <= remAfterXY; dz++ {
				if block.IsWater(w.GetBlock(center.Offset(dx, dy, dz))) {
					return true
				}
			}
		}
	}
	return false
}
