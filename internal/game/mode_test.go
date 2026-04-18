//go:build !ci

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModeManager_DefaultsSurvival(t *testing.T) {
	mm := NewModeManager()
	assert.Equal(t, ModeSurvival, mm.CurrentMode)
}

func TestSurvivalProperties(t *testing.T) {
	mm := NewModeManager()

	assert.True(t, mm.CanTakeDamage(), "survival: should take damage")
	assert.False(t, mm.CanBreakInstantly(), "survival: should not break instantly")
	assert.False(t, mm.HasInfiniteItems(), "survival: should not have infinite items")
	assert.False(t, mm.CanFly(), "survival: should not fly")
	assert.False(t, mm.IsNoClip(), "survival: should not noclip")
}

func TestCreativeProperties(t *testing.T) {
	mm := NewModeManager()
	mm.SetMode(ModeCreative)

	assert.Equal(t, ModeCreative, mm.CurrentMode)
	assert.False(t, mm.CanTakeDamage(), "creative: should not take damage")
	assert.True(t, mm.CanBreakInstantly(), "creative: should break instantly")
	assert.True(t, mm.HasInfiniteItems(), "creative: should have infinite items")
	assert.True(t, mm.CanFly(), "creative: should fly")
	assert.False(t, mm.IsNoClip(), "creative: should not noclip")
}

func TestSpectatorProperties(t *testing.T) {
	mm := NewModeManager()
	mm.SetMode(ModeSpectator)

	assert.Equal(t, ModeSpectator, mm.CurrentMode)
	assert.False(t, mm.CanTakeDamage(), "spectator: should not take damage")
	assert.False(t, mm.CanBreakInstantly(), "spectator: should not break instantly")
	assert.False(t, mm.HasInfiniteItems(), "spectator: should not have infinite items")
	assert.True(t, mm.CanFly(), "spectator: should fly")
	assert.True(t, mm.IsNoClip(), "spectator: should noclip")
}

func TestModeSwitching(t *testing.T) {
	mm := NewModeManager()

	// Survival -> Creative
	mm.SetMode(ModeCreative)
	assert.Equal(t, ModeCreative, mm.CurrentMode)
	assert.True(t, mm.CanFly())

	// Creative -> Spectator
	mm.SetMode(ModeSpectator)
	assert.Equal(t, ModeSpectator, mm.CurrentMode)
	assert.True(t, mm.IsNoClip())

	// Spectator -> Survival
	mm.SetMode(ModeSurvival)
	assert.Equal(t, ModeSurvival, mm.CurrentMode)
	assert.True(t, mm.CanTakeDamage())
	assert.False(t, mm.CanFly())
}

func TestGameModeConstants(t *testing.T) {
	assert.Equal(t, GameMode(0), ModeSurvival)
	assert.Equal(t, GameMode(1), ModeCreative)
	assert.Equal(t, GameMode(2), ModeSpectator)
}

func TestUnknownModeFallsBackToSurvival(t *testing.T) {
	mm := &ModeManager{CurrentMode: GameMode(99)}

	assert.True(t, mm.CanTakeDamage(), "unknown mode: should fall back to survival damage")
	assert.False(t, mm.CanFly(), "unknown mode: should fall back to survival fly")
	assert.False(t, mm.IsNoClip(), "unknown mode: should fall back to survival noclip")
}

func TestModeManagerImplementsModeChecker(t *testing.T) {
	// ModeChecker is defined in the player package. We verify the method
	// signatures match by calling each method. If the interface in player
	// changes and ModeManager no longer satisfies it, the compiler catches it
	// where ModeManager is assigned to a ModeChecker variable (in game.go).
	mm := NewModeManager()
	_ = mm.CanFly()
	_ = mm.CanBreakInstantly()
	_ = mm.HasInfiniteItems()
	_ = mm.CanTakeDamage()
	_ = mm.IsNoClip()
}
