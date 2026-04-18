package world

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ExplosionDamage records the damage and knockback applied to a single entity
// by an explosion.
type ExplosionDamage struct {
	Entity    ecs.Entity
	Damage    float32
	Knockback mcmath.Vec3
}

// ExplosionResult captures every block destroyed and every entity damaged by a
// single call to Explode.
type ExplosionResult struct {
	DestroyedBlocks []mcmath.BlockPos
	DamagedEntities []ExplosionDamage
}

// blastResistant reports whether a block type is immune to explosions.
func blastResistant(id uint16) bool {
	base := block.BaseID(id)
	return base == block.Bedrock || base == block.Obsidian
}

// Explode processes an explosion centered at center with the given power.
// Blocks within the blast radius are probabilistically destroyed (except for
// bedrock and obsidian). Entities within the radius take distance-attenuated
// damage and knockback. A deterministic *rand.Rand may be supplied for testing;
// if rng is nil a default source is used.
func Explode(center mcmath.Vec3, power float32, w *World, ecsWorld *ecs.World, rng *rand.Rand) ExplosionResult {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	radius := power * 1.5
	result := ExplosionResult{}

	if radius <= 0 {
		return result
	}

	// --- destroy blocks ---
	iRadius := int32(math.Ceil(float64(radius)))
	centerBlock := center.Floor()
	blockCenterOffset := mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5}

	for dx := -iRadius; dx <= iRadius; dx++ {
		for dy := -iRadius; dy <= iRadius; dy++ {
			for dz := -iRadius; dz <= iRadius; dz++ {
				bp := mcmath.BlockPos{
					X: centerBlock.X + dx,
					Y: centerBlock.Y + dy,
					Z: centerBlock.Z + dz,
				}

				dist := center.Distance(bp.ToVec3().Add(blockCenterOffset))
				if dist >= radius {
					continue
				}

				id := w.GetBlock(bp)
				if id == block.Air {
					continue
				}
				if blastResistant(id) {
					continue
				}

				attenuation := 1 - dist/radius
				if rng.Float32() < attenuation {
					w.SetBlock(bp, block.Air)
					result.DestroyedBlocks = append(result.DestroyedBlocks, bp)
				}
			}
		}
	}

	// --- damage entities ---
	if ecsWorld == nil {
		return result
	}

	transforms := ecs.GetStore[entity.Transform](ecsWorld)
	damageStore := ecs.GetStore[entity.Damage](ecsWorld)
	transforms.Each(func(e ecs.Entity, tf *entity.Transform) {
		dist := center.Distance(tf.Position)
		if dist >= radius {
			return
		}

		factor := 1 - dist/radius
		dmg := power * 2 * factor

		dir := tf.Position.Sub(center)
		if dir.LengthSq() == 0 {
			dir = mcmath.Vec3{Y: 1}
		}
		dir = dir.Normalize()
		kb := dir.Scale(power * factor)

		damageStore.Set(e, entity.Damage{
			Amount:    dmg,
			Knockback: kb,
		})

		result.DamagedEntities = append(result.DamagedEntities, ExplosionDamage{
			Entity:    e,
			Damage:    dmg,
			Knockback: kb,
		})
	})

	return result
}

// TNTEntity is an ECS component for a primed TNT block. FuseTime counts down
// each tick; when it reaches zero the TNT detonates.
type TNTEntity struct {
	FuseTime float32
	Power    float32
}

// TNTSystem decrements the fuse of every TNTEntity each tick and triggers an
// explosion when the fuse expires.
type TNTSystem struct {
	BlockWorld *World
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
		Explode(d.pos, d.power, s.BlockWorld, w, nil)
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
