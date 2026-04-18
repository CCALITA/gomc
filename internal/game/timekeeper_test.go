//go:build !ci

package game

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeKeeper(t *testing.T) {
	tk := NewTimeKeeper()
	assert.Equal(t, int64(0), tk.GameTime)
	assert.Equal(t, 0, tk.TimeOfDay())
	assert.Equal(t, 0, tk.DayCount())
}

func TestAdvance_ReturnsNewInstance(t *testing.T) {
	tk := NewTimeKeeper()
	tk2 := tk.Advance(100)

	assert.Equal(t, int64(0), tk.GameTime, "original should be unchanged")
	assert.Equal(t, int64(100), tk2.GameTime)
}

func TestAdvance_WrapsAtDayLength(t *testing.T) {
	tk := NewTimeKeeper()
	tk = tk.Advance(DayLength + 500)

	assert.Equal(t, 500, tk.TimeOfDay())
	assert.Equal(t, 1, tk.DayCount())
}

func TestTimeOfDay_WrapsBeyondDayLength(t *testing.T) {
	tests := []struct {
		name     string
		ticks    int
		expected int
	}{
		{"zero", 0, 0},
		{"mid-day", 12000, 12000},
		{"one full day", DayLength, 0},
		{"wrap around", DayLength + 1000, 1000},
		{"two days plus", 2*DayLength + 500, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewTimeKeeper().Advance(tt.ticks)
			assert.Equal(t, tt.expected, tk.TimeOfDay())
		})
	}
}

func TestDayCount(t *testing.T) {
	tests := []struct {
		name     string
		ticks    int
		expected int
	}{
		{"zero", 0, 0},
		{"almost one day", DayLength - 1, 0},
		{"exactly one day", DayLength, 1},
		{"three days", 3 * DayLength, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewTimeKeeper().Advance(tt.ticks)
			assert.Equal(t, tt.expected, tk.DayCount())
		})
	}
}

func TestSunAngle_Noon(t *testing.T) {
	// Noon is at tick 6000 => angle = (6000/24000) * 2*pi = pi/2
	tk := NewTimeKeeper().Advance(6000)
	expected := float32(math.Pi / 2)
	assert.InDelta(t, expected, tk.SunAngle(), 0.001)
}

func TestSunAngle_Midnight(t *testing.T) {
	// Midnight is at tick 18000 => angle = (18000/24000) * 2*pi = 3*pi/2
	tk := NewTimeKeeper().Advance(18000)
	expected := float32(3 * math.Pi / 2)
	assert.InDelta(t, expected, tk.SunAngle(), 0.001)
}

func TestSunAngle_Dawn(t *testing.T) {
	// Dawn is at tick 0 => angle = 0
	tk := NewTimeKeeper()
	assert.InDelta(t, float32(0), tk.SunAngle(), 0.001)
}

func TestSunAngle_Dusk(t *testing.T) {
	// Dusk is at tick 12000 => angle = pi
	tk := NewTimeKeeper().Advance(12000)
	expected := float32(math.Pi)
	assert.InDelta(t, expected, tk.SunAngle(), 0.001)
}

func TestAmbientLevel_Noon(t *testing.T) {
	// At noon (tick 6000), sin(pi/2) = 1, level = 0.6 + 0.4*1 = 1.0
	tk := NewTimeKeeper().Advance(6000)
	assert.InDelta(t, float32(1.0), tk.AmbientLevel(), 0.01)
}

func TestAmbientLevel_Midnight(t *testing.T) {
	// At midnight (tick 18000), sin(3*pi/2) = -1, level = 0.6 + 0.4*(-1) = 0.2
	tk := NewTimeKeeper().Advance(18000)
	assert.InDelta(t, float32(0.2), tk.AmbientLevel(), 0.01)
}

func TestAmbientLevel_DawnDusk(t *testing.T) {
	// At dawn (tick 0) and dusk (tick 12000), sin=0, level = 0.6
	tkDawn := NewTimeKeeper()
	assert.InDelta(t, float32(0.6), tkDawn.AmbientLevel(), 0.01)

	tkDusk := NewTimeKeeper().Advance(12000)
	assert.InDelta(t, float32(0.6), tkDusk.AmbientLevel(), 0.01)
}

func TestSkyColor_Noon(t *testing.T) {
	// At noon, sky should be blue (0.4, 0.6, 1.0)
	tk := NewTimeKeeper().Advance(6000)
	r, g, b := tk.SkyColor()
	assert.InDelta(t, 0.4, r, 0.05, "noon red")
	assert.InDelta(t, 0.6, g, 0.05, "noon green")
	assert.InDelta(t, 1.0, b, 0.05, "noon blue")
}

func TestSkyColor_Midnight(t *testing.T) {
	// At midnight, sky should be dark (0.05, 0.05, 0.15)
	tk := NewTimeKeeper().Advance(18000)
	r, g, b := tk.SkyColor()
	assert.InDelta(t, 0.05, r, 0.05, "midnight red")
	assert.InDelta(t, 0.05, g, 0.05, "midnight green")
	assert.InDelta(t, 0.15, b, 0.05, "midnight blue")
}

func TestSkyColor_DawnDusk(t *testing.T) {
	// At dawn/dusk, sky should be orange-ish (0.9, 0.5, 0.2)
	tk := NewTimeKeeper()
	r, g, b := tk.SkyColor()
	assert.InDelta(t, 0.9, r, 0.05, "dawn red")
	assert.InDelta(t, 0.5, g, 0.05, "dawn green")
	assert.InDelta(t, 0.2, b, 0.05, "dawn blue")
}

func TestIsDay(t *testing.T) {
	tests := []struct {
		name     string
		ticks    int
		expected bool
	}{
		{"dawn boundary", 6000, true},
		{"noon", 12000, true},
		{"just before dusk", 17999, true},
		{"dusk boundary", 18000, false},
		{"midnight", 0, false},
		{"late night", 23000, false},
		{"early morning", 5000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewTimeKeeper().Advance(tt.ticks)
			assert.Equal(t, tt.expected, tk.IsDay())
		})
	}
}

func TestIsNight(t *testing.T) {
	tests := []struct {
		name     string
		ticks    int
		expected bool
	}{
		{"midnight", 0, true},
		{"noon", 12000, false},
		{"dawn boundary", 6000, false},
		{"dusk boundary", 18000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewTimeKeeper().Advance(tt.ticks)
			assert.Equal(t, tt.expected, tk.IsNight())
		})
	}
}

func TestSunDirection_DuringDay(t *testing.T) {
	// At noon (tick 6000), sun angle is pi/2, so direction = (cos(pi/2), sin(pi/2), 0) = (0, 1, 0)
	tk := NewTimeKeeper().Advance(6000)
	x, y, z := tk.SunDirection()
	assert.InDelta(t, 0.0, x, 0.01, "noon x")
	assert.InDelta(t, 1.0, y, 0.01, "noon y should be > 0 during day")
	assert.InDelta(t, 0.0, z, 0.01, "z always 0")
}

func TestSunDirection_YPositiveDuringDay(t *testing.T) {
	// Check multiple daytime ticks to confirm Y > 0
	for ticks := 1; ticks < 12000; ticks += 1000 {
		tk := NewTimeKeeper().Advance(ticks)
		_, y, _ := tk.SunDirection()
		assert.Greater(t, y, float32(0),
			"sun Y should be > 0 at tick %d (daytime in angle terms)", ticks)
	}
}

func TestSunDirection_Midnight(t *testing.T) {
	// At midnight (tick 18000), angle = 3*pi/2, direction = (cos(3pi/2), sin(3pi/2), 0) = (0, -1, 0)
	tk := NewTimeKeeper().Advance(18000)
	x, y, z := tk.SunDirection()
	assert.InDelta(t, 0.0, x, 0.01, "midnight x")
	assert.InDelta(t, -1.0, y, 0.01, "midnight y")
	assert.InDelta(t, 0.0, z, 0.01, "midnight z")
}

func TestSunDirection_IsUnitVector(t *testing.T) {
	for ticks := 0; ticks < DayLength; ticks += 1000 {
		tk := NewTimeKeeper().Advance(ticks)
		x, y, z := tk.SunDirection()
		length := math.Sqrt(float64(x*x + y*y + z*z))
		assert.InDelta(t, 1.0, length, 0.001,
			"sun direction should be unit vector at tick %d", ticks)
	}
}
