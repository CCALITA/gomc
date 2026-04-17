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
