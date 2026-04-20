package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// playerBBox is the bounding box for a player entity (0.6 wide, 1.8 tall).
var playerBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
	Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
}

// mobBBox is the bounding box for standard hostile mobs (0.6 wide, 1.95 tall).
var mobBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
	Max: mcmath.Vec3{X: 0.3, Y: 1.95, Z: 0.3},
}

// itemBBox is the bounding box for a dropped item entity (0.25 cube).
var itemBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.125, Y: 0, Z: -0.125},
	Max: mcmath.Vec3{X: 0.125, Y: 0.25, Z: 0.125},
}

// cowBBox is the bounding box for a cow (0.9 wide, 1.4 tall).
var cowBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.45, Y: 0, Z: -0.45},
	Max: mcmath.Vec3{X: 0.45, Y: 1.4, Z: 0.45},
}

// pigBBox is the bounding box for a pig (0.9 wide, 0.9 tall).
var pigBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.45, Y: 0, Z: -0.45},
	Max: mcmath.Vec3{X: 0.45, Y: 0.9, Z: 0.45},
}

// sheepBBox is the bounding box for a sheep (0.9 wide, 1.3 tall).
var sheepBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.45, Y: 0, Z: -0.45},
	Max: mcmath.Vec3{X: 0.45, Y: 1.3, Z: 0.45},
}

// chickenBBox is the bounding box for a chicken (0.4 wide, 0.7 tall).
var chickenBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.2, Y: 0, Z: -0.2},
	Max: mcmath.Vec3{X: 0.2, Y: 0.7, Z: 0.2},
}

// SpawnPlayer creates a fully configured player entity.
func SpawnPlayer(w *ecs.World, name string, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(playerBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[Name](w).Set(e, Name{Value: name})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypePlayer})
	ecs.GetStore[Inventory](w).Set(e, Inventory{})
	ecs.GetStore[Armor](w).Set(e, Armor{})
	ecs.GetStore[Hunger](w).Set(e, NewHunger())
	ecs.GetStore[Experience](w).Set(e, Experience{})

	return e
}

// SpawnZombie creates a zombie mob entity.
func SpawnZombie(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(mobBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeZombie})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle})

	return e
}

// SpawnSkeleton creates a skeleton mob entity.
func SpawnSkeleton(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(mobBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeSkeleton})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle})

	return e
}

// SpawnCow creates a cow passive mob entity.
func SpawnCow(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(cowBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 10, Max: 10})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeCow})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle, Passive: true})
	ecs.GetStore[Breedable](w).Set(e, Breedable{Scale: 1.0})

	return e
}

// SpawnPig creates a pig passive mob entity.
func SpawnPig(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(pigBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 10, Max: 10})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypePig})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle, Passive: true})
	ecs.GetStore[Breedable](w).Set(e, Breedable{Scale: 1.0})

	return e
}

// SpawnSheep creates a sheep passive mob entity.
func SpawnSheep(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(sheepBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 8, Max: 8})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeSheep})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle, Passive: true})
	ecs.GetStore[Breedable](w).Set(e, Breedable{Scale: 1.0})

	return e
}

// SpawnChicken creates a chicken passive mob entity.
func SpawnChicken(w *ecs.World, pos mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(chickenBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[Health](w).Set(e, Health{Current: 4, Max: 4})
	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeChicken})
	ecs.GetStore[AI](w).Set(e, AI{State: AIIdle, Passive: true})
	ecs.GetStore[Breedable](w).Set(e, Breedable{Scale: 1.0})

	return e
}

// xpOrbBBox is the bounding box for an XP orb entity (0.2 cube).
var xpOrbBBox = mcmath.AABB{
	Min: mcmath.Vec3{X: -0.1, Y: 0, Z: -0.1},
	Max: mcmath.Vec3{X: 0.1, Y: 0.2, Z: 0.1},
}

// SpawnXPOrb creates an experience orb entity at the given position that
// awards the specified amount of XP. The orb has a 60-second lifetime.
func SpawnXPOrb(w *ecs.World, pos mcmath.Vec3, amount int) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(xpOrbBBox)
	body.Position = pos
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeXPOrb})
	ecs.GetStore[Experience](w).Set(e, Experience{XP: amount})
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: 60})

	return e
}

// SpawnItem creates a dropped item entity with a velocity and a limited
// lifetime of 300 seconds (5 minutes).
func SpawnItem(w *ecs.World, itemID uint16, pos mcmath.Vec3, velocity mcmath.Vec3) ecs.Entity {
	e := w.NewEntity()

	ecs.GetStore[Transform](w).Set(e, Transform{Position: pos})

	body := physics.NewBody(itemBBox)
	body.Position = pos
	body.Velocity = velocity
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	ecs.GetStore[EntityTypeComp](w).Set(e, EntityTypeComp{Type: TypeItem})
	ecs.GetStore[Inventory](w).Set(e, Inventory{
		Slots: func() [36]ItemSlot {
			var s [36]ItemSlot
			s[0] = ItemSlot{ItemID: itemID, Count: 1}
			return s
		}(),
	})
	ecs.GetStore[Lifetime](w).Set(e, Lifetime{Remaining: 300})

	return e
}
