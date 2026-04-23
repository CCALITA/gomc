package inventory

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Armor crafting recipe tests
// ---------------------------------------------------------------------------

func TestCraftLeatherArmor(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(g *CraftingGrid)
		expect uint16
	}{
		{
			name: "LeatherHelmet",
			setup: func(g *CraftingGrid) {
				// Row 0: full
				g.SetSlot(0, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Leather, 1))
				// Row 1: left + right
				g.SetSlot(1, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Leather, 1))
			},
			expect: item.LeatherHelmet,
		},
		{
			name: "LeatherChestplate",
			setup: func(g *CraftingGrid) {
				// Row 0: left + right
				g.SetSlot(0, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Leather, 1))
				// Row 1: full
				g.SetSlot(1, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(1, 1, item.NewItemStack(item.Leather, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Leather, 1))
				// Row 2: full
				g.SetSlot(2, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(2, 1, item.NewItemStack(item.Leather, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.Leather, 1))
			},
			expect: item.LeatherChestplate,
		},
		{
			name: "LeatherLeggings",
			setup: func(g *CraftingGrid) {
				// Row 0: full
				g.SetSlot(0, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Leather, 1))
				// Row 1: left + right
				g.SetSlot(1, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Leather, 1))
				// Row 2: left + right
				g.SetSlot(2, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.Leather, 1))
			},
			expect: item.LeatherLeggings,
		},
		{
			name: "LeatherBoots",
			setup: func(g *CraftingGrid) {
				// Row 0: left + right
				g.SetSlot(0, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Leather, 1))
				// Row 1: left + right
				g.SetSlot(1, 0, item.NewItemStack(item.Leather, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Leather, 1))
			},
			expect: item.LeatherBoots,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g CraftingGrid
			tc.setup(&g)
			result := g.GetResult()
			assert.Equal(t, tc.expect, result.ItemID)
			assert.Equal(t, 1, result.Count)
		})
	}
}

func TestCraftIronArmor(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(g *CraftingGrid)
		expect uint16
	}{
		{
			name: "IronHelmet",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.IronIngot, 1))
			},
			expect: item.IronHelmet,
		},
		{
			name: "IronChestplate",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 1, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(2, 1, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.IronIngot, 1))
			},
			expect: item.IronChestplate,
		},
		{
			name: "IronLeggings",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.IronIngot, 1))
			},
			expect: item.IronLeggings,
		},
		{
			name: "IronBoots",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.IronIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.IronIngot, 1))
			},
			expect: item.IronBoots,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g CraftingGrid
			tc.setup(&g)
			result := g.GetResult()
			assert.Equal(t, tc.expect, result.ItemID)
			assert.Equal(t, 1, result.Count)
		})
	}
}

func TestCraftGoldArmor(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(g *CraftingGrid)
		expect uint16
	}{
		{
			name: "GoldHelmet",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.GoldIngot, 1))
			},
			expect: item.GoldHelmet,
		},
		{
			name: "GoldChestplate",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 1, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(2, 1, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.GoldIngot, 1))
			},
			expect: item.GoldChestplate,
		},
		{
			name: "GoldLeggings",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.GoldIngot, 1))
			},
			expect: item.GoldLeggings,
		},
		{
			name: "GoldBoots",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.GoldIngot, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.GoldIngot, 1))
			},
			expect: item.GoldBoots,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g CraftingGrid
			tc.setup(&g)
			result := g.GetResult()
			assert.Equal(t, tc.expect, result.ItemID)
			assert.Equal(t, 1, result.Count)
		})
	}
}

func TestCraftDiamondArmor(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(g *CraftingGrid)
		expect uint16
	}{
		{
			name: "DiamondHelmet",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Diamond, 1))
			},
			expect: item.DiamondHelmet,
		},
		{
			name: "DiamondChestplate",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 1, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(2, 1, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.Diamond, 1))
			},
			expect: item.DiamondChestplate,
		},
		{
			name: "DiamondLeggings",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 1, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(2, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(2, 2, item.NewItemStack(item.Diamond, 1))
			},
			expect: item.DiamondLeggings,
		},
		{
			name: "DiamondBoots",
			setup: func(g *CraftingGrid) {
				g.SetSlot(0, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(0, 2, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 0, item.NewItemStack(item.Diamond, 1))
				g.SetSlot(1, 2, item.NewItemStack(item.Diamond, 1))
			},
			expect: item.DiamondBoots,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g CraftingGrid
			tc.setup(&g)
			result := g.GetResult()
			assert.Equal(t, tc.expect, result.ItemID)
			assert.Equal(t, 1, result.Count)
		})
	}
}
