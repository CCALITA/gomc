package game

import "math"

const (
	// DayLength is the number of ticks in a full in-game day (matches Minecraft).
	DayLength = 24000

	// Tick boundaries for day/night.
	dayStart = 6000
	dayEnd   = 18000
)

// TimeKeeper tracks in-game time measured in ticks and derives sun position,
// ambient light, and sky colour from the current time of day.
type TimeKeeper struct {
	GameTime int64
}

// NewTimeKeeper returns a TimeKeeper starting at tick 0 (dawn).
func NewTimeKeeper() *TimeKeeper {
	return &TimeKeeper{GameTime: 0}
}

// Advance moves time forward by the given number of ticks.
func (tk *TimeKeeper) Advance(ticks int) *TimeKeeper {
	return &TimeKeeper{GameTime: tk.GameTime + int64(ticks)}
}

// TimeOfDay returns the current tick within the day cycle (0-23999).
func (tk *TimeKeeper) TimeOfDay() int {
	t := int(tk.GameTime % DayLength)
	if t < 0 {
		t += DayLength
	}
	return t
}

// DayCount returns the number of full days that have elapsed.
func (tk *TimeKeeper) DayCount() int {
	if tk.GameTime < 0 {
		return 0
	}
	return int(tk.GameTime / DayLength)
}

// SunAngle returns the sun's angle in radians.
// 0 = dawn (tick 0), pi/2 = noon (tick 6000), pi = dusk (tick 12000),
// 3pi/2 = midnight (tick 18000).
//
// The mapping treats tick 0 as dawn (angle 0) and proceeds linearly through
// a full 2*pi rotation over DayLength ticks.
func (tk *TimeKeeper) SunAngle() float32 {
	tod := float64(tk.TimeOfDay())
	angle := (tod / DayLength) * 2.0 * math.Pi
	return float32(angle)
}

// SunDirection returns a unit vector pointing toward the sun.
// The sun orbits in the XY plane based on SunAngle.
func (tk *TimeKeeper) SunDirection() (x, y, z float32) {
	angle := float64(tk.SunAngle())
	return float32(math.Cos(angle)), float32(math.Sin(angle)), 0
}

// AmbientLevel returns the ambient light multiplier (0.2 at midnight, 1.0 at noon).
// Uses smooth sine interpolation based on the sun angle.
func (tk *TimeKeeper) AmbientLevel() float32 {
	angle := float64(tk.SunAngle())
	// sin(angle) ranges from -1 (midnight) to +1 (noon).
	// Map [-1,1] to [0.2, 1.0] => level = 0.6 + 0.4 * sin(angle)
	sinVal := math.Sin(angle)
	return float32(0.6 + 0.4*sinVal)
}

// SkyColor returns the sky colour as (r, g, b) in [0,1].
//
//   - Noon (sin~1):   blue   (0.4, 0.6, 1.0)
//   - Dawn/Dusk:      orange (0.9, 0.5, 0.2)
//   - Midnight:       dark   (0.05, 0.05, 0.15)
//
// The colour is interpolated between these three based on the sun's sine value.
func (tk *TimeKeeper) SkyColor() (r, g, b float32) {
	angle := float64(tk.SunAngle())
	sinVal := math.Sin(angle) // -1 (midnight) to +1 (noon)

	// Day colours (blue sky).
	dayR, dayG, dayB := 0.4, 0.6, 1.0
	// Dawn/dusk colours (orange).
	horizR, horizG, horizB := 0.9, 0.5, 0.2
	// Night colours (dark blue).
	nightR, nightG, nightB := 0.05, 0.05, 0.15

	if sinVal >= 0 {
		// Between horizon (sin=0) and noon (sin=1).
		t := sinVal
		r = float32(lerp(horizR, dayR, t))
		g = float32(lerp(horizG, dayG, t))
		b = float32(lerp(horizB, dayB, t))
	} else {
		// Between horizon (sin=0) and midnight (sin=-1).
		t := -sinVal
		r = float32(lerp(horizR, nightR, t))
		g = float32(lerp(horizG, nightG, t))
		b = float32(lerp(horizB, nightB, t))
	}
	return r, g, b
}

// IsDay returns true when the time of day is between 6000 and 18000 ticks
// (inclusive of 6000, exclusive of 18000).
func (tk *TimeKeeper) IsDay() bool {
	tod := tk.TimeOfDay()
	return tod >= dayStart && tod < dayEnd
}

// IsNight returns true when it is not day.
func (tk *TimeKeeper) IsNight() bool {
	return !tk.IsDay()
}

// lerp linearly interpolates between a and b by t in [0,1].
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
