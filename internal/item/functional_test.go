package item

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompassAngle_North(t *testing.T) {
	angle := CompassAngle(0, 0, 0, -10)
	assert.InDelta(t, math.Pi, math.Abs(float64(angle)), 0.001)
}

func TestCompassAngle_South(t *testing.T) {
	angle := CompassAngle(0, 0, 0, 10)
	assert.InDelta(t, 0.0, float64(angle), 0.001)
}

func TestCompassAngle_East(t *testing.T) {
	angle := CompassAngle(0, 0, 10, 0)
	assert.InDelta(t, math.Pi/2, float64(angle), 0.001)
}

func TestCompassAngle_West(t *testing.T) {
	angle := CompassAngle(0, 0, -10, 0)
	assert.InDelta(t, -math.Pi/2, float64(angle), 0.001)
}

func TestCompassAngle_SamePosition(t *testing.T) {
	angle := CompassAngle(5, 5, 5, 5)
	assert.InDelta(t, 0.0, float64(angle), 0.001)
}

func TestClockPhase_Noon(t *testing.T) {
	assert.InDelta(t, 0.0, float64(ClockPhase(6000)), 0.001)
}

func TestClockPhase_Midnight(t *testing.T) {
	assert.InDelta(t, 0.5, float64(ClockPhase(18000)), 0.001)
}

func TestClockPhase_Dawn(t *testing.T) {
	assert.InDelta(t, 0.75, float64(ClockPhase(0)), 0.001)
}

func TestClockPhase_Dusk(t *testing.T) {
	assert.InDelta(t, 0.25, float64(ClockPhase(12000)), 0.001)
}

func TestClockPhase_NegativeTick(t *testing.T) {
	assert.InDelta(t, 0.5, float64(ClockPhase(-6000)), 0.001)
}

func TestTickToHoursMinutes_Dawn(t *testing.T) {
	h, m := TickToHoursMinutes(0)
	assert.Equal(t, 6, h)
	assert.Equal(t, 0, m)
}

func TestTickToHoursMinutes_Noon(t *testing.T) {
	h, m := TickToHoursMinutes(6000)
	assert.Equal(t, 12, h)
	assert.Equal(t, 0, m)
}

func TestTickToHoursMinutes_Dusk(t *testing.T) {
	h, m := TickToHoursMinutes(12000)
	assert.Equal(t, 18, h)
	assert.Equal(t, 0, m)
}

func TestTickToHoursMinutes_Midnight(t *testing.T) {
	h, m := TickToHoursMinutes(18000)
	assert.Equal(t, 0, h)
	assert.Equal(t, 0, m)
}

func TestTickToHoursMinutes_MidMorning(t *testing.T) {
	h, m := TickToHoursMinutes(3000)
	assert.Equal(t, 9, h)
	assert.Equal(t, 0, m)
}

func TestTickIsDaytime(t *testing.T) {
	assert.True(t, TickIsDaytime(0))
	assert.True(t, TickIsDaytime(6000))
	assert.True(t, TickIsDaytime(11999))
	assert.False(t, TickIsDaytime(12000))
	assert.False(t, TickIsDaytime(18000))
	assert.False(t, TickIsDaytime(23999))
}

func TestCompassProperties(t *testing.T) {
	p := GetProperties(Compass)
	assert.Equal(t, "Compass", p.Name)
	assert.True(t, p.IsCompass)
	assert.False(t, p.IsClock)
}

func TestClockProperties(t *testing.T) {
	p := GetProperties(Clock)
	assert.Equal(t, "Clock", p.Name)
	assert.True(t, p.IsClock)
	assert.False(t, p.IsCompass)
}

func TestRedstoneProperties(t *testing.T) {
	p := GetProperties(RedstoneItem)
	assert.Equal(t, "Redstone", p.Name)
	assert.Equal(t, DefaultMaxStackSize, p.MaxStackSize)
}
