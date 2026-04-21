package tick

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// Power propagation across 15 blocks
// ---------------------------------------------------------------------------

func TestPropagateRedstone_15Blocks(t *testing.T) {
	w := newStubWorld()

	// Place a lever (on) at x=0, then 15 wire blocks x=1..15.
	src := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	w.SetBlock(src, block.WithRedstonePower(block.Lever, block.MaxRedstonePower))

	for x := int32(1); x <= 15; x++ {
		w.SetBlock(mcmath.BlockPos{X: x, Y: 64, Z: 0}, block.RedstoneWire)
	}

	PropagateRedstone(w, src)

	for x := int32(1); x <= 15; x++ {
		pos := mcmath.BlockPos{X: x, Y: 64, Z: 0}
		expected := 15 - int(x)
		if expected < 0 {
			expected = 0
		}
		got := block.RedstonePowerLevel(w.GetBlock(pos))
		assert.Equal(t, expected, got, "wire at x=%d should have power %d, got %d", x, expected, got)
	}
}

// ---------------------------------------------------------------------------
// Signal decay: wire at distance 15 has power 0
// ---------------------------------------------------------------------------

func TestPropagateRedstone_SignalDecay(t *testing.T) {
	w := newStubWorld()

	src := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	w.SetBlock(src, block.WithRedstonePower(block.Lever, block.MaxRedstonePower))

	for x := int32(1); x <= 16; x++ {
		w.SetBlock(mcmath.BlockPos{X: x, Y: 64, Z: 0}, block.RedstoneWire)
	}

	PropagateRedstone(w, src)

	// Wire at x=15 should be power 0, x=16 should be power 0.
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(mcmath.BlockPos{X: 15, Y: 64, Z: 0})))
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(mcmath.BlockPos{X: 16, Y: 64, Z: 0})))
}

// ---------------------------------------------------------------------------
// No propagation past 15 blocks
// ---------------------------------------------------------------------------

func TestPropagateRedstone_NoPropagationPast15(t *testing.T) {
	w := newStubWorld()

	src := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	w.SetBlock(src, block.WithRedstonePower(block.RedstoneTorch, block.MaxRedstonePower))

	for x := int32(1); x <= 20; x++ {
		w.SetBlock(mcmath.BlockPos{X: x, Y: 64, Z: 0}, block.RedstoneWire)
	}

	PropagateRedstone(w, src)

	// Blocks at x=16..20 should all be power 0.
	for x := int32(16); x <= 20; x++ {
		pos := mcmath.BlockPos{X: x, Y: 64, Z: 0}
		assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(pos)),
			"wire at x=%d should have power 0", x)
	}
}

// ---------------------------------------------------------------------------
// Lever toggle
// ---------------------------------------------------------------------------

func TestLeverToggle(t *testing.T) {
	w := newStubWorld()

	pos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	wirePos := mcmath.BlockPos{X: 6, Y: 64, Z: 5}

	// Start with lever off.
	w.SetBlock(pos, block.Lever)
	w.SetBlock(wirePos, block.RedstoneWire)

	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(pos)))

	// Toggle on.
	w.SetBlock(pos, block.WithRedstonePower(block.Lever, block.MaxRedstonePower))
	PropagateRedstone(w, pos)

	assert.Equal(t, block.MaxRedstonePower, block.RedstonePowerLevel(w.GetBlock(pos)))
	assert.Equal(t, 14, block.RedstonePowerLevel(w.GetBlock(wirePos)))

	// Toggle off.
	w.SetBlock(pos, block.Lever)
	ResetRedstone(w, pos)

	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(pos)))
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(wirePos)))
}

// ---------------------------------------------------------------------------
// Button timeout (20 ticks)
// ---------------------------------------------------------------------------

func TestButtonTimeout(t *testing.T) {
	w := newStubWorld()
	reg := NewHandlerRegistry()
	RegisterRedstoneHandlers(reg)

	btnPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	wirePos := mcmath.BlockPos{X: 6, Y: 64, Z: 5}

	// Activate button.
	w.SetBlock(btnPos, block.WithRedstonePower(block.StoneButton, block.MaxRedstonePower))
	w.SetBlock(wirePos, block.RedstoneWire)
	PropagateRedstone(w, btnPos)

	assert.Equal(t, block.MaxRedstonePower, block.RedstonePowerLevel(w.GetBlock(btnPos)))
	assert.Equal(t, 14, block.RedstonePowerLevel(w.GetBlock(wirePos)))

	// Schedule button timeout.
	ticker := NewTickerWithSeed(w, reg, 42)
	ticker.Schedule(btnPos, block.StoneButton, buttonTimeoutTicks, 0)

	// Before timeout: still active.
	ticker.ProcessScheduledTicks(19)
	assert.Equal(t, block.MaxRedstonePower, block.RedstonePowerLevel(w.GetBlock(btnPos)))

	// At timeout: deactivated.
	ticker.ProcessScheduledTicks(20)
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(btnPos)))
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(wirePos)))
}

// ---------------------------------------------------------------------------
// Torch inversion
// ---------------------------------------------------------------------------

func TestTorchInversion(t *testing.T) {
	w := newStubWorld()

	torchPos := mcmath.BlockPos{X: 5, Y: 65, Z: 5}
	belowPos := mcmath.BlockPos{X: 5, Y: 64, Z: 5}
	wirePos := mcmath.BlockPos{X: 6, Y: 65, Z: 5}

	// Place torch on an unpowered block with wire next to it.
	w.SetBlock(belowPos, block.RedstoneWire) // unpowered wire below
	w.SetBlock(torchPos, block.WithRedstonePower(block.RedstoneTorch, block.MaxRedstonePower))
	w.SetBlock(wirePos, block.RedstoneWire)

	// Torch should stay on when block below is not powered.
	torchUpdateHandler(w, torchPos)
	assert.Equal(t, block.MaxRedstonePower, block.RedstonePowerLevel(w.GetBlock(torchPos)),
		"torch should be on when block below is unpowered")

	// Power the block below.
	w.SetBlock(belowPos, block.WithRedstonePower(block.RedstoneWire, 5))

	// Torch should turn off (inversion).
	torchUpdateHandler(w, torchPos)
	assert.Equal(t, 0, block.RedstonePowerLevel(w.GetBlock(torchPos)),
		"torch should turn off when block below is powered")
}

// ---------------------------------------------------------------------------
// RegisterRedstoneHandlers integration
// ---------------------------------------------------------------------------

func TestRegisterRedstoneHandlers(t *testing.T) {
	reg := NewHandlerRegistry()
	RegisterRedstoneHandlers(reg)

	_, ok := reg.Get(block.StoneButton)
	assert.True(t, ok, "stone button handler should be registered")

	_, ok = reg.Get(block.RedstoneTorch)
	assert.True(t, ok, "redstone torch handler should be registered")
}

// ---------------------------------------------------------------------------
// IsPowerSource tests
// ---------------------------------------------------------------------------

func TestIsPowerSource(t *testing.T) {
	assert.True(t, block.IsPowerSource(block.RedstoneTorch), "torch is always a source")
	assert.True(t, block.IsPowerSource(block.WithRedstonePower(block.RedstoneTorch, 15)))
	assert.True(t, block.IsPowerSource(block.WithRedstonePower(block.Lever, 15)))
	assert.False(t, block.IsPowerSource(block.Lever), "lever off is not a source")
	assert.True(t, block.IsPowerSource(block.WithRedstonePower(block.StoneButton, 15)))
	assert.False(t, block.IsPowerSource(block.StoneButton), "button off is not a source")
	assert.False(t, block.IsPowerSource(block.Stone), "stone is not a source")
}

// ---------------------------------------------------------------------------
// Power encoding round-trip
// ---------------------------------------------------------------------------

func TestRedstonePowerEncoding(t *testing.T) {
	for level := 0; level <= 15; level++ {
		id := block.WithRedstonePower(block.RedstoneWire, level)
		assert.Equal(t, level, block.RedstonePowerLevel(id), "level %d round-trip", level)
		assert.Equal(t, block.RedstoneWire, block.BaseID(id), "base ID preserved for level %d", level)
	}
}
