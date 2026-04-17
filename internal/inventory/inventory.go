// Package inventory provides inventory management, crafting, and smelting
// for a Minecraft-style voxel game.
package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// Inventory holds a fixed-size collection of item stacks.
type Inventory struct {
	slots []item.ItemStack
}

// NewInventory creates a new Inventory with the given number of slots.
func NewInventory(size int) *Inventory {
	if size < 0 {
		size = 0
	}
	return &Inventory{
		slots: make([]item.ItemStack, size),
	}
}

// Size returns the number of slots in the inventory.
func (inv *Inventory) Size() int {
	return len(inv.slots)
}

// GetSlot returns the item stack at the given index.
// Returns an empty ItemStack if the index is out of bounds.
func (inv *Inventory) GetSlot(index int) item.ItemStack {
	if index < 0 || index >= len(inv.slots) {
		return item.ItemStack{}
	}
	return inv.slots[index]
}

// SetSlot places the given stack into the slot at index.
// Does nothing if the index is out of bounds.
func (inv *Inventory) SetSlot(index int, stack item.ItemStack) {
	if index < 0 || index >= len(inv.slots) {
		return
	}
	inv.slots[index] = stack
}

// AddItem tries to add the given stack to the inventory. It first attempts
// to merge into existing compatible slots, then fills empty slots.
// Returns whatever portion of the stack that could not fit.
func (inv *Inventory) AddItem(stack item.ItemStack) item.ItemStack {
	if stack.IsEmpty() {
		return stack
	}

	// First pass: merge into existing compatible stacks.
	for i := range inv.slots {
		if inv.slots[i].IsEmpty() {
			continue
		}
		stack = inv.slots[i].Merge(stack)
		if stack.IsEmpty() {
			return stack
		}
	}

	// Second pass: place into empty slots.
	for i := range inv.slots {
		if !inv.slots[i].IsEmpty() {
			continue
		}
		max := item.MaxStack(stack.ItemID)
		if stack.Count <= max {
			inv.slots[i] = stack
			return item.ItemStack{}
		}
		inv.slots[i] = item.ItemStack{
			ItemID:     stack.ItemID,
			Count:      max,
			Durability: stack.Durability,
		}
		stack.Count -= max
	}

	return stack
}

// RemoveItem removes up to count items from the slot at index and returns
// the removed stack. Returns an empty stack if the index is invalid or the
// slot is empty.
func (inv *Inventory) RemoveItem(index int, count int) item.ItemStack {
	if index < 0 || index >= len(inv.slots) || count <= 0 {
		return item.ItemStack{}
	}
	slot := inv.slots[index]
	if slot.IsEmpty() {
		return item.ItemStack{}
	}

	taken, remaining := slot.Split(count)
	if remaining.IsEmpty() {
		inv.slots[index] = item.ItemStack{}
	} else {
		inv.slots[index] = remaining
	}
	return taken
}

// FindItem searches for the first slot containing the given item ID.
// Returns the index and true if found, or (0, false) otherwise.
func (inv *Inventory) FindItem(itemID uint16) (int, bool) {
	for i, slot := range inv.slots {
		if !slot.IsEmpty() && slot.ItemID == itemID {
			return i, true
		}
	}
	return 0, false
}

// Clear empties every slot in the inventory.
func (inv *Inventory) Clear() {
	for i := range inv.slots {
		inv.slots[i] = item.ItemStack{}
	}
}

// IsEmpty reports whether all slots are empty.
func (inv *Inventory) IsEmpty() bool {
	for _, slot := range inv.slots {
		if !slot.IsEmpty() {
			return false
		}
	}
	return true
}
