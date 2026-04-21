package entity

import (
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
)

// Profession identifies the trade specialisation of a villager.
type Profession uint8

// Profession constants.
const (
	Farmer     Profession = 0
	Librarian  Profession = 1
	Blacksmith Profession = 2
	Cleric     Profession = 3
	Butcher    Profession = 4
)

// TradeOffer represents a single villager trade.
type TradeOffer struct {
	InputItem1 item.ItemStack
	InputItem2 item.ItemStack // zero value means no second input
	Output     item.ItemStack
	MaxUses    int
	Uses       int
}

// IsExhausted reports whether the trade has reached its maximum uses.
func (t TradeOffer) IsExhausted() bool {
	return t.Uses >= t.MaxUses
}

// VillagerData holds the profession and trade list for a villager entity.
type VillagerData struct {
	Prof   Profession
	Trades []TradeOffer
}

// DefaultTrades returns the starting trade offers for the given profession.
func DefaultTrades(prof Profession) []TradeOffer {
	switch prof {
	case Farmer:
		return []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.Wheat, Count: 20},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    16,
			},
			{
				InputItem1: item.ItemStack{ItemID: item.Emerald, Count: 1},
				Output:     item.ItemStack{ItemID: item.Bread, Count: 6},
				MaxUses:    12,
			},
		}
	case Librarian:
		return []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.Paper, Count: 24},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    16,
			},
			{
				InputItem1: item.ItemStack{ItemID: item.Emerald, Count: 3},
				Output:     item.ItemStack{ItemID: item.Bookshelf, Count: 1},
				MaxUses:    12,
			},
		}
	case Blacksmith:
		return []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.Coal, Count: 15},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    16,
			},
			{
				InputItem1: item.ItemStack{ItemID: item.Emerald, Count: 5},
				Output:     item.ItemStack{ItemID: item.IronPickaxe, Count: 1},
				MaxUses:    3,
			},
		}
	case Cleric:
		return []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.GoldIngot, Count: 3},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    12,
			},
			{
				InputItem1: item.ItemStack{ItemID: item.Emerald, Count: 4},
				Output:     item.ItemStack{ItemID: item.GoldenApple, Count: 1},
				MaxUses:    8,
			},
		}
	case Butcher:
		return []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.Steak, Count: 14},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    16,
			},
			{
				InputItem1: item.ItemStack{ItemID: item.Emerald, Count: 1},
				Output:     item.ItemStack{ItemID: item.CookedPorkchop, Count: 5},
				MaxUses:    12,
			},
		}
	default:
		return nil
	}
}

// ExecuteTrade attempts to perform the trade at tradeIndex, consuming items
// from playerInv and adding the output. Returns true if the trade succeeded.
func ExecuteTrade(data *VillagerData, tradeIndex int, playerInv *inventory.Inventory) bool {
	if tradeIndex < 0 || tradeIndex >= len(data.Trades) {
		return false
	}
	trade := &data.Trades[tradeIndex]
	if trade.IsExhausted() {
		return false
	}

	// Check that the player has sufficient input items.
	if !hasEnough(playerInv, trade.InputItem1) {
		return false
	}
	if !trade.InputItem2.IsEmpty() && !hasEnough(playerInv, trade.InputItem2) {
		return false
	}

	// Remove input items.
	removeItems(playerInv, trade.InputItem1)
	if !trade.InputItem2.IsEmpty() {
		removeItems(playerInv, trade.InputItem2)
	}

	// Add output to player inventory.
	playerInv.AddItem(trade.Output)
	trade.Uses++
	return true
}

// hasEnough reports whether the inventory contains at least stack.Count items
// of stack.ItemID across all slots.
func hasEnough(inv *inventory.Inventory, stack item.ItemStack) bool {
	return countItem(inv, stack.ItemID) >= stack.Count
}

// countItem returns the total count of the given item across all slots.
func countItem(inv *inventory.Inventory, id item.ItemID) int {
	total := 0
	for i := range inv.Size() {
		slot := inv.GetSlot(i)
		if !slot.IsEmpty() && slot.ItemID == id {
			total += slot.Count
		}
	}
	return total
}

// removeItems removes exactly stack.Count items of stack.ItemID from the
// inventory, consuming from the first matching slots found.
func removeItems(inv *inventory.Inventory, stack item.ItemStack) {
	remaining := stack.Count
	for i := range inv.Size() {
		if remaining <= 0 {
			break
		}
		slot := inv.GetSlot(i)
		if slot.IsEmpty() || slot.ItemID != stack.ItemID {
			continue
		}
		removed := inv.RemoveItem(i, remaining)
		remaining -= removed.Count
	}
}
