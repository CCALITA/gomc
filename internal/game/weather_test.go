//go:build !ci

package game

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWeatherManager_InitialStateClear(t *testing.T) {
	wm := NewWeatherManager()
	assert.Equal(t, Clear, wm.State())
}

func TestNewWeatherManager_TicksInRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		wm := NewWeatherManager()
		assert.GreaterOrEqual(t, wm.ticksRemaining, minWeatherDuration)
		assert.LessOrEqual(t, wm.ticksRemaining, maxWeatherDuration)
	}
}

func TestAdvance_DecrementsTicks(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 1000, r)

	wm2 := wm.Advance(100)
	assert.Equal(t, 900, wm2.ticksRemaining)
}

func TestAdvance_Immutability(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 1000, r)

	wm2 := wm.Advance(100)

	assert.Equal(t, 1000, wm.ticksRemaining, "original should be unchanged")
	assert.Equal(t, 900, wm2.ticksRemaining, "new instance should be decremented")
	assert.NotSame(t, wm, wm2)
}

func TestAdvance_TransitionAtZero(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 10, r)

	wm2 := wm.Advance(10)
	// Remaining hit zero, so a transition should occur and new duration assigned.
	assert.GreaterOrEqual(t, wm2.ticksRemaining, minWeatherDuration)
	assert.LessOrEqual(t, wm2.ticksRemaining, maxWeatherDuration)
}

func TestAdvance_TransitionResetsTransitionTick(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 10, r)

	wm2 := wm.Advance(10)
	assert.Equal(t, 0, wm2.transitionTick, "transition tick should reset after state change")
}

func TestRain_IsRaining(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Rain, 1000, r)

	assert.True(t, wm.IsRaining())
	assert.False(t, wm.IsThundering())
}

func TestThunder_IsRainingAndThundering(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Thunder, 1000, r)

	assert.True(t, wm.IsRaining())
	assert.True(t, wm.IsThundering())
}

func TestClear_IsNotRaining(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 1000, r)

	assert.False(t, wm.IsRaining())
	assert.False(t, wm.IsThundering())
}

func TestSkyDarkening_Values(t *testing.T) {
	tests := []struct {
		name     string
		state    WeatherState
		expected float32
	}{
		{"clear", Clear, 0.0},
		{"rain", Rain, 0.3},
		{"thunder", Thunder, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(42))
			wm := newWeatherManagerWithRNG(tt.state, 1000, r)
			assert.InDelta(t, tt.expected, wm.SkyDarkening(), 0.001)
		})
	}
}

func TestLightning_OnlyDuringThunder(t *testing.T) {
	tests := []struct {
		name  string
		state WeatherState
	}{
		{"clear", Clear},
		{"rain", Rain},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(42))
			wm := newWeatherManagerWithRNG(tt.state, 1000, r)
			// Try many ticks; no strikes should occur.
			for i := 0; i < 1000; i++ {
				_, _, strike := wm.LightningStrike(0, 0)
				assert.False(t, strike, "lightning should not occur during %s", tt.name)
			}
		})
	}
}

func TestLightning_NearPlayer(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	wm := newWeatherManagerWithRNG(Thunder, 1000, r)

	playerX := float32(100.0)
	playerZ := float32(200.0)

	// Run enough ticks to get at least one strike.
	gotStrike := false
	for i := 0; i < 10000; i++ {
		x, z, strike := wm.LightningStrike(playerX, playerZ)
		if strike {
			gotStrike = true
			dx := float64(x - playerX)
			dz := float64(z - playerZ)
			dist := math.Sqrt(dx*dx + dz*dz)
			assert.LessOrEqual(t, dist, lightningRange*math.Sqrt(2)+0.1,
				"lightning should be within range of player")
			break
		}
	}
	assert.True(t, gotStrike, "expected at least one lightning strike in 10000 attempts")
}

func TestTransitionProgress(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Clear, 1000, r)

	// At start, transition progress is 0.
	assert.InDelta(t, 0.0, wm.TransitionProgress(), 0.001)

	// After 100 ticks (half the 200-tick transition length).
	wm2 := wm.Advance(100)
	assert.InDelta(t, 0.5, wm2.TransitionProgress(), 0.001)

	// After 200 ticks (full transition).
	wm3 := wm.Advance(200)
	assert.InDelta(t, 1.0, wm3.TransitionProgress(), 0.001)

	// Beyond transition length, clamped to 1.0.
	wm4 := wm.Advance(500)
	assert.InDelta(t, 1.0, wm4.TransitionProgress(), 0.001)
}

func TestTransitions_ProduceVariety(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	wm := newWeatherManagerWithRNG(Clear, 1, r)

	seen := map[WeatherState]bool{}
	for i := 0; i < 200; i++ {
		wm = wm.Advance(wm.ticksRemaining)
		seen[wm.State()] = true
	}

	assert.True(t, seen[Clear], "should see Clear over many cycles")
	assert.True(t, seen[Rain], "should see Rain over many cycles")
	assert.True(t, seen[Thunder], "should see Thunder over many cycles")
}

func TestString_Representation(t *testing.T) {
	tests := []struct {
		state    WeatherState
		expected string
	}{
		{Clear, "Clear"},
		{Rain, "Rain"},
		{Thunder, "Thunder"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			r := rand.New(rand.NewSource(42))
			wm := newWeatherManagerWithRNG(tt.state, 1000, r)
			assert.Equal(t, tt.expected, wm.String())
		})
	}
}

func TestWeatherState_Constants(t *testing.T) {
	assert.Equal(t, WeatherState(0), Clear)
	assert.Equal(t, WeatherState(1), Rain)
	assert.Equal(t, WeatherState(2), Thunder)
}

func TestAdvance_NoTransitionWhenTicksRemain(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	wm := newWeatherManagerWithRNG(Rain, 5000, r)

	wm2 := wm.Advance(100)
	assert.Equal(t, Rain, wm2.State(), "state should not change when ticks remain")
}
