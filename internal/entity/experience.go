package entity

import (
	"math"

	"github.com/fanxiyao/gomc/internal/ecs"
)

// XP reward constants for each mob type.
const (
	xpRewardZombie   = 5
	xpRewardSkeleton = 5
	xpRewardCreeper  = 5
	xpRewardSpider   = 5
)

// Experience tracks the player's experience level, current XP within that
// level, and total XP collected over the lifetime.
type Experience struct {
	Level   int
	XP      int
	TotalXP int
}

// XPForNextLevel returns the XP required to advance from the given level
// to the next. The formula follows Minecraft conventions:
//
//	levels  0-16: 2*level + 7
//	levels 17-31: 5*level - 38
//	levels 32+  : 9*level - 158
func XPForNextLevel(level int) int {
	switch {
	case level < 0:
		return XPForNextLevel(0)
	case level <= 16:
		return 2*level + 7
	case level <= 31:
		return 5*level - 38
	default:
		return 9*level - 158
	}
}

// AddXP awards the given amount of XP to the player entity, leveling up
// as many times as needed. The Experience component is updated in-place.
func AddXP(w *ecs.World, player ecs.Entity, amount int) {
	store := ecs.GetStore[Experience](w)
	exp, ok := store.Get(player)
	if !ok {
		return
	}

	exp.TotalXP += amount
	exp.XP += amount

	for exp.XP >= XPForNextLevel(exp.Level) {
		exp.XP -= XPForNextLevel(exp.Level)
		exp.Level++
	}
}

// xpForMobType returns the XP reward for killing the given mob type,
// or 0 if the mob type does not grant XP.
func xpForMobType(mobType uint8) int {
	switch mobType {
	case TypeZombie:
		return xpRewardZombie
	case TypeSkeleton:
		return xpRewardSkeleton
	case TypeCreeper:
		return xpRewardCreeper
	case TypeSpider:
		return xpRewardSpider
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// XPSystem
// ---------------------------------------------------------------------------

// XPSystem detects dead mobs (Health.Current <= 0) each tick and awards XP
// to the nearest player. It runs before HealthSystem destroys the entities.
type XPSystem struct{}

// Update scans for mobs with zero or negative health and awards XP to the
// nearest player entity.
func (s *XPSystem) Update(w *ecs.World, _ float64) {
	// Collect player positions.
	var players []playerEntry
	ecs.Query2[Transform, EntityTypeComp](w, func(e ecs.Entity, t *Transform, et *EntityTypeComp) {
		if et.Type == TypePlayer {
			players = append(players, playerEntry{entity: e, pos: t.Position})
		}
	})

	if len(players) == 0 {
		return
	}

	// Find dead mobs and award XP.
	ecs.Query3[Health, EntityTypeComp, Transform](w, func(e ecs.Entity, h *Health, et *EntityTypeComp, t *Transform) {
		if h.Current > 0 {
			return
		}

		reward := xpForMobType(et.Type)
		if reward <= 0 {
			return
		}

		nearest := findNearestPlayer(t.Position, players)
		if nearest.dist == math.MaxFloat32 {
			return
		}

		AddXP(w, nearest.entity, reward)
	})
}
