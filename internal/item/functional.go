package item

import "math"

// TicksPerDay is the number of game ticks in a full Minecraft day cycle.
const TicksPerDay int64 = 24000

// CompassAngle returns the angle in radians from the player position to the
// spawn point, projected onto the XZ plane. The angle is measured clockwise
// from the positive Z axis (south), matching Minecraft conventions.
func CompassAngle(playerX, playerZ, spawnX, spawnZ float32) float32 {
	dx := float64(spawnX - playerX)
	dz := float64(spawnZ - playerZ)
	return float32(math.Atan2(dx, dz))
}

// ClockPhase returns a value in [0.0, 1.0) representing the current position
// in the day/night cycle. 0.0 = noon, 0.5 = midnight, wrapping back to 1.0 = noon.
// Based on a 24000-tick day cycle where tick 0 = 6:00 AM and tick 6000 = noon.
func ClockPhase(gameTick int64) float32 {
	t := gameTick % TicksPerDay
	if t < 0 {
		t += TicksPerDay
	}
	shifted := (t + TicksPerDay - 6000) % TicksPerDay
	return float32(shifted) / float32(TicksPerDay)
}

// TickToHoursMinutes converts a game tick to 24-hour clock hours and minutes.
// Tick 0 = 6:00, tick 6000 = 12:00 (noon), tick 12000 = 18:00, tick 18000 = 0:00 (midnight).
func TickToHoursMinutes(gameTick int64) (hour, minute int) {
	t := gameTick % TicksPerDay
	if t < 0 {
		t += TicksPerDay
	}
	totalMinutes := (t * 24 * 60) / TicksPerDay
	totalMinutes += 6 * 60
	totalMinutes %= 24 * 60
	hour = int(totalMinutes / 60)
	minute = int(totalMinutes % 60)
	return hour, minute
}

// TickIsDaytime reports whether the given game tick falls during daytime
// (ticks 0-11999, i.e., 6:00 AM to 6:00 PM).
func TickIsDaytime(gameTick int64) bool {
	t := gameTick % TicksPerDay
	if t < 0 {
		t += TicksPerDay
	}
	return t < 12000
}
