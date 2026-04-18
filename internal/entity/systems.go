package entity

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// noBlocks is a helper that returns no block AABBs, used as a default
// collision provider when none is configured.
func noBlocks(_ mcmath.AABB) []mcmath.AABB { return nil }

// ---------------------------------------------------------------------------
// PhysicsSystem
// ---------------------------------------------------------------------------

// PhysicsSystem applies physics simulation to entities with Transform and
// PhysicsBody components.
type PhysicsSystem struct {
	// GetBlockAABBs returns solid block AABBs overlapping the given region.
	// When nil, a no-collision fallback is used.
	GetBlockAABBs func(mcmath.AABB) []mcmath.AABB
}

// Update integrates physics for every entity that has both Transform and
// PhysicsBody.
func (s *PhysicsSystem) Update(w *ecs.World, dt float64) {
	getter := s.GetBlockAABBs
	if getter == nil {
		getter = noBlocks
	}
	dt32 := float32(dt)

	ecs.Query2[Transform, PhysicsBody](w, func(e ecs.Entity, t *Transform, pb *PhysicsBody) {
		// Sync body position from transform before resolve.
		pb.Body.Position = t.Position

		physics.ResolveCollision(pb.Body, dt32, getter)

		// Write back resolved position.
		t.Position = pb.Body.Position
	})
}

// ---------------------------------------------------------------------------
// AISystem
// ---------------------------------------------------------------------------

// AISystem runs the mob behaviour state machine for entities with AI and
// Transform components.
type AISystem struct {
	// Rand is the random source; defaults to the global rand if nil.
	Rand *rand.Rand

	// ChaseRange overrides the default chase range when > 0.
	ChaseRange float32
	// AttackRange overrides the default attack range when > 0.
	AttackRange float32
}

func (s *AISystem) rng() *rand.Rand {
	return s.Rand
}

// chaseRange returns the effective chase range, falling back to the
// package-level default when no override is configured.
func (s *AISystem) chaseRange() float32 {
	if s.ChaseRange > 0 {
		return s.ChaseRange
	}
	return aiChaseRange
}

// attackRange returns the effective attack range, falling back to the
// package-level default when no override is configured.
func (s *AISystem) attackRange() float32 {
	if s.AttackRange > 0 {
		return s.AttackRange
	}
	return aiAttackRange
}

const (
	aiChaseRange  float32 = 16.0
	aiAttackRange float32 = 2.0

	// aiAttackCooldown is the minimum time between mob attacks (seconds).
	aiAttackCooldown float64 = 1.0

	// zombieDamage is the damage dealt per Zombie attack.
	zombieDamage float32 = 3.0
	// skeletonDamage is the damage dealt per Skeleton melee attack.
	skeletonDamage float32 = 2.0

	// mobKnockbackStrength is the horizontal knockback speed applied by mob attacks.
	mobKnockbackStrength float32 = 6.0
	// mobKnockbackUpward is the upward velocity component of mob knockback.
	mobKnockbackUpward float32 = 4.0
)

// playerEntry holds a player entity and its position for AI target detection.
type playerEntry struct {
	entity ecs.Entity
	pos    mcmath.Vec3
}

// nearestInfo holds the result of a nearest-player search.
type nearestInfo struct {
	dist   float32
	pos    mcmath.Vec3
	entity ecs.Entity
}

// Update processes AI state transitions for every AI entity.
func (s *AISystem) Update(w *ecs.World, dt float64) {
	// Collect player positions so mobs can detect them.
	var players []playerEntry

	ecs.Query2[Transform, EntityTypeComp](w, func(e ecs.Entity, t *Transform, et *EntityTypeComp) {
		if et.Type == TypePlayer {
			players = append(players, playerEntry{entity: e, pos: t.Position})
		}
	})

	ecs.Query2[AI, Transform](w, func(e ecs.Entity, ai *AI, t *Transform) {
		ai.Timer -= dt

		if ai.Passive {
			s.handlePassive(ai)
			return
		}

		nearest := findNearestPlayer(t.Position, players)

		switch ai.State {
		case AIIdle:
			s.handleIdleOrWander(ai, nearest, AIWander, s.randomWanderTime)
		case AIWander:
			s.handleIdleOrWander(ai, nearest, AIIdle, s.randomIdleTime)
		case AIChase:
			s.handleChase(ai, t, nearest, dt)
		case AIAttack:
			s.handleAttack(w, e, ai, t, nearest)
		case AIFlee:
			s.handleFlee(ai, nearest)
		}
	})
}

// handlePassive handles AI state transitions for passive mobs (idle/wander only).
func (s *AISystem) handlePassive(ai *AI) {
	switch ai.State {
	case AIIdle:
		if ai.Timer <= 0 {
			ai.State = AIWander
			ai.Timer = s.randomWanderTime()
		}
	case AIWander:
		if ai.Timer <= 0 {
			ai.State = AIIdle
			ai.Timer = s.randomIdleTime()
		}
	default:
		ai.State = AIIdle
		ai.Timer = s.randomIdleTime()
	}
}

// findNearestPlayer returns the nearest player to the given position.
func findNearestPlayer(pos mcmath.Vec3, players []playerEntry) nearestInfo {
	info := nearestInfo{dist: math.MaxFloat32}
	for _, p := range players {
		d := pos.Distance(p.pos)
		if d < info.dist {
			info.dist = d
			info.pos = p.pos
			info.entity = p.entity
		}
	}
	return info
}

// handleIdleOrWander handles the shared idle/wander logic: transition to chase
// if a player is in range, otherwise switch to fallbackState when the timer expires.
func (s *AISystem) handleIdleOrWander(ai *AI, nearest nearestInfo, fallbackState uint8, timerFn func() float64) {
	if nearest.dist <= s.chaseRange() {
		ai.State = AIChase
		ai.Target = nearest.entity
		return
	}
	if ai.Timer <= 0 {
		ai.State = fallbackState
		ai.Timer = timerFn()
	}
}

// handleChase moves the mob toward the nearest player, switching to attack or idle.
func (s *AISystem) handleChase(ai *AI, t *Transform, nearest nearestInfo, dt float64) {
	ai.Target = nearest.entity
	if nearest.dist > s.chaseRange() {
		ai.State = AIIdle
		ai.Timer = s.randomIdleTime()
		return
	}
	if nearest.dist <= s.attackRange() {
		ai.State = AIAttack
		return
	}
	dir := nearest.pos.Sub(t.Position).Normalize()
	speed := float32(4.0)
	t.Position = t.Position.Add(dir.Scale(speed * float32(dt)))
}

// handleAttack deals damage on cooldown or transitions back to chase.
func (s *AISystem) handleAttack(w *ecs.World, e ecs.Entity, ai *AI, t *Transform, nearest nearestInfo) {
	if nearest.dist > s.attackRange() {
		ai.State = AIChase
		ai.Target = nearest.entity
		return
	}
	if ai.Timer <= 0 {
		ai.Timer = aiAttackCooldown
		dmgAmount := mobDamageForEntity(w, e)
		if dmgAmount > 0 {
			applyMobDamage(w, nearest.entity, t.Position, dmgAmount)
		}
	}
}

// handleFlee transitions back to idle when the player is far enough away.
func (s *AISystem) handleFlee(ai *AI, nearest nearestInfo) {
	if nearest.dist > s.chaseRange() {
		ai.State = AIIdle
		ai.Timer = s.randomIdleTime()
	}
}

func (s *AISystem) randomWanderTime() float64 {
	if s.Rand != nil {
		return 3.0 + s.Rand.Float64()*4.0
	}
	return 3.0 + rand.Float64()*4.0
}

func (s *AISystem) randomIdleTime() float64 {
	if s.Rand != nil {
		return 1.0 + s.Rand.Float64()*3.0
	}
	return 1.0 + rand.Float64()*3.0
}

// mobDamageForEntity returns the attack damage for the given mob entity
// based on its EntityTypeComp. Returns 0 for unknown types.
func mobDamageForEntity(w *ecs.World, e ecs.Entity) float32 {
	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	if !ok {
		return 0
	}
	switch et.Type {
	case TypeZombie:
		return zombieDamage
	case TypeSkeleton:
		return skeletonDamage
	default:
		return 0
	}
}

// applyMobDamage applies a Damage component to the target entity with
// knockback directed away from the attacker position.
func applyMobDamage(w *ecs.World, target ecs.Entity, attackerPos mcmath.Vec3, amount float32) {
	targetTransform, ok := ecs.GetStore[Transform](w).Get(target)
	if !ok {
		return
	}

	kb := KnockbackFromTo(attackerPos, targetTransform.Position, mobKnockbackStrength, mobKnockbackUpward)

	ecs.GetStore[Damage](w).Set(target, Damage{
		Amount:    amount,
		Knockback: kb,
	})
}

// ---------------------------------------------------------------------------
// LifetimeSystem
// ---------------------------------------------------------------------------

// LifetimeSystem decrements the Lifetime component and destroys the entity
// when it expires.
type LifetimeSystem struct{}

// Update ticks every Lifetime component, destroying expired entities.
func (s *LifetimeSystem) Update(w *ecs.World, dt float64) {
	store := ecs.GetStore[Lifetime](w)
	var toDestroy []ecs.Entity

	store.Each(func(e ecs.Entity, lt *Lifetime) {
		lt.Remaining -= dt
		if lt.Remaining <= 0 {
			toDestroy = append(toDestroy, e)
		}
	})

	for _, e := range toDestroy {
		w.DestroyEntity(e)
	}
}

// ---------------------------------------------------------------------------
// DamageSystem
// ---------------------------------------------------------------------------

// DamageSystem applies pending Damage to Health and PhysicsBody, then removes
// the Damage component.
type DamageSystem struct{}

// Update processes all Damage components.
func (s *DamageSystem) Update(w *ecs.World, dt float64) {
	dmgStore := ecs.GetStore[Damage](w)
	healthStore := ecs.GetStore[Health](w)
	pbStore := ecs.GetStore[PhysicsBody](w)
	armorStore := ecs.GetStore[Armor](w)

	var processed []ecs.Entity

	dmgStore.Each(func(e ecs.Entity, d *Damage) {
		amount := d.Amount

		if a, ok := armorStore.Get(e); ok {
			amount = a.DamageReduction(amount)
			a.DamageArmor(1)
		}

		if h, ok := healthStore.Get(e); ok {
			h.Current -= amount
		}

		// Apply knockback to physics body.
		if pb, ok := pbStore.Get(e); ok {
			pb.Body.Velocity = pb.Body.Velocity.Add(d.Knockback)
		}

		processed = append(processed, e)
	})

	for _, e := range processed {
		dmgStore.Remove(e)
	}
}

// ---------------------------------------------------------------------------
// HealthSystem
// ---------------------------------------------------------------------------

// HealthSystem destroys entities whose health has dropped to zero or below.
// Player entities are skipped because their death is handled by the game layer.
type HealthSystem struct{}

// Update checks all Health components and destroys dead non-player entities.
func (s *HealthSystem) Update(w *ecs.World, dt float64) {
	store := ecs.GetStore[Health](w)
	etStore := ecs.GetStore[EntityTypeComp](w)
	var toDestroy []ecs.Entity

	store.Each(func(e ecs.Entity, h *Health) {
		if h.Current <= 0 {
			// Skip player entities; game layer handles death/respawn.
			if et, ok := etStore.Get(e); ok && et.Type == TypePlayer {
				return
			}
			toDestroy = append(toDestroy, e)
		}
	})

	for _, e := range toDestroy {
		w.DestroyEntity(e)
	}
}
