package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// ---------------------------------------------------------------------------
// Equip / Unequip round-trip
// ---------------------------------------------------------------------------

func TestArmor_EquipUnequipRoundTrip(t *testing.T) {
	a := Armor{}
	piece := ArmorPiece{ItemID: 404, DefensePoints: 2, Durability: 165, MaxDurability: 165}

	// Equip into empty slot returns zero piece.
	prev := a.Equip(SlotHelmet, piece)
	assert.True(t, prev.IsEmpty())

	// GetSlot returns the equipped piece.
	got := a.GetSlot(SlotHelmet)
	assert.Equal(t, uint16(404), got.ItemID)
	assert.Equal(t, 2, got.DefensePoints)
	assert.Equal(t, 165, got.Durability)

	// Unequip returns the piece and clears the slot.
	removed := a.Unequip(SlotHelmet)
	assert.Equal(t, uint16(404), removed.ItemID)
	assert.True(t, a.GetSlot(SlotHelmet).IsEmpty())
}

func TestArmor_EquipReplacesExisting(t *testing.T) {
	a := Armor{}
	first := ArmorPiece{ItemID: 400, DefensePoints: 1, Durability: 55, MaxDurability: 55}
	second := ArmorPiece{ItemID: 404, DefensePoints: 2, Durability: 165, MaxDurability: 165}

	a.Equip(SlotHelmet, first)
	prev := a.Equip(SlotHelmet, second)
	assert.Equal(t, uint16(400), prev.ItemID)
	assert.Equal(t, uint16(404), a.GetSlot(SlotHelmet).ItemID)
}

// ---------------------------------------------------------------------------
// TotalDefense
// ---------------------------------------------------------------------------

func TestArmor_TotalDefense(t *testing.T) {
	a := Armor{}
	// Full diamond: helmet=3, chest=8, legs=6, boots=3 => 20
	a.Equip(SlotHelmet, ArmorPiece{ItemID: 412, DefensePoints: 3, Durability: 363, MaxDurability: 363})
	a.Equip(SlotChestplate, ArmorPiece{ItemID: 413, DefensePoints: 8, Durability: 528, MaxDurability: 528})
	a.Equip(SlotLeggings, ArmorPiece{ItemID: 414, DefensePoints: 6, Durability: 495, MaxDurability: 495})
	a.Equip(SlotBoots, ArmorPiece{ItemID: 415, DefensePoints: 3, Durability: 429, MaxDurability: 429})

	assert.Equal(t, 20, a.TotalDefense())
}

func TestArmor_TotalDefense_Partial(t *testing.T) {
	a := Armor{}
	a.Equip(SlotChestplate, ArmorPiece{ItemID: 405, DefensePoints: 6, Durability: 240, MaxDurability: 240})
	assert.Equal(t, 6, a.TotalDefense())
}

func TestArmor_TotalDefense_Empty(t *testing.T) {
	a := Armor{}
	assert.Equal(t, 0, a.TotalDefense())
}

// ---------------------------------------------------------------------------
// DamageReduction formula
// ---------------------------------------------------------------------------

func TestArmor_DamageReduction(t *testing.T) {
	tests := []struct {
		name     string
		defense  int
		raw      float32
		expected float32
	}{
		{
			name:     "no armor",
			defense:  0,
			raw:      10,
			expected: 10,
		},
		{
			name:     "full diamond 20 def",
			defense:  20,
			raw:      10,
			expected: 10 * (1.0 - 20.0/25.0), // 10 * 0.2 = 2
		},
		{
			name:     "partial iron 15 def",
			defense:  15,
			raw:      20,
			expected: 20 * (1.0 - 15.0/25.0), // 20 * 0.4 = 8
		},
		{
			name:     "defense capped at 20",
			defense:  25,
			raw:      10,
			expected: 10 * (1.0 - 20.0/25.0), // capped: 10 * 0.2 = 2
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Armor{}
			// Put all defense on one slot for simplicity.
			a.Equip(SlotChestplate, ArmorPiece{
				ItemID:        999,
				DefensePoints: tc.defense,
				Durability:    100,
				MaxDurability: 100,
			})
			result := a.DamageReduction(tc.raw)
			assert.InDelta(t, float64(tc.expected), float64(result), 0.001)
		})
	}
}

// ---------------------------------------------------------------------------
// Armor durability wear
// ---------------------------------------------------------------------------

func TestArmor_DamageArmor(t *testing.T) {
	a := Armor{}
	a.Equip(SlotHelmet, ArmorPiece{ItemID: 400, DefensePoints: 1, Durability: 3, MaxDurability: 55})
	a.Equip(SlotBoots, ArmorPiece{ItemID: 403, DefensePoints: 1, Durability: 2, MaxDurability: 65})

	// First hit: both lose 1 durability.
	a.DamageArmor(1)
	assert.Equal(t, 2, a.GetSlot(SlotHelmet).Durability)
	assert.Equal(t, 1, a.GetSlot(SlotBoots).Durability)

	// Second hit: boots reaches 0, should be removed.
	a.DamageArmor(1)
	assert.Equal(t, 1, a.GetSlot(SlotHelmet).Durability)
	assert.True(t, a.GetSlot(SlotBoots).IsEmpty())

	// Third hit: helmet reaches 0, should be removed.
	a.DamageArmor(1)
	assert.True(t, a.GetSlot(SlotHelmet).IsEmpty())
}

func TestArmor_DamageArmor_SkipsEmpty(t *testing.T) {
	a := Armor{}
	// No armor equipped; should not panic.
	a.DamageArmor(1)
	assert.Equal(t, 0, a.TotalDefense())
}

// ---------------------------------------------------------------------------
// Empty armor = no damage reduction
// ---------------------------------------------------------------------------

func TestArmor_EmptyNoReduction(t *testing.T) {
	a := Armor{}
	result := a.DamageReduction(10)
	assert.Equal(t, float32(10), result)
}

// ---------------------------------------------------------------------------
// DamageSystem integration with armor
// ---------------------------------------------------------------------------

func TestDamageSystem_WithArmor(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})

	body := physics.NewBody(mcmath.AABB{})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})

	// Equip iron chestplate: 6 defense.
	armor := Armor{}
	armor.Equip(SlotChestplate, ArmorPiece{
		ItemID:        405,
		DefensePoints: 6,
		Durability:    240,
		MaxDurability: 240,
	})
	ecs.GetStore[Armor](w).Set(e, armor)

	// Apply 10 raw damage.
	ecs.GetStore[Damage](w).Set(e, Damage{
		Amount:    10,
		Knockback: mcmath.Vec3{X: 0, Y: 5, Z: 0},
	})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	// Expected: 10 * (1 - 6/25) = 10 * 0.76 = 7.6
	h, _ := ecs.GetStore[Health](w).Get(e)
	assert.InDelta(t, 20.0-7.6, float64(h.Current), 0.01)

	// Armor durability should have been reduced by 1.
	a, _ := ecs.GetStore[Armor](w).Get(e)
	assert.Equal(t, 239, a.GetSlot(SlotChestplate).Durability)

	// Damage component should be removed.
	assert.False(t, ecs.GetStore[Damage](w).Has(e))

	// Knockback should still apply.
	pb, _ := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.Equal(t, float32(5), pb.Body.Velocity.Y)
}

func TestDamageSystem_WithoutArmor_Unchanged(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	body := physics.NewBody(mcmath.AABB{})
	ecs.GetStore[PhysicsBody](w).Set(e, PhysicsBody{Body: &body})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 5})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	h, _ := ecs.GetStore[Health](w).Get(e)
	assert.Equal(t, float32(15), h.Current)
}

// ---------------------------------------------------------------------------
// SpawnPlayer includes empty Armor component
// ---------------------------------------------------------------------------

func TestSpawnPlayer_HasArmor(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := SpawnPlayer(w, "Steve", pos)

	a, ok := ecs.GetStore[Armor](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 0, a.TotalDefense())
}

// ---------------------------------------------------------------------------
// ArmorPiece.IsEmpty
// ---------------------------------------------------------------------------

func TestArmorPiece_IsEmpty(t *testing.T) {
	assert.True(t, ArmorPiece{}.IsEmpty())
	assert.False(t, ArmorPiece{ItemID: 400}.IsEmpty())
}
