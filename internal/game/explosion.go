package game

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/world"
)

// applyExplosionDamage queries the ECS for entities near the explosion center
// and applies distance-attenuated damage and knockback via the Damage
// component. This keeps all ECS logic out of the world data layer.
func applyExplosionDamage(result world.BlockExplosionResult, ecsWorld *ecs.World) {
	if ecsWorld == nil || result.BlastRadius <= 0 {
		return
	}

	transforms := ecs.GetStore[entity.Transform](ecsWorld)
	damageStore := ecs.GetStore[entity.Damage](ecsWorld)

	transforms.Each(func(e ecs.Entity, tf *entity.Transform) {
		dist := result.Center.Distance(tf.Position)
		if dist >= result.BlastRadius {
			return
		}

		factor := 1 - dist/result.BlastRadius
		dmg := result.Power * 2 * factor

		dir := tf.Position.Sub(result.Center)
		if dir.LengthSq() == 0 {
			dir = mcmath.Vec3{Y: 1}
		}
		dir = dir.Normalize()
		kb := dir.Scale(result.Power * factor)

		damageStore.Set(e, entity.Damage{
			Amount:    dmg,
			Knockback: kb,
		})
	})
}

// TNTEntity is an ECS component for a primed TNT block. FuseTime counts down
// each tick; when it reaches zero the TNT detonates.
type TNTEntity struct {
	FuseTime float32
	Power    float32
}

// TNTSystem decrements the fuse of every TNTEntity each tick and triggers an
// explosion when the fuse expires. Block destruction is delegated to
// world.ExplodeBlocks; entity damage is applied via applyExplosionDamage.
type TNTSystem struct {
	BlockWorld *world.World
}

// Update implements ecs.System.
func (s *TNTSystem) Update(w *ecs.World, dt float64) {
	tntStore := ecs.GetStore[TNTEntity](w)
	transformStore := ecs.GetStore[entity.Transform](w)

	// Collect entities whose fuse has expired so we can detonate them after
	// iteration (modifying the store during Each is unsafe).
	type detonation struct {
		entity ecs.Entity
		pos    mcmath.Vec3
		power  float32
	}
	var pending []detonation

	ecs.Query2(w, func(e ecs.Entity, tnt *TNTEntity, tf *entity.Transform) {
		tnt.FuseTime -= float32(dt)
		if tnt.FuseTime <= 0 {
			pending = append(pending, detonation{
				entity: e,
				pos:    tf.Position,
				power:  tnt.Power,
			})
		}
	})

	for _, d := range pending {
		result := world.ExplodeBlocks(d.pos, d.power, s.BlockWorld, nil)
		applyExplosionDamage(result, w)
		tntStore.Remove(d.entity)
		transformStore.Remove(d.entity)
		w.DestroyEntity(d.entity)
	}
}

// SpawnTNT creates a primed TNT entity at pos with the given power. The default
// fuse is 4 seconds. A safety Lifetime of 5 seconds ensures cleanup if the
// explosion somehow fails to destroy the entity.
func SpawnTNT(w *ecs.World, pos mcmath.Vec3, power float32) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[entity.Transform](w).Set(e, entity.Transform{Position: pos})
	ecs.GetStore[TNTEntity](w).Set(e, TNTEntity{FuseTime: 4.0, Power: power})
	ecs.GetStore[entity.Lifetime](w).Set(e, entity.Lifetime{Remaining: 5.0})

	return e
}
