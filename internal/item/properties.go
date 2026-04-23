package item

import "strings"

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
	registerBlock(Wool, "Wool", Wool)

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
	registerTool(WoodenPickaxe, "Wooden Pickaxe", ToolPickaxe, LevelWood, 59)
	registerTool(StonePickaxe, "Stone Pickaxe", ToolPickaxe, LevelStone, 131)
	registerTool(IronPickaxe, "Iron Pickaxe", ToolPickaxe, LevelIron, 250)
	registerTool(DiamondPickaxe, "Diamond Pickaxe", ToolPickaxe, LevelDiamond, 1561)

	// Axes
	registerTool(WoodenAxe, "Wooden Axe", ToolAxe, LevelWood, 59)
	registerTool(StoneAxe, "Stone Axe", ToolAxe, LevelStone, 131)
	registerTool(IronAxe, "Iron Axe", ToolAxe, LevelIron, 250)
	registerTool(DiamondAxe, "Diamond Axe", ToolAxe, LevelDiamond, 1561)

	// Axe attack damage values.
	setAttackDamage := func(id ItemID, dmg float32) {
		p := properties[id]
		p.AttackDamage = dmg
		properties[id] = p
	}
	setAttackDamage(WoodenAxe, 7)
	setAttackDamage(StoneAxe, 9)
	setAttackDamage(IronAxe, 9)
	setAttackDamage(DiamondAxe, 9)

	// Shovels
	registerTool(WoodenShovel, "Wooden Shovel", ToolShovel, LevelWood, 59)
	registerTool(StoneShovel, "Stone Shovel", ToolShovel, LevelStone, 131)
	registerTool(IronShovel, "Iron Shovel", ToolShovel, LevelIron, 250)
	registerTool(DiamondShovel, "Diamond Shovel", ToolShovel, LevelDiamond, 1561)

	// Swords
	registerTool(WoodenSword, "Wooden Sword", ToolSword, LevelWood, 59)
	registerTool(StoneSword, "Stone Sword", ToolSword, LevelStone, 131)
	registerTool(IronSword, "Iron Sword", ToolSword, LevelIron, 250)
	registerTool(DiamondSword, "Diamond Sword", ToolSword, LevelDiamond, 1561)

	// Sword attack damage values.
	setAttackDamage(WoodenSword, 4)
	setAttackDamage(StoneSword, 5)
	setAttackDamage(IronSword, 6)
	setAttackDamage(DiamondSword, 7)

	// Hoes
	registerTool(WoodenHoe, "Wooden Hoe", ToolHoe, LevelWood, 59)
	registerTool(StoneHoe, "Stone Hoe", ToolHoe, LevelStone, 131)
	registerTool(IronHoe, "Iron Hoe", ToolHoe, LevelIron, 250)
	registerTool(DiamondHoe, "Diamond Hoe", ToolHoe, LevelDiamond, 1561)

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
	registerMaterial(Leather, "Leather")
	registerMaterial(RedstoneItem, "Redstone")
	registerMaterial(Gunpowder, "Gunpowder")

	// ---- functional items ----
	properties[Compass] = ItemProperties{
		Name:         "Compass",
		MaxStackSize: DefaultMaxStackSize,
		ToolType:     ToolNone,
		IsCompass:    true,
	}
	properties[Clock] = ItemProperties{
		Name:         "Clock",
		MaxStackSize: DefaultMaxStackSize,
		ToolType:     ToolNone,
		IsClock:      true,
	}

	// ---- decoration items ----
	properties[PaintingItem] = ItemProperties{
		Name:         "Painting",
		MaxStackSize: DefaultMaxStackSize,
		ToolType:     ToolNone,
	}
	properties[ItemFrameItem] = ItemProperties{
		Name:         "Item Frame",
		MaxStackSize: DefaultMaxStackSize,
		ToolType:     ToolNone,
	}

	// ---- shears ----
	properties[Shears] = ItemProperties{
		Name:         "Shears",
		MaxStackSize: 1,
		Durability:   238,
		ToolType:     ToolNone,
	}

	// ---- utility / building blocks ----
	registerBlock(OakDoor, "Oak Door", OakDoor)
	registerBlock(OakFence, "Oak Fence", OakFence)
	registerBlock(OakFenceGate, "Oak Fence Gate", OakFenceGate)
	registerBlock(Ladder, "Ladder", Ladder)
	registerBlock(IronBlock, "Iron Block", IronBlock)
	registerBlock(GoldBlock, "Gold Block", GoldBlock)
	registerBlock(DiamondBlock, "Diamond Block", DiamondBlock)
	registerBlock(AnvilItem, "Anvil", AnvilItem)
	registerBlock(BedItem, "Bed", BedItem)
	registerBlock(SignItem, "Sign", SignItem)

	// ---- boat (non-block, non-tool) ----
	properties[Boat] = ItemProperties{
		Name:         "Boat",
		MaxStackSize: 1,
		ToolType:     ToolNone,
	}

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

	// ---- food items ----
	registerFood := func(id ItemID, name string, foodRestore int, foodSaturation float64) {
		properties[id] = ItemProperties{
			Name:           name,
			MaxStackSize:   DefaultMaxStackSize,
			ToolType:       ToolNone,
			FoodRestore:    foodRestore,
			FoodSaturation: foodSaturation,
		}
	}

	registerFood(Apple, "Apple", 4, 2.4)
	registerFood(Bread, "Bread", 5, 6.0)
	registerFood(CookedPorkchop, "Cooked Porkchop", 8, 12.8)
	registerFood(Steak, "Steak", 8, 12.8)
	registerFood(GoldenApple, "Golden Apple", 4, 9.6)
	registerFood(Cookie, "Cookie", 2, 0.4)
	registerFood(Carrot, "Carrot", 3, 3.6)
	registerFood(BakedPotato, "Baked Potato", 5, 6.0)

	// ---- armor items ----
	registerArmor := func(id ItemID, name string, slot, defense, durability int) {
		properties[id] = ItemProperties{
			Name:         name,
			MaxStackSize: 1,
			Durability:   durability,
			ToolType:     ToolNone,
			ArmorSlot:    slot,
			ArmorDefense: defense,
		}
	}

	// Leather armor
	registerArmor(LeatherHelmet, "Leather Helmet", 0, 1, 55)
	registerArmor(LeatherChestplate, "Leather Chestplate", 1, 3, 80)
	registerArmor(LeatherLeggings, "Leather Leggings", 2, 2, 75)
	registerArmor(LeatherBoots, "Leather Boots", 3, 1, 65)

	// Iron armor
	registerArmor(IronHelmet, "Iron Helmet", 0, 2, 165)
	registerArmor(IronChestplate, "Iron Chestplate", 1, 6, 240)
	registerArmor(IronLeggings, "Iron Leggings", 2, 5, 225)
	registerArmor(IronBoots, "Iron Boots", 3, 2, 195)

	// Gold armor
	registerArmor(GoldHelmet, "Gold Helmet", 0, 2, 77)
	registerArmor(GoldChestplate, "Gold Chestplate", 1, 5, 112)
	registerArmor(GoldLeggings, "Gold Leggings", 2, 3, 105)
	registerArmor(GoldBoots, "Gold Boots", 3, 1, 91)

	// Diamond armor
	registerArmor(DiamondHelmet, "Diamond Helmet", 0, 3, 363)
	registerArmor(DiamondChestplate, "Diamond Chestplate", 1, 8, 528)
	registerArmor(DiamondLeggings, "Diamond Leggings", 2, 6, 495)
	registerArmor(DiamondBoots, "Diamond Boots", 3, 3, 429)

	// ---- villager trading materials ----
	registerMaterial(Emerald, "Emerald")
	registerMaterial(Paper, "Paper")
	registerMaterial(Wheat, "Wheat")
	registerBlock(Bookshelf, "Bookshelf", Bookshelf)

	// ---- brewing items ----
	registerMaterial(WaterBottle, "Water Bottle")
	registerMaterial(AwkwardPotion, "Awkward Potion")
	registerMaterial(SpeedPotion, "Speed Potion")
	registerMaterial(StrengthPotion, "Strength Potion")
	registerMaterial(RegenPotion, "Regen Potion")
	registerMaterial(PoisonPotion, "Poison Potion")
	registerMaterial(NetherWart, "Nether Wart")
	registerMaterial(BlazePowder, "Blaze Powder")
	registerMaterial(SpiderEye, "Spider Eye")
	registerMaterial(GhastTear, "Ghast Tear")
	registerMaterial(Sugar, "Sugar")
	registerBlock(BrewingStandItem, "Brewing Stand", BrewingStandItem)

	// ---- fishing items ----
	properties[FishingRod] = ItemProperties{
		Name:         "Fishing Rod",
		MaxStackSize: 1,
		Durability:   64,
		ToolType:     ToolNone,
	}
	registerFood(Cod, "Cod", 2, 0.4)
	registerFood(Salmon, "Salmon", 2, 0.4)
	registerFood(CookedCod, "Cooked Cod", 5, 6.0)
	registerFood(CookedSalmon, "Cooked Salmon", 6, 9.6)
	registerMaterial(StringItem, "String")
	registerMaterial(Bowl, "Bowl")
	properties[Bow] = ItemProperties{
		Name:         "Bow",
		MaxStackSize: 1,
		Durability:   384,
		ToolType:     ToolNone,
	}
	registerMaterial(Book, "Book")
	properties[Saddle] = ItemProperties{
		Name:         "Saddle",
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

// GetAttackDamage returns the base melee attack damage for the given item.
// Weapons (swords, axes) return their configured AttackDamage; all other items
// (including bare fist / Air) return 1.0.
func GetAttackDamage(id ItemID) float32 {
	props := GetProperties(id)
	if props.AttackDamage > 0 {
		return props.AttackDamage
	}
	return 1.0
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

// HasDurability reports whether the item has durability (tools, armor, shears, etc.).
func HasDurability(id ItemID) bool {
	return GetProperties(id).Durability > 0
}

// IsFood reports whether the item restores hunger when eaten.
func IsFood(id ItemID) bool {
	return GetProperties(id).FoodRestore > 0
}

// GetIDByName performs a case-insensitive search of the properties table
// and returns the ItemID for the first item whose Name matches. Spaces and
// underscores are treated as equivalent so that both "diamond_pickaxe" and
// "Diamond Pickaxe" resolve to the same item. Returns (0, false) when no
// match is found.
func GetIDByName(name string) (ItemID, bool) {
	norm := strings.ToLower(strings.ReplaceAll(name, "_", " "))
	for id, p := range properties {
		if strings.ToLower(p.Name) == norm {
			return id, true
		}
	}
	return 0, false
}
