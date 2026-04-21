package entity

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// EntityType constants identify the kind of entity.
const (
	// TypePlayer identifies a player-controlled entity.
	TypePlayer uint8 = 1
	// TypeZombie identifies a zombie hostile mob.
	TypeZombie uint8 = 2
	// TypeSkeleton identifies a skeleton hostile mob.
	TypeSkeleton uint8 = 3
	// TypeCreeper identifies a creeper hostile mob.
	TypeCreeper uint8 = 4
	// TypeItem identifies a dropped item entity.
	TypeItem uint8 = 5
	// TypeArrow identifies an arrow projectile.
	TypeArrow uint8 = 6
	// TypeCow identifies a cow passive mob.
	TypeCow uint8 = 7
	// TypePig identifies a pig passive mob.
	TypePig uint8 = 8
	// TypeSheep identifies a sheep passive mob.
	TypeSheep uint8 = 9
	// TypeChicken identifies a chicken passive mob.
	TypeChicken uint8 = 10
	// TypeSpider identifies a spider hostile mob.
	TypeSpider uint8 = 11
	// TypeEnderman identifies an enderman hostile mob.
	TypeEnderman uint8 = 12
	// TypeMinecart identifies a minecart entity.
	TypeMinecart uint8 = 15
)

// AI state constants define the mob behaviour state machine states.
const (
	// AIIdle means the mob is standing still.
	AIIdle uint8 = 0
	// AIWander means the mob is moving randomly.
	AIWander uint8 = 1
	// AIChase means the mob is pursuing a target.
	AIChase uint8 = 2
	// AIFlee means the mob is running away from a threat.
	AIFlee uint8 = 3
	// AIAttack means the mob is attacking its target.
	AIAttack uint8 = 4
)

// Transform holds position and orientation for an entity.
type Transform struct {
	Position mcmath.Vec3
	Yaw      float32
	Pitch    float32
}

// PhysicsBody wraps a physics.Body for ECS storage.
type PhysicsBody struct {
	Body *physics.Body
}

// Health tracks current and maximum hit points.
type Health struct {
	Current float32
	Max     float32
}

// Name holds a display name for an entity.
type Name struct {
	Value string
}

// EntityTypeComp identifies the kind of entity.
type EntityTypeComp struct {
	Type uint8
}

// AI holds mob behaviour state machine data.
type AI struct {
	State           uint8
	Target          ecs.Entity
	Timer           float64
	Passive         bool // Passive mobs only Idle and Wander, never Chase or Attack.
	SpiderClimb     bool // Spider can climb walls when chasing.
	EndermanTeleport bool // Enderman teleports when damaged.
}

// ItemSlot represents an item stack in an inventory slot.
type ItemSlot struct {
	ItemID uint16
	Count  int
}

// Inventory holds the player inventory and active hotbar index.
type Inventory struct {
	Slots       [36]ItemSlot
	HotbarIndex int
}

// Lifetime tracks how long a temporary entity has left before despawning.
type Lifetime struct {
	Remaining float64
}

// Damage is a one-shot component applied to deal damage and knockback.
type Damage struct {
	Amount    float32
	Knockback mcmath.Vec3
}

// MinecartData holds minecart-specific state for rail riding.
type MinecartData struct {
	OnRail    bool
	Speed     float32
	Direction mcmath.Direction
	Rider     ecs.Entity
}

// KnockbackFromTo computes a knockback vector directed horizontally from
// the attacker position toward the target position, with the given
// horizontal strength and upward component.
func KnockbackFromTo(attackerPos, targetPos mcmath.Vec3, horizontal, upward float32) mcmath.Vec3 {
	dir := targetPos.Sub(attackerPos)
	dir.Y = 0
	dir = dir.Normalize()
	return mcmath.Vec3{
		X: dir.X * horizontal,
		Y: upward,
		Z: dir.Z * horizontal,
	}
}
