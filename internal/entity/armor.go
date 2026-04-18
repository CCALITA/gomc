package entity

// ArmorSlot identifies one of the four equipment slots.
type ArmorSlot int

const (
	SlotHelmet     ArmorSlot = 0
	SlotChestplate ArmorSlot = 1
	SlotLeggings   ArmorSlot = 2
	SlotBoots      ArmorSlot = 3
)

// armorSlotCount is the number of armor slots.
const armorSlotCount = 4

// maxDefensePoints caps the total defense used in damage reduction.
const maxDefensePoints = 20.0

// defenseReductionDivisor is the divisor in the damage-reduction formula.
const defenseReductionDivisor = 25.0

// ArmorPiece represents a single equipped armor item.
type ArmorPiece struct {
	ItemID        uint16
	DefensePoints int
	Durability    int
	MaxDurability int
}

// IsEmpty reports whether the slot has no armor equipped.
func (p ArmorPiece) IsEmpty() bool {
	return p.ItemID == 0
}

// Armor is an ECS component tracking the four armor equipment slots.
type Armor struct {
	Slots [armorSlotCount]ArmorPiece
}

// Equip places an armor piece in the given slot and returns the piece
// that was previously in that slot (zero value if empty).
func (a *Armor) Equip(slot ArmorSlot, piece ArmorPiece) ArmorPiece {
	prev := a.Slots[slot]
	a.Slots[slot] = piece
	return prev
}

// Unequip removes and returns the armor piece from the given slot.
func (a *Armor) Unequip(slot ArmorSlot) ArmorPiece {
	prev := a.Slots[slot]
	a.Slots[slot] = ArmorPiece{}
	return prev
}

// GetSlot returns the armor piece in the given slot.
func (a *Armor) GetSlot(slot ArmorSlot) ArmorPiece {
	return a.Slots[slot]
}

// TotalDefense returns the sum of defense points across all equipped slots.
func (a *Armor) TotalDefense() int {
	total := 0
	for _, piece := range a.Slots {
		total += piece.DefensePoints
	}
	return total
}

// DamageReduction applies the armor defense formula and returns the
// reduced damage amount: rawDamage * (1.0 - min(20, totalDefense) / 25.0).
func (a *Armor) DamageReduction(rawDamage float32) float32 {
	defense := float32(a.TotalDefense())
	if defense > maxDefensePoints {
		defense = maxDefensePoints
	}
	return rawDamage * (1.0 - defense/defenseReductionDivisor)
}

// DamageArmor reduces the durability of every equipped piece by amount.
// Pieces whose durability reaches zero are removed (unequipped).
func (a *Armor) DamageArmor(amount int) {
	for i := range a.Slots {
		if a.Slots[i].IsEmpty() {
			continue
		}
		a.Slots[i].Durability -= amount
		if a.Slots[i].Durability <= 0 {
			a.Slots[i] = ArmorPiece{}
		}
	}
}
