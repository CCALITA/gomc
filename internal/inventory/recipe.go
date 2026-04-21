package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// Recipe defines a crafting recipe.
type Recipe struct {
	Pattern   [3][3]uint16   // item IDs; 0 means empty
	Result    item.ItemStack // the output stack
	Shapeless bool           // if true, slot positions do not matter
}

// recipes is the global slice of registered recipes.
var recipes []Recipe

func init() {
	registerRecipes()
}

// RegisterRecipe adds a recipe to the global recipe list.
func RegisterRecipe(r Recipe) {
	recipes = append(recipes, r)
}

// GetRecipes returns a copy of the global recipe list.
func GetRecipes() []Recipe {
	out := make([]Recipe, len(recipes))
	copy(out, recipes)
	return out
}

// registerRecipes populates the global recipe list with all known recipes.
func registerRecipes() {
	// -- OakPlanks: 1 OakLog -> 4 OakPlanks (shapeless)
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakLog, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		Result:    item.NewItemStack(item.OakPlanks, 4),
		Shapeless: true,
	})

	// -- Sticks: 2 OakPlanks (vertical) -> 4 Sticks
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, 0, 0},
			{item.OakPlanks, 0, 0},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.Stick, 4),
	})

	// -- CraftingTable: 4 OakPlanks (2x2) -> 1 CraftingTable
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, item.OakPlanks, 0},
			{item.OakPlanks, item.OakPlanks, 0},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.CraftingTable, 1),
	})

	// -- Pickaxes: 3 material + 2 sticks (T shape)
	registerPickaxe(item.OakPlanks, item.WoodenPickaxe)
	registerPickaxe(item.Cobblestone, item.StonePickaxe)
	registerPickaxe(item.IronIngot, item.IronPickaxe)
	registerPickaxe(item.Diamond, item.DiamondPickaxe)

	// -- Axes: L shape (2 material top-right + 1 material mid-right + 2 sticks)
	registerAxe(item.OakPlanks, item.WoodenAxe)
	registerAxe(item.Cobblestone, item.StoneAxe)
	registerAxe(item.IronIngot, item.IronAxe)
	registerAxe(item.Diamond, item.DiamondAxe)

	// -- Shovels: 1 material + 2 sticks (vertical)
	registerShovel(item.OakPlanks, item.WoodenShovel)
	registerShovel(item.Cobblestone, item.StoneShovel)
	registerShovel(item.IronIngot, item.IronShovel)
	registerShovel(item.Diamond, item.DiamondShovel)

	// -- Swords: 2 material + 1 stick (vertical)
	registerSword(item.OakPlanks, item.WoodenSword)
	registerSword(item.Cobblestone, item.StoneSword)
	registerSword(item.IronIngot, item.IronSword)
	registerSword(item.Diamond, item.DiamondSword)

	// -- Furnace: 8 Cobblestone ring
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Cobblestone, item.Cobblestone, item.Cobblestone},
			{item.Cobblestone, 0, item.Cobblestone},
			{item.Cobblestone, item.Cobblestone, item.Cobblestone},
		},
		Result: item.NewItemStack(item.Furnace, 1),
	})

	// -- Chest: 8 OakPlanks ring
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, item.OakPlanks, item.OakPlanks},
			{item.OakPlanks, 0, item.OakPlanks},
			{item.OakPlanks, item.OakPlanks, item.OakPlanks},
		},
		Result: item.NewItemStack(item.Chest, 1),
	})

	// -- Torch: 1 Coal + 1 Stick (vertical)
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Coal, 0, 0},
			{item.Stick, 0, 0},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.Torch, 4),
	})

	// -- Hoes: 2 material + 2 sticks (T-shape: top row 2 material, then sticks below)
	registerHoe(item.OakPlanks, item.WoodenHoe)
	registerHoe(item.Cobblestone, item.StoneHoe)
	registerHoe(item.IronIngot, item.IronHoe)
	registerHoe(item.Diamond, item.DiamondHoe)

	// -- Bucket: 3 iron ingots V-shape
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.IronIngot, 0, item.IronIngot},
			{0, item.IronIngot, 0},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.Bucket, 1),
	})

	// -- Ladder: sticks in H-pattern -> 3 ladders
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Stick, 0, item.Stick},
			{item.Stick, item.Stick, item.Stick},
			{item.Stick, 0, item.Stick},
		},
		Result: item.NewItemStack(item.Ladder, 3),
	})

	// -- Oak Door: 6 planks in 2x3 -> 3 doors
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, item.OakPlanks, 0},
			{item.OakPlanks, item.OakPlanks, 0},
			{item.OakPlanks, item.OakPlanks, 0},
		},
		Result: item.NewItemStack(item.OakDoor, 3),
	})

	// -- Oak Fence: 2 rows of [plank, stick, plank] -> 3 fences
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, item.Stick, item.OakPlanks},
			{item.OakPlanks, item.Stick, item.OakPlanks},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.OakFence, 3),
	})

	// -- Oak Fence Gate: 2 rows of [stick, plank, stick]
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Stick, item.OakPlanks, item.Stick},
			{item.Stick, item.OakPlanks, item.Stick},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.OakFenceGate, 1),
	})

	// -- Shears: 2 iron ingots diagonal
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{0, item.IronIngot, 0},
			{item.IronIngot, 0, 0},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.Shears, 1),
	})

	// -- Boat: 5 planks U-shape
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.OakPlanks, 0, item.OakPlanks},
			{item.OakPlanks, item.OakPlanks, item.OakPlanks},
			{0, 0, 0},
		},
		Result: item.NewItemStack(item.Boat, 1),
	})

	// -- Mineral Blocks: 9 material in 3x3 -> 1 block
	registerFilledBlock(item.IronIngot, item.IronBlock)
	registerFilledBlock(item.GoldIngot, item.GoldBlock)
	registerFilledBlock(item.Diamond, item.DiamondBlock)

	// -- Block -> Material decomposition (shapeless)
	registerBlockDecomposition(item.IronBlock, item.IronIngot)
	registerBlockDecomposition(item.GoldBlock, item.GoldIngot)
	registerBlockDecomposition(item.DiamondBlock, item.Diamond)

	// -- Compass: 4 Iron Ingots + 1 Redstone (plus pattern)
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{0, item.IronIngot, 0},
			{item.IronIngot, item.RedstoneItem, item.IronIngot},
			{0, item.IronIngot, 0},
		},
		Result: item.NewItemStack(item.Compass, 1),
	})

	// -- Clock: 4 Gold Ingots + 1 Redstone (plus pattern)
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{0, item.GoldIngot, 0},
			{item.GoldIngot, item.RedstoneItem, item.GoldIngot},
			{0, item.GoldIngot, 0},
		},
		Result: item.NewItemStack(item.Clock, 1),
	})

	// -- Painting: 8 Sticks + 1 Wool center
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Stick, item.Stick, item.Stick},
			{item.Stick, item.Wool, item.Stick},
			{item.Stick, item.Stick, item.Stick},
		},
		Result: item.NewItemStack(item.PaintingItem, 1),
	})

	// -- Item Frame: 8 Sticks + 1 Leather center
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.Stick, item.Stick, item.Stick},
			{item.Stick, item.Leather, item.Stick},
			{item.Stick, item.Stick, item.Stick},
		},
		Result: item.NewItemStack(item.ItemFrameItem, 1),
	})

	// -- Anvil: 3 IronBlock on top + 4 IronIngot (T-shape)
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{item.IronBlock, item.IronBlock, item.IronBlock},
			{0, item.IronIngot, 0},
			{item.IronIngot, item.IronIngot, item.IronIngot},
		},
		Result: item.NewItemStack(item.AnvilItem, 1),
	})

	// -- Fishing Rod: 3 Sticks diagonal + 2 String
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{0, 0, item.Stick},
			{0, item.Stick, item.StringItem},
			{item.Stick, 0, item.StringItem},
		},
		Result: item.NewItemStack(item.FishingRod, 1),
	})
}

// registerPickaxe registers a pickaxe recipe: 3 material on top, 2 sticks vertical center.
func registerPickaxe(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, material, material},
			{0, item.Stick, 0},
			{0, item.Stick, 0},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerAxe registers an axe recipe: L-shape with 3 material + 2 sticks.
func registerAxe(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, material, 0},
			{material, item.Stick, 0},
			{0, item.Stick, 0},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerShovel registers a shovel recipe: 1 material + 2 sticks vertical.
func registerShovel(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, 0, 0},
			{item.Stick, 0, 0},
			{item.Stick, 0, 0},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerSword registers a sword recipe: 2 material + 1 stick vertical.
func registerSword(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, 0, 0},
			{material, 0, 0},
			{item.Stick, 0, 0},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerHoe registers a hoe recipe: 2 material on top row + 2 sticks vertical.
func registerHoe(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, material, 0},
			{0, item.Stick, 0},
			{0, item.Stick, 0},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerFilledBlock registers a 3x3 filled grid recipe: 9 material -> 1 block.
func registerFilledBlock(material, result uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{material, material, material},
			{material, material, material},
			{material, material, material},
		},
		Result: item.NewItemStack(result, 1),
	})
}

// registerBlockDecomposition registers a shapeless recipe: 1 block -> 9 material.
func registerBlockDecomposition(block, material uint16) {
	RegisterRecipe(Recipe{
		Pattern: [3][3]uint16{
			{block, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		Result:    item.NewItemStack(material, 9),
		Shapeless: true,
	})
}
