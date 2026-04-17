package item

// DefaultMaxStackSize is used when no explicit stack size is set.
const DefaultMaxStackSize = 64

// properties maps every known ItemID to its ItemProperties.
var properties map[ItemID]ItemProperties

func init() {
	properties = make(map[ItemID]ItemProperties)

	// ---- block items (stackable, no durability) ----
	registerBlock := func(id ItemID, name string, blockID uint16) {
		properties[id] = ItemProperties{
			Name:         name,
			MaxStackSize: DefaultMaxStackSize,
			IsBlock:      true,
			BlockID:      blockID,
			ToolType:     ToolNone,
		}
	}

	registerBlock(Stone, "Stone", Stone)
	registerBlock(Dirt, "Dirt", Dirt)
	registerBlock(Grass, "Grass", Grass)
	registerBlock(Sand, "Sand", Sand)
	registerBlock(Gravel, "Gravel", Gravel)
	registerBlock(OakLog, "Oak Log", OakLog)
	registerBlock(OakPlanks, "Oak Planks", OakPlanks)
	registerBlock(Cobblestone, "Cobblestone", Cobblestone)
	registerBlock(Glass, "Glass", Glass)
	registerBlock(OakLeaves, "Oak Leaves", OakLeaves)
	registerBlock(IronOre, "Iron Ore", IronOre)
	registerBlock(GoldOre, "Gold Ore", GoldOre)
	registerBlock(DiamondOre, "Diamond Ore", DiamondOre)
	registerBlock(CoalOre, "Coal Ore", CoalOre)
	registerBlock(Bedrock, "Bedrock", Bedrock)
	registerBlock(Water, "Water", Water)
	registerBlock(Lava, "Lava", Lava)
	registerBlock(CraftingTable, "Crafting Table", CraftingTable)
	registerBlock(Furnace, "Furnace", Furnace)
	registerBlock(Chest, "Chest", Chest)
	registerBlock(TNT, "TNT", TNT)
	registerBlock(Obsidian, "Obsidian", Obsidian)
	registerBlock(Torch, "Torch", Torch)

	// ---- tools ----
	registerTool := func(id ItemID, name, toolType string, level, durability int) {
		properties[id] = ItemProperties{
			Name:         name,
			MaxStackSize: 1,
			Durability:   durability,
			ToolType:     toolType,
			ToolLevel:    level,
		}
	}

	// Pickaxes
	registerTool(WoodenPickaxe, "Wooden Pickaxe", ToolPickaxe, LevelWood, 60)
	registerTool(StonePickaxe, "Stone Pickaxe", ToolPickaxe, LevelStone, 132)
	registerTool(IronPickaxe, "Iron Pickaxe", ToolPickaxe, LevelIron, 251)
	registerTool(DiamondPickaxe, "Diamond Pickaxe", ToolPickaxe, LevelDiamond, 1562)

	// Axes
	registerTool(WoodenAxe, "Wooden Axe", ToolAxe, LevelWood, 60)
	registerTool(StoneAxe, "Stone Axe", ToolAxe, LevelStone, 132)
	registerTool(IronAxe, "Iron Axe", ToolAxe, LevelIron, 251)
	registerTool(DiamondAxe, "Diamond Axe", ToolAxe, LevelDiamond, 1562)

	// Shovels
	registerTool(WoodenShovel, "Wooden Shovel", ToolShovel, LevelWood, 60)
	registerTool(StoneShovel, "Stone Shovel", ToolShovel, LevelStone, 132)
	registerTool(IronShovel, "Iron Shovel", ToolShovel, LevelIron, 251)
	registerTool(DiamondShovel, "Diamond Shovel", ToolShovel, LevelDiamond, 1562)

	// Swords
	registerTool(WoodenSword, "Wooden Sword", ToolSword, LevelWood, 60)
	registerTool(StoneSword, "Stone Sword", ToolSword, LevelStone, 132)
	registerTool(IronSword, "Iron Sword", ToolSword, LevelIron, 251)
	registerTool(DiamondSword, "Diamond Sword", ToolSword, LevelDiamond, 1562)

	// Hoes
	registerTool(WoodenHoe, "Wooden Hoe", ToolHoe, LevelWood, 60)
	registerTool(StoneHoe, "Stone Hoe", ToolHoe, LevelStone, 132)
	registerTool(IronHoe, "Iron Hoe", ToolHoe, LevelIron, 251)
	registerTool(DiamondHoe, "Diamond Hoe", ToolHoe, LevelDiamond, 1562)

	// ---- materials ----
	registerMaterial := func(id ItemID, name string) {
		properties[id] = ItemProperties{
			Name:         name,
			MaxStackSize: DefaultMaxStackSize,
			ToolType:     ToolNone,
		}
	}

	registerMaterial(Stick, "Stick")
	registerMaterial(Coal, "Coal")
	registerMaterial(IronIngot, "Iron Ingot")
	registerMaterial(GoldIngot, "Gold Ingot")
	registerMaterial(Diamond, "Diamond")

	// ---- buckets (stack to 16, except empty bucket stacks to 16) ----
	properties[Bucket] = ItemProperties{
		Name:         "Bucket",
		MaxStackSize: 16,
		ToolType:     ToolNone,
	}
	properties[WaterBucket] = ItemProperties{
		Name:         "Water Bucket",
		MaxStackSize: 1,
		ToolType:     ToolNone,
	}
	properties[LavaBucket] = ItemProperties{
		Name:         "Lava Bucket",
		MaxStackSize: 1,
		ToolType:     ToolNone,
	}
}

// GetProperties returns the static properties for the given item ID.
// If the ID is unknown, a zero-value ItemProperties is returned with
// MaxStackSize defaulting to DefaultMaxStackSize.
func GetProperties(id ItemID) ItemProperties {
	if p, ok := properties[id]; ok {
		return p
	}
	return ItemProperties{
		MaxStackSize: DefaultMaxStackSize,
		ToolType:     ToolNone,
	}
}

// IsStackable reports whether the item can be stacked (max stack > 1).
func IsStackable(id ItemID) bool {
	return GetProperties(id).MaxStackSize > 1
}

// IsTool reports whether the item is a tool (has a ToolType other than "none").
func IsTool(id ItemID) bool {
	return GetProperties(id).ToolType != ToolNone
}

// MaxStack returns the maximum stack size for the given item ID.
func MaxStack(id ItemID) int {
	return GetProperties(id).MaxStackSize
}
