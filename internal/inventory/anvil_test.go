package inventory

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/item"
)

func TestCalculateAnvilOutput_RenameOnly(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 250,
	}

	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:      input,
		OutputName: "Excalibur",
	})

	if output.CustomName != "Excalibur" {
		t.Fatalf("expected custom name %q, got %q", "Excalibur", output.CustomName)
	}
	if xpCost != 1 {
		t.Fatalf("expected xp cost 1, got %d", xpCost)
	}
	if output.ItemID != item.IronSword {
		t.Fatalf("expected item ID %d, got %d", item.IronSword, output.ItemID)
	}
}

func TestCalculateAnvilOutput_RepairCombinesDurability(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.IronPickaxe,
		Count:      1,
		Durability: 100,
	}
	material := item.ItemStack{
		ItemID:     item.IronPickaxe,
		Count:      1,
		Durability: 80,
	}

	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:    input,
		Material: material,
	})

	maxDur := item.GetProperties(item.IronPickaxe).Durability
	expectedDur := 100 + 80 + maxDur*12/100
	if expectedDur > maxDur {
		expectedDur = maxDur
	}

	if output.Durability != expectedDur {
		t.Fatalf("expected durability %d, got %d", expectedDur, output.Durability)
	}
	if xpCost < 2 {
		t.Fatalf("expected xp cost >= 2, got %d", xpCost)
	}
}

func TestCalculateAnvilOutput_EnchantmentMerging(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.DiamondSword,
		Count:      1,
		Durability: 1561,
		Enchantments: []item.Enchantment{
			{ID: "sharpness", Level: 2},
			{ID: "fire_aspect", Level: 1},
		},
	}
	material := item.ItemStack{
		ItemID:     item.DiamondSword,
		Count:      1,
		Durability: 1561,
		Enchantments: []item.Enchantment{
			{ID: "sharpness", Level: 3},
			{ID: "looting", Level: 2},
		},
	}

	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:    input,
		Material: material,
	})

	// Sharpness should be upgraded to 3, fire_aspect kept at 1, looting added at 2.
	enchMap := make(map[string]int)
	for _, e := range output.Enchantments {
		enchMap[e.ID] = e.Level
	}

	if enchMap["sharpness"] != 3 {
		t.Fatalf("expected sharpness 3, got %d", enchMap["sharpness"])
	}
	if enchMap["fire_aspect"] != 1 {
		t.Fatalf("expected fire_aspect 1, got %d", enchMap["fire_aspect"])
	}
	if enchMap["looting"] != 2 {
		t.Fatalf("expected looting 2, got %d", enchMap["looting"])
	}

	// XP cost: 2 (repair) + 3 (sharpness upgrade) + 2 (looting new) = 7
	if xpCost != 7 {
		t.Fatalf("expected xp cost 7, got %d", xpCost)
	}
}

func TestCalculateAnvilOutput_IncompatibleItemsRejected(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 250,
	}
	material := item.ItemStack{
		ItemID:     item.IronPickaxe,
		Count:      1,
		Durability: 250,
	}

	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:    input,
		Material: material,
	})

	if !output.IsEmpty() {
		t.Fatalf("expected empty output for incompatible items, got %+v", output)
	}
	if xpCost != 0 {
		t.Fatalf("expected xp cost 0, got %d", xpCost)
	}
}

func TestCalculateAnvilOutput_EmptyInput(t *testing.T) {
	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		OutputName: "Test",
	})

	if !output.IsEmpty() {
		t.Fatalf("expected empty output for empty input")
	}
	if xpCost != 0 {
		t.Fatalf("expected xp cost 0, got %d", xpCost)
	}
}

func TestCalculateAnvilOutput_XPCostScalesWithEnchantments(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.DiamondPickaxe,
		Count:      1,
		Durability: 1561,
	}
	material := item.ItemStack{
		ItemID:     item.DiamondPickaxe,
		Count:      1,
		Durability: 1561,
		Enchantments: []item.Enchantment{
			{ID: "efficiency", Level: 5},
			{ID: "unbreaking", Level: 3},
		},
	}

	_, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:    input,
		Material: material,
	})

	// Repair: 2, efficiency 5: +5, unbreaking 3: +3 = 10
	if xpCost != 10 {
		t.Fatalf("expected xp cost 10, got %d", xpCost)
	}
}

func TestCanRepair_SameToolType(t *testing.T) {
	a := item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 100}
	b := item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 80}

	if !CanRepair(a, b) {
		t.Fatal("expected CanRepair to return true for same tool")
	}
}

func TestCanRepair_DifferentToolType(t *testing.T) {
	a := item.ItemStack{ItemID: item.IronSword, Count: 1, Durability: 250}
	b := item.ItemStack{ItemID: item.IronPickaxe, Count: 1, Durability: 250}

	if CanRepair(a, b) {
		t.Fatal("expected CanRepair to return false for different tool types")
	}
}

func TestCanRepair_NonDurabilityItems(t *testing.T) {
	a := item.ItemStack{ItemID: item.OakPlanks, Count: 1}
	b := item.ItemStack{ItemID: item.OakPlanks, Count: 1}

	if CanRepair(a, b) {
		t.Fatal("expected CanRepair to return false for non-durability items")
	}
}

func TestCalculateAnvilOutput_RepairWithRename(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 100,
	}
	material := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 80,
	}

	output, xpCost := CalculateAnvilOutput(AnvilOperation{
		Input:      input,
		Material:   material,
		OutputName: "Bane",
	})

	if output.CustomName != "Bane" {
		t.Fatalf("expected custom name %q, got %q", "Bane", output.CustomName)
	}
	// Repair (2) + rename (1) = 3
	if xpCost != 3 {
		t.Fatalf("expected xp cost 3, got %d", xpCost)
	}
}

func TestCalculateAnvilOutput_DoesNotMutateInput(t *testing.T) {
	input := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 250,
		Enchantments: []item.Enchantment{
			{ID: "sharpness", Level: 1},
		},
	}
	origEnchLen := len(input.Enchantments)

	material := item.ItemStack{
		ItemID:     item.IronSword,
		Count:      1,
		Durability: 250,
		Enchantments: []item.Enchantment{
			{ID: "looting", Level: 1},
		},
	}

	CalculateAnvilOutput(AnvilOperation{
		Input:    input,
		Material: material,
	})

	if len(input.Enchantments) != origEnchLen {
		t.Fatalf("input enchantments were mutated: expected %d, got %d", origEnchLen, len(input.Enchantments))
	}
}
