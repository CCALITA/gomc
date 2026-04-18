package block

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// BlockState -- encode/decode round-trips
// ---------------------------------------------------------------------------

func TestNewBlockState_RoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		id          BlockID
		orientation Orientation
		waterlogged bool
	}{
		{"air default", Air, OrientNorth, false},
		{"stone south", Stone, OrientSouth, false},
		{"oak_log east waterlogged", OakLog, OrientEast, true},
		{"glass west", Glass, OrientWest, false},
		{"torch up waterlogged", Torch, OrientUp, true},
		{"bedrock down", Bedrock, OrientDown, false},
		{"glowstone north waterlogged", Glowstone, OrientNorth, true},
		{"emerald_ore south", EmeraldOre, OrientSouth, false},
		{"max block id", BlockID(0xFFFF), OrientDown, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewBlockState(tc.id, tc.orientation, tc.waterlogged)
			assert.Equal(t, tc.id, s.StateBlockID(), "block ID round-trip")
			assert.Equal(t, tc.orientation, s.StateOrientation(), "orientation round-trip")
			assert.Equal(t, tc.waterlogged, s.StateWaterlogged(), "waterlogged round-trip")
			assert.False(t, s.StatePowered(), "powered should default to false")
		})
	}
}

func TestBlockState_WithPowered(t *testing.T) {
	s := NewBlockState(Stone, OrientEast, false)
	assert.False(t, s.StatePowered())

	powered := s.WithPowered(true)
	assert.True(t, powered.StatePowered())
	// Other fields unchanged
	assert.Equal(t, Stone, powered.StateBlockID())
	assert.Equal(t, OrientEast, powered.StateOrientation())
	assert.False(t, powered.StateWaterlogged())

	// Toggle back
	unpowered := powered.WithPowered(false)
	assert.False(t, unpowered.StatePowered())
	assert.Equal(t, s, unpowered)
}

func TestBlockState_WithOrientation(t *testing.T) {
	s := NewBlockState(OakLog, OrientNorth, true)
	assert.Equal(t, OrientNorth, s.StateOrientation())

	rotated := s.WithOrientation(OrientUp)
	assert.Equal(t, OrientUp, rotated.StateOrientation())
	assert.Equal(t, OakLog, rotated.StateBlockID())
	assert.True(t, rotated.StateWaterlogged())
}

func TestBlockState_WithWaterlogged(t *testing.T) {
	s := NewBlockState(Glass, OrientSouth, false)
	assert.False(t, s.StateWaterlogged())

	wet := s.WithWaterlogged(true)
	assert.True(t, wet.StateWaterlogged())
	assert.Equal(t, Glass, wet.StateBlockID())
	assert.Equal(t, OrientSouth, wet.StateOrientation())

	dry := wet.WithWaterlogged(false)
	assert.False(t, dry.StateWaterlogged())
	assert.Equal(t, s, dry)
}

func TestBlockState_OrientationClamp(t *testing.T) {
	// Orientations > 5 should clamp to 0 (North)
	s := NewBlockState(Stone, Orientation(10), false)
	assert.Equal(t, OrientNorth, s.StateOrientation())

	rotated := s.WithOrientation(Orientation(255))
	assert.Equal(t, OrientNorth, rotated.StateOrientation())
}

func TestBlockState_AllOrientations(t *testing.T) {
	orientations := []Orientation{
		OrientNorth, OrientSouth, OrientEast,
		OrientWest, OrientUp, OrientDown,
	}
	for _, o := range orientations {
		s := NewBlockState(Stone, o, false)
		assert.Equal(t, o, s.StateOrientation(), "orientation %d", o)
	}
}

func TestBlockState_AllBitsCombined(t *testing.T) {
	// Set every field and verify they don't interfere
	s := NewBlockState(DiamondBlock, OrientDown, true)
	s = s.WithPowered(true)

	assert.Equal(t, DiamondBlock, s.StateBlockID())
	assert.Equal(t, OrientDown, s.StateOrientation())
	assert.True(t, s.StateWaterlogged())
	assert.True(t, s.StatePowered())
}

func TestBlockState_ZeroValueIsAirNorthDry(t *testing.T) {
	var s BlockState
	assert.Equal(t, Air, s.StateBlockID())
	assert.Equal(t, OrientNorth, s.StateOrientation())
	assert.False(t, s.StateWaterlogged())
	assert.False(t, s.StatePowered())
}

func TestBlockState_Immutability(t *testing.T) {
	original := NewBlockState(Stone, OrientEast, false)

	// Mutations return new values, original is unchanged
	_ = original.WithPowered(true)
	_ = original.WithWaterlogged(true)
	_ = original.WithOrientation(OrientDown)

	assert.Equal(t, Stone, original.StateBlockID())
	assert.Equal(t, OrientEast, original.StateOrientation())
	assert.False(t, original.StateWaterlogged())
	assert.False(t, original.StatePowered())
}

func TestBlockState_DistinctIDs(t *testing.T) {
	// Ensure different block IDs produce different states even with same flags
	s1 := NewBlockState(Stone, OrientNorth, false)
	s2 := NewBlockState(Dirt, OrientNorth, false)
	assert.NotEqual(t, s1, s2)
	assert.Equal(t, Stone, s1.StateBlockID())
	assert.Equal(t, Dirt, s2.StateBlockID())
}

func TestBlockState_WoolIDs(t *testing.T) {
	// Verify wool block IDs round-trip through BlockState
	wools := []struct {
		id   BlockID
		name string
	}{
		{WhiteWool, "white"},
		{OrangeWool, "orange"},
		{MagentaWool, "magenta"},
		{LightBlueWool, "light_blue"},
		{YellowWool, "yellow"},
		{LimeWool, "lime"},
		{PinkWool, "pink"},
		{GrayWool, "gray"},
		{LightGrayWool, "light_gray"},
		{CyanWool, "cyan"},
		{PurpleWool, "purple"},
		{BlueWool, "blue"},
		{BrownWool, "brown"},
		{GreenWool, "green"},
		{RedWool, "red"},
		{BlackWool, "black"},
	}
	for _, w := range wools {
		t.Run(w.name, func(t *testing.T) {
			s := NewBlockState(w.id, OrientUp, true)
			s = s.WithPowered(true)
			assert.Equal(t, w.id, s.StateBlockID())
			assert.Equal(t, OrientUp, s.StateOrientation())
			assert.True(t, s.StateWaterlogged())
			assert.True(t, s.StatePowered())
		})
	}
}

func TestBlockState_WoodVariantIDs(t *testing.T) {
	woods := []BlockID{
		BirchLog, BirchPlanks, BirchLeaves,
		SpruceLog, SprucePlanks, SpruceLeaves,
		JungleLog, JunglePlanks, JungleLeaves,
		DarkOakLog, DarkOakPlanks, DarkOakLeaves,
		AcaciaLog, AcaciaPlanks, AcaciaLeaves,
	}
	for _, id := range woods {
		s := NewBlockState(id, OrientSouth, false)
		assert.Equal(t, id, s.StateBlockID(), "wood variant %d", id)
	}
}

func TestBlockState_FunctionalBlockIDs(t *testing.T) {
	blocks := []BlockID{
		IronBlock, GoldBlock, DiamondBlock, Bookshelf, TNT,
		Rail, Ladder, Farmland, WheatCrop, Sugarcane,
		Cactus, Clay, Bricks, NetherRack, SoulSand,
		Glowstone, EndStone, Snow, Ice, Sponge,
		RedstoneOre, LapisOre, EmeraldOre,
	}
	for _, id := range blocks {
		s := NewBlockState(id, OrientWest, true)
		s = s.WithPowered(true)
		assert.Equal(t, id, s.StateBlockID(), "functional block %d", id)
		assert.Equal(t, OrientWest, s.StateOrientation())
		assert.True(t, s.StateWaterlogged())
		assert.True(t, s.StatePowered())
	}
}
