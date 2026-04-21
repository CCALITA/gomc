package item

// Enchantment represents a single enchantment applied to an item.
type Enchantment struct {
	ID    string // enchantment identifier (e.g. "sharpness", "efficiency")
	Level int    // enchantment level (1-5 typically)
}

// ItemStack represents a quantity of a single item type, optionally with
// remaining durability for tools, enchantments, and a custom display name.
type ItemStack struct {
	ItemID       ItemID
	Count        int
	Durability   int           // current remaining durability; 0 for non-tool items
	Enchantments []Enchantment // enchantments applied to this item; nil if none
	CustomName   string        // custom display name set by anvil; empty if unset
}

// NewItemStack creates an ItemStack for the given item and count.
// If the item has durability, the stack's Durability is initialised to the
// item's maximum durability.
func NewItemStack(id ItemID, count int) ItemStack {
	props := GetProperties(id)
	dur := props.Durability
	return ItemStack{
		ItemID:     id,
		Count:      count,
		Durability: dur,
	}
}

// IsEmpty reports whether the stack contains zero items.
func (s ItemStack) IsEmpty() bool {
	return s.Count <= 0
}

// CanStackWith reports whether s can be merged with other.
// Two stacks are compatible when they share the same ItemID, both are
// stackable (max stack > 1), and tools cannot stack with each other
// because their durability may differ.
func (s ItemStack) CanStackWith(other ItemStack) bool {
	if s.ItemID != other.ItemID {
		return false
	}
	if !IsStackable(s.ItemID) {
		return false
	}
	return true
}

// Merge attempts to combine other into s (filling s up to its max stack).
// It returns the remaining items from other that could not fit.
// If the stacks are incompatible, other is returned unchanged.
func (s *ItemStack) Merge(other ItemStack) ItemStack {
	if !s.CanStackWith(other) {
		return other
	}

	max := MaxStack(s.ItemID)
	space := max - s.Count
	if space <= 0 {
		return other
	}

	transfer := other.Count
	if transfer > space {
		transfer = space
	}

	s.Count += transfer
	return ItemStack{
		ItemID:     other.ItemID,
		Count:      other.Count - transfer,
		Durability: other.Durability,
	}
}

// Split removes up to amount items from the stack and returns both the
// taken stack and the remaining stack.
func (s ItemStack) Split(amount int) (taken, remaining ItemStack) {
	if amount <= 0 {
		return ItemStack{ItemID: s.ItemID, Durability: s.Durability}, s
	}
	if amount >= s.Count {
		return s, ItemStack{ItemID: s.ItemID, Durability: s.Durability}
	}
	taken = ItemStack{
		ItemID:     s.ItemID,
		Count:      amount,
		Durability: s.Durability,
	}
	remaining = ItemStack{
		ItemID:     s.ItemID,
		Count:      s.Count - amount,
		Durability: s.Durability,
	}
	return taken, remaining
}
