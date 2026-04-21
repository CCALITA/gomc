package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// AnvilOperation describes the inputs to an anvil interaction.
type AnvilOperation struct {
	Input      item.ItemStack // the item being modified
	Material   item.ItemStack // the sacrifice item (empty for rename-only)
	OutputName string         // desired custom name (empty to keep original)
}

// CalculateAnvilOutput computes the result and XP level cost for an anvil
// operation. If the operation is invalid, it returns an empty stack and zero
// cost.
func CalculateAnvilOutput(op AnvilOperation) (output item.ItemStack, xpCost int) {
	if op.Input.IsEmpty() {
		return item.ItemStack{}, 0
	}

	renameOnly := op.Material.IsEmpty()
	if renameOnly && op.OutputName == "" {
		// Nothing to do.
		return item.ItemStack{}, 0
	}

	if renameOnly {
		return calculateRename(op.Input, op.OutputName)
	}

	if !CanRepair(op.Input, op.Material) {
		return item.ItemStack{}, 0
	}

	return calculateRepairAndCombine(op.Input, op.Material, op.OutputName)
}

// CanRepair reports whether the material can be used to repair or combine with
// the input item. Both items must share the same tool type (or both have
// durability and the same item ID for non-tool items like armor).
func CanRepair(input, material item.ItemStack) bool {
	if input.IsEmpty() || material.IsEmpty() {
		return false
	}

	inputProps := item.GetProperties(input.ItemID)
	materialProps := item.GetProperties(material.ItemID)

	// Same item ID always works (combining two identical tools/armor).
	if input.ItemID == material.ItemID {
		return item.HasDurability(input.ItemID)
	}

	// Different items are compatible only when both are tools of the same type.
	if inputProps.ToolType != item.ToolNone && inputProps.ToolType == materialProps.ToolType {
		return true
	}

	return false
}

// calculateRename returns the input with a new custom name. Costs 1 XP level.
func calculateRename(input item.ItemStack, name string) (item.ItemStack, int) {
	out := copyStack(input)
	out.CustomName = name
	return out, 1
}

// calculateRepairAndCombine merges durability and enchantments from material
// onto input. Optionally applies a rename.
func calculateRepairAndCombine(input, material item.ItemStack, name string) (item.ItemStack, int) {
	out := copyStack(input)
	xpCost := 0

	// Repair: combine durability.
	if item.HasDurability(input.ItemID) && item.HasDurability(material.ItemID) {
		maxDur := item.GetProperties(input.ItemID).Durability
		combined := input.Durability + material.Durability
		bonus := maxDur * 12 / 100 // 12% repair bonus
		combined += bonus
		if combined > maxDur {
			combined = maxDur
		}
		out.Durability = combined
		xpCost += 2
	}

	// Combine enchantments from material onto output.
	enchantCost := mergeEnchantments(&out, material.Enchantments)
	xpCost += enchantCost

	// Rename.
	if name != "" {
		out.CustomName = name
		xpCost++
	}

	// Minimum cost is 1.
	if xpCost < 1 {
		xpCost = 1
	}

	return out, xpCost
}

// mergeEnchantments applies enchantments from src onto dst. If dst already has
// the same enchantment at an equal or higher level, the source is skipped.
// If the source level is higher, the source level wins. Returns the XP cost
// contribution from enchantment merging.
func mergeEnchantments(dst *item.ItemStack, src []item.Enchantment) int {
	if len(src) == 0 {
		return 0
	}

	cost := 0
	existing := make(map[string]int, len(dst.Enchantments))
	for i, e := range dst.Enchantments {
		existing[e.ID] = i
	}

	for _, se := range src {
		if idx, ok := existing[se.ID]; ok {
			// Same enchantment: higher level wins.
			if se.Level > dst.Enchantments[idx].Level {
				dst.Enchantments[idx] = se
				cost += se.Level
			}
		} else {
			// New enchantment: append.
			dst.Enchantments = append(dst.Enchantments, se)
			cost += se.Level
			existing[se.ID] = len(dst.Enchantments) - 1
		}
	}

	return cost
}

// copyStack returns a deep copy of the given item stack, including
// enchantments.
func copyStack(s item.ItemStack) item.ItemStack {
	out := item.ItemStack{
		ItemID:     s.ItemID,
		Count:      s.Count,
		Durability: s.Durability,
		CustomName: s.CustomName,
	}
	if len(s.Enchantments) > 0 {
		out.Enchantments = make([]item.Enchantment, len(s.Enchantments))
		copy(out.Enchantments, s.Enchantments)
	}
	return out
}
