package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// XPForNextLevel formula
// ---------------------------------------------------------------------------

func TestXPForNextLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"level 0", 0, 7},
		{"level 1", 1, 9},
		{"level 10", 10, 27},
		{"level 16", 16, 39},
		{"level 17", 17, 47},
		{"level 25", 25, 87},
		{"level 31", 31, 117},
		{"level 32", 32, 130},
		{"level 40", 40, 202},
		{"negative level", -1, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := XPForNextLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// AddXP — level-up
// ---------------------------------------------------------------------------

func TestAddXP_LevelUp(t *testing.T) {
	w := ecs.NewWorld()
	player := w.NewEntity()

	// Level 0 needs 7 XP to level up.
	ecs.GetStore[Experience](w).Set(player, Experience{Level: 0, XP: 0, TotalXP: 0})

	AddXP(w, player, 7)

	exp, ok := ecs.GetStore[Experience](w).Get(player)
	assert.True(t, ok)
	assert.Equal(t, 1, exp.Level)
	assert.Equal(t, 0, exp.XP)
	assert.Equal(t, 7, exp.TotalXP)
}

// ---------------------------------------------------------------------------
// AddXP — no level-up
// ---------------------------------------------------------------------------

func TestAddXP_NoLevelUp(t *testing.T) {
	w := ecs.NewWorld()
	player := w.NewEntity()

	ecs.GetStore[Experience](w).Set(player, Experience{Level: 0, XP: 0, TotalXP: 0})

	AddXP(w, player, 3)

	exp, ok := ecs.GetStore[Experience](w).Get(player)
	assert.True(t, ok)
	assert.Equal(t, 0, exp.Level)
	assert.Equal(t, 3, exp.XP)
	assert.Equal(t, 3, exp.TotalXP)
}

// ---------------------------------------------------------------------------
// AddXP — multiple level-ups
// ---------------------------------------------------------------------------

func TestAddXP_MultipleLevelUps(t *testing.T) {
	w := ecs.NewWorld()
	player := w.NewEntity()

	ecs.GetStore[Experience](w).Set(player, Experience{Level: 0, XP: 0, TotalXP: 0})

	// Level 0 needs 7, level 1 needs 9 -> total 16 for two levels.
	AddXP(w, player, 20)

	exp, ok := ecs.GetStore[Experience](w).Get(player)
	assert.True(t, ok)
	assert.Equal(t, 2, exp.Level)
	assert.Equal(t, 4, exp.XP) // 20 - 7 - 9 = 4
	assert.Equal(t, 20, exp.TotalXP)
}

// ---------------------------------------------------------------------------
// AddXP — entity without Experience component (no-op)
// ---------------------------------------------------------------------------

func TestAddXP_NoExperience(t *testing.T) {
	w := ecs.NewWorld()
	player := w.NewEntity()

	// No Experience component set; AddXP should be a no-op.
	AddXP(w, player, 10)

	_, ok := ecs.GetStore[Experience](w).Get(player)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// XPSystem — awards XP on mob death
// ---------------------------------------------------------------------------

func TestXPSystem_AwardsOnMobDeath(t *testing.T) {
	w := ecs.NewWorld()

	// Player with experience.
	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{X: 0, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})
	ecs.GetStore[Experience](w).Set(player, Experience{})

	// Dead zombie near the player.
	zombie := w.NewEntity()
	ecs.GetStore[Transform](w).Set(zombie, Transform{Position: mcmath.Vec3{X: 5, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(zombie, EntityTypeComp{Type: TypeZombie})
	ecs.GetStore[Health](w).Set(zombie, Health{Current: 0, Max: 20})

	sys := &XPSystem{}
	sys.Update(w, 0.05)

	exp, ok := ecs.GetStore[Experience](w).Get(player)
	assert.True(t, ok)
	assert.Equal(t, 5, exp.TotalXP)
	assert.Equal(t, 5, exp.XP)
}

// ---------------------------------------------------------------------------
// XPSystem — alive mobs are ignored
// ---------------------------------------------------------------------------

func TestXPSystem_AliveMobIgnored(t *testing.T) {
	w := ecs.NewWorld()

	player := w.NewEntity()
	ecs.GetStore[Transform](w).Set(player, Transform{Position: mcmath.Vec3{}})
	ecs.GetStore[EntityTypeComp](w).Set(player, EntityTypeComp{Type: TypePlayer})
	ecs.GetStore[Experience](w).Set(player, Experience{})

	zombie := w.NewEntity()
	ecs.GetStore[Transform](w).Set(zombie, Transform{Position: mcmath.Vec3{X: 5, Y: 0, Z: 0}})
	ecs.GetStore[EntityTypeComp](w).Set(zombie, EntityTypeComp{Type: TypeZombie})
	ecs.GetStore[Health](w).Set(zombie, Health{Current: 10, Max: 20})

	sys := &XPSystem{}
	sys.Update(w, 0.05)

	exp, _ := ecs.GetStore[Experience](w).Get(player)
	assert.Equal(t, 0, exp.TotalXP)
}

// ---------------------------------------------------------------------------
// xpForMobType
// ---------------------------------------------------------------------------

func TestXPForMobType(t *testing.T) {
	assert.Equal(t, 5, xpForMobType(TypeZombie))
	assert.Equal(t, 5, xpForMobType(TypeSkeleton))
	assert.Equal(t, 5, xpForMobType(TypeCreeper))
	assert.Equal(t, 5, xpForMobType(TypeSpider))
	assert.Equal(t, 0, xpForMobType(TypePlayer))
	assert.Equal(t, 0, xpForMobType(TypeCow))
}
