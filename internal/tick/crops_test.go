package tick

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// WheatCrop growth tests
// ---------------------------------------------------------------------------

func TestWheatCrop_GrowsOneStage(t *testing.T) {
	w := newStubWorld()

	cropPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	w.SetBlock(cropPos.Below(), block.Farmland)
	w.SetBlock(cropPos, block.WheatCropStage(0))

	// Run the handler many times; with 1/3 chance it should advance at
	// least once in 100 attempts.
	advanced := false
	for range 100 {
		wheatCropHandler(w, cropPos)
		if block.CropGrowthStage(w.GetBlock(cropPos)) > 0 {
			advanced = true
			break
		}
	}
	assert.True(t, advanced, "crop should advance at least once in 100 ticks")
}

func TestWheatCrop_DoesNotExceedMature(t *testing.T) {
	w := newStubWorld()

	cropPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	w.SetBlock(cropPos.Below(), block.Farmland)
	w.SetBlock(cropPos, block.WheatCropStage(block.CropMatureStage))

	for range 50 {
		wheatCropHandler(w, cropPos)
	}

	stage := block.CropGrowthStage(w.GetBlock(cropPos))
	assert.Equal(t, block.CropMatureStage, stage, "mature crop should not grow further")
}

func TestWheatCrop_PopsOffWithoutFarmland(t *testing.T) {
	w := newStubWorld()

	cropPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	w.SetBlock(cropPos.Below(), block.Dirt) // not farmland
	w.SetBlock(cropPos, block.WheatCropStage(3))

	wheatCropHandler(w, cropPos)

	assert.Equal(t, block.Air, w.GetBlock(cropPos),
		"crop should pop off when block below is not farmland")
}

func TestWheatCrop_StaysOnFarmland(t *testing.T) {
	w := newStubWorld()

	cropPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	w.SetBlock(cropPos.Below(), block.Farmland)
	w.SetBlock(cropPos, block.WheatCropStage(3))

	// Even if growth doesn't happen, the crop should remain.
	wheatCropHandler(w, cropPos)

	assert.True(t, block.IsCrop(w.GetBlock(cropPos)),
		"crop should remain on farmland")
}

func TestWheatCrop_GrowthStageProgression(t *testing.T) {
	w := newStubWorld()

	cropPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	w.SetBlock(cropPos.Below(), block.Farmland)
	w.SetBlock(cropPos, block.WheatCropStage(0))

	// Force growth by calling the handler until each stage is reached.
	for expectedStage := 1; expectedStage <= block.CropMatureStage; expectedStage++ {
		for range 1000 {
			wheatCropHandler(w, cropPos)
			if block.CropGrowthStage(w.GetBlock(cropPos)) >= expectedStage {
				break
			}
		}
		assert.GreaterOrEqual(t, block.CropGrowthStage(w.GetBlock(cropPos)), expectedStage,
			"crop should reach stage %d", expectedStage)
	}

	assert.Equal(t, block.CropMatureStage, block.CropGrowthStage(w.GetBlock(cropPos)),
		"crop should reach mature stage")
}

// ---------------------------------------------------------------------------
// Farmland hydration tests
// ---------------------------------------------------------------------------

func TestFarmland_StaysHydratedNearWater(t *testing.T) {
	w := newStubWorld()

	farmPos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(farmPos, block.Farmland)
	w.SetBlock(farmPos.Offset(3, 0, 0), block.Water)

	farmlandHandler(w, farmPos)

	assert.Equal(t, block.Farmland, w.GetBlock(farmPos),
		"farmland should stay hydrated when water is within range")
}

func TestFarmland_DriesWithoutWater(t *testing.T) {
	w := newStubWorld()

	farmPos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(farmPos, block.Farmland)
	// No water anywhere nearby.

	farmlandHandler(w, farmPos)

	assert.Equal(t, block.Dirt, w.GetBlock(farmPos),
		"farmland should revert to dirt without nearby water")
}

func TestFarmland_HydrationAtMaxRange(t *testing.T) {
	w := newStubWorld()

	farmPos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(farmPos, block.Farmland)
	// Place water at exactly manhattan distance 4 (dx=4, dy=0, dz=0).
	w.SetBlock(farmPos.Offset(4, 0, 0), block.Water)

	farmlandHandler(w, farmPos)

	assert.Equal(t, block.Farmland, w.GetBlock(farmPos),
		"farmland should be hydrated at max range")
}

func TestFarmland_DriesWhenWaterTooFar(t *testing.T) {
	w := newStubWorld()

	farmPos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(farmPos, block.Farmland)
	// Place water at manhattan distance 5 (just beyond range).
	w.SetBlock(farmPos.Offset(5, 0, 0), block.Water)

	farmlandHandler(w, farmPos)

	assert.Equal(t, block.Dirt, w.GetBlock(farmPos),
		"farmland should dry when water is beyond range")
}

func TestFarmland_HydrationOneBlockAbove(t *testing.T) {
	w := newStubWorld()

	farmPos := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(farmPos, block.Farmland)
	// Place water one block above and 3 blocks away (manhattan 4).
	w.SetBlock(farmPos.Offset(3, 1, 0), block.Water)

	farmlandHandler(w, farmPos)

	assert.Equal(t, block.Farmland, w.GetBlock(farmPos),
		"farmland should detect water one block above")
}

// ---------------------------------------------------------------------------
// hasNearbyWater helper tests
// ---------------------------------------------------------------------------

func TestHasNearbyWater_FlowingWaterCounts(t *testing.T) {
	w := newStubWorld()

	center := mcmath.BlockPos{X: 10, Y: 64, Z: 10}
	w.SetBlock(center.Offset(2, 0, 0), block.FlowingWaterLevel(3))

	assert.True(t, hasNearbyWater(w, center, 4),
		"flowing water should count for hydration")
}

func TestHasNearbyWater_NoWater(t *testing.T) {
	w := newStubWorld()

	center := mcmath.BlockPos{X: 10, Y: 64, Z: 10}

	assert.False(t, hasNearbyWater(w, center, 4),
		"should return false with no water nearby")
}

// ---------------------------------------------------------------------------
// RegisterCropHandlers integration test
// ---------------------------------------------------------------------------

func TestRegisterCropHandlers_RegistersBothHandlers(t *testing.T) {
	reg := NewHandlerRegistry()
	RegisterCropHandlers(reg)

	_, ok := reg.Get(block.WheatCrop)
	assert.True(t, ok, "wheat crop handler should be registered")

	_, ok = reg.Get(block.Farmland)
	assert.True(t, ok, "farmland handler should be registered")
}

func TestRegisterDefaults_IncludesCropHandlers(t *testing.T) {
	reg := NewHandlerRegistry()
	RegisterDefaults(reg)

	_, ok := reg.Get(block.WheatCrop)
	assert.True(t, ok, "RegisterDefaults should include wheat crop handler")

	_, ok = reg.Get(block.Farmland)
	assert.True(t, ok, "RegisterDefaults should include farmland handler")
}

// ---------------------------------------------------------------------------
// Registry base-ID fallback for crops with encoded stages
// ---------------------------------------------------------------------------

func TestRegistry_LooksUpCropByBaseID(t *testing.T) {
	reg := NewHandlerRegistry()
	RegisterCropHandlers(reg)

	// WheatCropStage(3) encodes stage in upper bits — registry should
	// fall back to the base WheatCrop ID.
	_, ok := reg.Get(block.WheatCropStage(3))
	assert.True(t, ok, "registry should find handler for crop with encoded stage")
}
