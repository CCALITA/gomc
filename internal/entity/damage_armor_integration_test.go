package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
)

// ---------------------------------------------------------------------------
// DamageSystem + Armor integration tests
// ---------------------------------------------------------------------------

func TestDamageArmorIntegration_ReducedDamage(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[Armor](w).Set(e, Armor{
		Slots: [armorSlotCount]ArmorPiece{
			{ItemID: 1, DefensePoints: 2, Durability: 10, MaxDurability: 10}, // helmet
			{ItemID: 2, DefensePoints: 6, Durability: 10, MaxDurability: 10}, // chestplate
			{ItemID: 3, DefensePoints: 5, Durability: 10, MaxDurability: 10}, // leggings
			{ItemID: 4, DefensePoints: 2, Durability: 10, MaxDurability: 10}, // boots
		},
	})
	// Total defense = 15, reduction factor = 1 - 15/25 = 0.4
	// Raw damage 10 -> reduced to 10 * 0.4 = 4
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 10})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.InDelta(t, 16.0, float64(h.Current), 0.01, "health should be 20 - 4 = 16")
}

func TestDamageArmorIntegration_DurabilityDecreases(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[Armor](w).Set(e, Armor{
		Slots: [armorSlotCount]ArmorPiece{
			{ItemID: 1, DefensePoints: 2, Durability: 5, MaxDurability: 10},
			{}, // empty chestplate slot
			{}, // empty leggings slot
			{ItemID: 4, DefensePoints: 2, Durability: 8, MaxDurability: 10},
		},
	})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 6})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	a, ok := ecs.GetStore[Armor](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, 4, a.Slots[SlotHelmet].Durability, "helmet durability should decrease by 1")
	assert.Equal(t, 7, a.Slots[SlotBoots].Durability, "boots durability should decrease by 1")
	// Empty slots stay empty.
	assert.True(t, a.Slots[SlotChestplate].IsEmpty())
	assert.True(t, a.Slots[SlotLeggings].IsEmpty())
}

func TestDamageArmorIntegration_ArmorRemovedAtZeroDurability(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[Armor](w).Set(e, Armor{
		Slots: [armorSlotCount]ArmorPiece{
			{ItemID: 1, DefensePoints: 2, Durability: 1, MaxDurability: 10}, // will break
			{},                                                               // empty
			{ItemID: 3, DefensePoints: 5, Durability: 1, MaxDurability: 10}, // will break
			{ItemID: 4, DefensePoints: 2, Durability: 5, MaxDurability: 10}, // survives
		},
	})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 8})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	a, ok := ecs.GetStore[Armor](w).Get(e)
	assert.True(t, ok)
	assert.True(t, a.Slots[SlotHelmet].IsEmpty(), "helmet should break at durability 0")
	assert.True(t, a.Slots[SlotLeggings].IsEmpty(), "leggings should break at durability 0")
	assert.False(t, a.Slots[SlotBoots].IsEmpty(), "boots should survive with durability > 0")
	assert.Equal(t, 4, a.Slots[SlotBoots].Durability)
}

func TestDamageArmorIntegration_NoArmorFullDamage(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()

	ecs.GetStore[Health](w).Set(e, Health{Current: 20, Max: 20})
	ecs.GetStore[Damage](w).Set(e, Damage{Amount: 7})

	sys := &DamageSystem{}
	sys.Update(w, 0.0)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(13), h.Current, "without armor, full 7 damage should apply")

	// Damage component should be removed after processing.
	assert.False(t, ecs.GetStore[Damage](w).Has(e))
}
