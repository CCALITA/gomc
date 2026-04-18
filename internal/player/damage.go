package player

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
	"github.com/fanxiyao/gomc/internal/world"
)

// Damage type constants identify the source of environmental damage.
const (
	DamageTypeFall  = "fall"
	DamageTypeLava  = "lava"
	DamageTypeDrown = "drown"
	DamageTypeFire  = "fire"
)

const (
	// fallDamageThreshold is the minimum fall distance (in blocks) before
	// fall damage is applied. Falls of 3 blocks or fewer deal no damage.
	fallDamageThreshold float32 = 3.0

	// lavaDamagePerSecond is the amount of damage dealt per second while
	// the player's feet are in lava.
	lavaDamagePerSecond float32 = 4.0

	// drownBreathHoldTime is the number of seconds a player can remain
	// submerged before drowning damage begins.
	drownBreathHoldTime float32 = 15.0

	// drownDamagePerSecond is the damage dealt per second once the breath
	// hold timer expires.
	drownDamagePerSecond float32 = 2.0

	// fireDamagePerSecond is the damage dealt per second while standing
	// on a fire block.
	fireDamagePerSecond float32 = 1.0
)

// DamageEvent describes a single damage instance produced by the tracker.
type DamageEvent struct {
	Amount float32
	Type   string
}

// PlayerDamageTracker tracks environmental hazard state and produces
// DamageEvents each tick.
type PlayerDamageTracker struct {
	FallDistance     float32
	LavaContactTime float32
	UnderwaterTime  float32
	FireTime        float32
	wasOnGround     bool
}

// NewPlayerDamageTracker returns a tracker initialised with the player
// already considered on the ground (no pending fall).
func NewPlayerDamageTracker() *PlayerDamageTracker {
	return &PlayerDamageTracker{wasOnGround: true}
}

// Update advances the damage tracker by dt seconds and returns any damage
// events that occurred. playerPos is the entity's foot-level position.
func (t *PlayerDamageTracker) Update(
	body *physics.Body,
	w *world.World,
	playerPos mcmath.Vec3,
	dt float32,
) []DamageEvent {
	var events []DamageEvent

	feetBlock := playerPos.Floor()

	events = t.updateFall(body, dt, events)
	events = t.updateLava(w, feetBlock, dt, events)
	events = t.updateDrowning(w, playerPos, dt, events)
	events = t.updateFire(w, feetBlock, dt, events)

	return events
}

// updateFall tracks fall distance and emits a fall damage event on landing.
func (t *PlayerDamageTracker) updateFall(
	body *physics.Body,
	dt float32,
	events []DamageEvent,
) []DamageEvent {
	if !body.OnGround && body.Velocity.Y < 0 {
		t.FallDistance += -body.Velocity.Y * dt
	}

	// Transition from airborne to grounded.
	if body.OnGround && !t.wasOnGround {
		if t.FallDistance > fallDamageThreshold {
			damage := t.FallDistance - fallDamageThreshold
			events = append(events, DamageEvent{Amount: damage, Type: DamageTypeFall})
		}
		t.FallDistance = 0
	}

	t.wasOnGround = body.OnGround
	return events
}

// updateLava checks whether the player's feet are in lava and applies
// continuous damage.
func (t *PlayerDamageTracker) updateLava(
	w *world.World,
	feetBlock mcmath.BlockPos,
	dt float32,
	events []DamageEvent,
) []DamageEvent {
	blockID := w.GetBlock(feetBlock)

	if block.IsLava(blockID) {
		t.LavaContactTime += dt
		damage := lavaDamagePerSecond * dt
		events = append(events, DamageEvent{Amount: damage, Type: DamageTypeLava})
	} else {
		t.LavaContactTime = 0
	}

	return events
}

// updateDrowning checks whether the player's head is submerged in water
// and starts dealing damage after the breath-hold period expires.
func (t *PlayerDamageTracker) updateDrowning(
	w *world.World,
	playerPos mcmath.Vec3,
	dt float32,
	events []DamageEvent,
) []DamageEvent {
	headPos := playerPos.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0})
	headBlock := headPos.Floor()
	blockID := w.GetBlock(headBlock)

	if block.IsWater(blockID) {
		t.UnderwaterTime += dt
		if t.UnderwaterTime >= drownBreathHoldTime {
			damage := drownDamagePerSecond * dt
			events = append(events, DamageEvent{Amount: damage, Type: DamageTypeDrown})
		}
	} else {
		t.UnderwaterTime = 0
	}

	return events
}

// updateFire checks whether the player is standing on a fire block and
// applies continuous damage.
func (t *PlayerDamageTracker) updateFire(
	w *world.World,
	feetBlock mcmath.BlockPos,
	dt float32,
	events []DamageEvent,
) []DamageEvent {
	blockID := w.GetBlock(feetBlock)

	if block.IsFire(blockID) {
		t.FireTime += dt
		damage := fireDamagePerSecond * dt
		events = append(events, DamageEvent{Amount: damage, Type: DamageTypeFire})
	} else {
		t.FireTime = 0
	}

	return events
}
