package game

import (
	"math/rand"
)

// WeatherState represents the current weather condition.
type WeatherState int

const (
	Clear   WeatherState = 0
	Rain    WeatherState = 1
	Thunder WeatherState = 2
)

const (
	minWeatherDuration = 12000
	maxWeatherDuration = 24000
	lightningRange     = 128.0
	lightningChance    = 300 // 1 in 300 chance per tick
)

// WeatherManager tracks the current weather state and handles transitions
// between weather conditions. It follows the immutable advance pattern: each
// call to Advance returns a new WeatherManager rather than mutating in place.
type WeatherManager struct {
	currentState   WeatherState
	ticksRemaining int
	transitionTick int // ticks elapsed since the last state change (for blending)
	transitionLen  int // total transition length for smooth blending
	rng            *rand.Rand
}

// NewWeatherManager returns a WeatherManager starting in Clear weather with
// a random duration between 12000 and 24000 ticks.
func NewWeatherManager() *WeatherManager {
	r := rand.New(rand.NewSource(rand.Int63()))
	return &WeatherManager{
		currentState:   Clear,
		ticksRemaining: minWeatherDuration + r.Intn(maxWeatherDuration-minWeatherDuration+1),
		transitionTick: 0,
		transitionLen:  200,
		rng:            r,
	}
}

// newWeatherManagerWithRNG creates a WeatherManager with a specific RNG (for testing).
func newWeatherManagerWithRNG(state WeatherState, remaining int, rng *rand.Rand) *WeatherManager {
	return &WeatherManager{
		currentState:   state,
		ticksRemaining: remaining,
		transitionTick: 0,
		transitionLen:  200,
		rng:            rng,
	}
}

// Advance moves the weather simulation forward by the given number of ticks
// and returns a new WeatherManager. The original is not modified.
func (wm *WeatherManager) Advance(ticks int) *WeatherManager {
	remaining := wm.ticksRemaining - ticks
	transitionTick := wm.transitionTick + ticks

	if remaining > 0 {
		return &WeatherManager{
			currentState:   wm.currentState,
			ticksRemaining: remaining,
			transitionTick: transitionTick,
			transitionLen:  wm.transitionLen,
			rng:            wm.rng,
		}
	}

	// Transition to a new state.
	nextState := nextWeatherState(wm.currentState, wm.rng)
	newDuration := minWeatherDuration + wm.rng.Intn(maxWeatherDuration-minWeatherDuration+1)

	return &WeatherManager{
		currentState:   nextState,
		ticksRemaining: newDuration,
		transitionTick: 0,
		transitionLen:  200,
		rng:            wm.rng,
	}
}

// nextWeatherState picks the next weather state based on transition
// probabilities:
//
//	Clear   -> Rain (50%)  or Clear (50%)
//	Rain    -> Thunder (30%) or Clear (70%)
//	Thunder -> Rain (60%)  or Clear (40%)
func nextWeatherState(current WeatherState, r *rand.Rand) WeatherState {
	roll := r.Float64()
	switch current {
	case Clear:
		if roll < 0.5 {
			return Rain
		}
		return Clear
	case Rain:
		if roll < 0.3 {
			return Thunder
		}
		return Clear
	case Thunder:
		if roll < 0.6 {
			return Rain
		}
		return Clear
	default:
		return Clear
	}
}

// State returns the current weather state.
func (wm *WeatherManager) State() WeatherState {
	return wm.currentState
}

// IsRaining returns true when the weather is Rain or Thunder.
func (wm *WeatherManager) IsRaining() bool {
	return wm.currentState == Rain || wm.currentState == Thunder
}

// IsThundering returns true when the weather is Thunder.
func (wm *WeatherManager) IsThundering() bool {
	return wm.currentState == Thunder
}

// SkyDarkening returns a darkening factor applied to the sky:
//
//	Clear   -> 0.0
//	Rain    -> 0.3
//	Thunder -> 0.5
func (wm *WeatherManager) SkyDarkening() float32 {
	switch wm.currentState {
	case Rain:
		return 0.3
	case Thunder:
		return 0.5
	default:
		return 0.0
	}
}

// TransitionProgress returns a value in [0, 1] representing how far through
// the current weather transition the manager is. This can be used for smooth
// visual blending between weather states.
func (wm *WeatherManager) TransitionProgress() float32 {
	if wm.transitionLen <= 0 {
		return 1.0
	}
	progress := float32(wm.transitionTick) / float32(wm.transitionLen)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// LightningStrike checks whether a lightning strike occurs this tick. During
// thunder weather there is a 1/300 chance per tick. When a strike occurs, the
// position is within 128 blocks of the given player position.
func (wm *WeatherManager) LightningStrike(playerX, playerZ float32) (x, z float32, strike bool) {
	if wm.currentState != Thunder {
		return 0, 0, false
	}

	if wm.rng.Intn(lightningChance) != 0 {
		return 0, 0, false
	}

	offsetX := (wm.rng.Float32()*2 - 1) * lightningRange
	offsetZ := (wm.rng.Float32()*2 - 1) * lightningRange

	return playerX + offsetX, playerZ + offsetZ, true
}

// String returns the human-readable name of the current weather state.
func (wm *WeatherManager) String() string {
	switch wm.currentState {
	case Clear:
		return "Clear"
	case Rain:
		return "Rain"
	case Thunder:
		return "Thunder"
	default:
		return "Clear"
	}
}
