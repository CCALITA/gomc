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
