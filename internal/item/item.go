// Package item defines item types, properties, and item-stack operations
// for a Minecraft-style voxel game.
package item

// ItemID is the numeric identifier for an item type.
type ItemID = uint16

// ToolType constants describe what kind of tool an item is.
const (
	ToolNone    = "none"
	ToolPickaxe = "pickaxe"
	ToolAxe     = "axe"
	ToolShovel  = "shovel"
	ToolSword   = "sword"
	ToolHoe     = "hoe"
)

// Tool-level constants.
const (
	LevelHand    = 0
	LevelWood    = 1
	LevelStone   = 2
	LevelIron    = 3
	LevelDiamond = 4
)

// ItemProperties describes the static properties of an item type.
type ItemProperties struct {
	Name           string
	MaxStackSize   int
	Durability     int     // 0 = no durability
	IsBlock        bool
	BlockID        uint16  // meaningful only when IsBlock is true
	ToolType       string  // one of the Tool* constants
	ToolLevel      int     // one of the Level* constants
	FoodRestore    int     // hunger points restored when eaten; 0 for non-food
	FoodSaturation float64 // saturation points restored when eaten; 0 for non-food
	ArmorSlot      int     // armor equipment slot (0=helmet, 1=chest, 2=legs, 3=boots); only meaningful when ArmorDefense > 0
	ArmorDefense   int     // defense points provided by this armor piece; 0 for non-armor
	AttackDamage   float32 // base attack damage dealt by this weapon; 0 for non-weapons
	IsCompass      bool    // true for compass items that point to spawn
	IsClock        bool    // true for clock items that show day/night cycle
}

// IsArmor reports whether the item is an armor piece.
func (p ItemProperties) IsArmor() bool {
	return p.ArmorDefense > 0
}

// ----- Block item IDs (items that correspond to placeable blocks) -----

const (
	Air           ItemID = iota // 0 – not a real item, reserved
	Stone                       // 1
	Dirt                        // 2
	Grass                       // 3
	Sand                        // 4
	Gravel                      // 5
	OakLog                      // 6
	OakPlanks                   // 7
	Cobblestone                 // 8
	Glass                       // 9
	OakLeaves                   // 10
	IronOre                     // 11
	GoldOre                     // 12
	DiamondOre                  // 13
	CoalOre                     // 14
	Bedrock                     // 15
	Water                       // 16
	Lava                        // 17
	CraftingTable               // 18
	Furnace                     // 19
	Chest                       // 20
	TNT                         // 21
	Obsidian                    // 22
	Torch                       // 23
)

// ----- Non-block (tool / material) item IDs -----

const (
	WoodenPickaxe  ItemID = iota + 100
	StonePickaxe          // 101
	IronPickaxe           // 102
	DiamondPickaxe        // 103

	WoodenAxe  // 104
	StoneAxe   // 105
	IronAxe    // 106
	DiamondAxe // 107

	WoodenShovel  // 108
	StoneShovel   // 109
	IronShovel    // 110
	DiamondShovel // 111

	WoodenSword  // 112
	StoneSword   // 113
	IronSword    // 114
	DiamondSword // 115

	WoodenHoe  // 116
	StoneHoe   // 117
	IronHoe    // 118
	DiamondHoe // 119

	Stick       // 120
	Coal        // 121
	IronIngot   // 122
	GoldIngot   // 123
	Diamond     // 124
	Bucket      // 125
	WaterBucket // 126
	LavaBucket  // 127
)

// ----- Food item IDs -----

const (
	Apple         ItemID = iota + 200 // 200
	Bread                             // 201
	CookedPorkchop                    // 202
	Steak                             // 203
	GoldenApple                       // 204
	Cookie                            // 205
	Carrot                            // 206
	BakedPotato                       // 207
)

// ----- Utility / building item IDs -----

const (
	Shears       ItemID = 323
	OakDoor      ItemID = 330
	OakFence     ItemID = 331
	OakFenceGate ItemID = 332
	Ladder       ItemID = 334
	Boat         ItemID = 335
	IronBlock    ItemID = 340
	GoldBlock    ItemID = 341
	DiamondBlock ItemID = 342

	// Base materials used in stair/slab crafting
	Sandstone    ItemID = 343
	BirchPlanks  ItemID = 344
	SprucePlanks ItemID = 345
)

// ----- Stair / slab item IDs -----

const (
	OakStairs         ItemID = iota + 350 // 350
	CobblestoneStairs                     // 351
	StoneStairs                           // 352
	BirchStairs                           // 353
	SpruceStairs                          // 354
	SandstoneStairs                       // 355
	OakSlab                               // 356
	CobblestoneSlab                       // 357
	StoneSlab                             // 358
	BirchSlab                             // 359
	SpruceSlab                            // 360
	SandstoneSlab                         // 361
)

// ----- Armor item IDs -----

const (
	LeatherHelmet     ItemID = iota + 400 // 400
	LeatherChestplate                     // 401
	LeatherLeggings                       // 402
	LeatherBoots                          // 403

	IronHelmet     // 404
	IronChestplate // 405
	IronLeggings   // 406
	IronBoots      // 407

	GoldHelmet     // 408
	GoldChestplate // 409
	GoldLeggings   // 410
	GoldBoots      // 411

	DiamondHelmet     // 412
	DiamondChestplate // 413
	DiamondLeggings   // 414
	DiamondBoots      // 415
)

// Crop and farming items.
const (
	WheatSeeds ItemID = 250
	Wheat      ItemID = 251
)

// Functional items.
const (
	Compass      ItemID = 262
	Clock        ItemID = 263
	RedstoneItem ItemID = 264
)

// Furniture items.
const (
	BedItem   ItemID = 370
	SignItem   ItemID = 371
	AnvilItem  ItemID = 372
)

// Decoration items.
const (
	PaintingItem  ItemID = 380
	ItemFrameItem ItemID = 381
)

// Crafting material items.
const (
	Leather ItemID = 128
	Wool    ItemID = 24
)

// Villager trading items.
const (
	Emerald   ItemID = 450
	Paper     ItemID = 451
	Bookshelf ItemID = 452
)

// Fishing items.
const (
	FishingRod   ItemID = 460
	Cod          ItemID = 461
	Salmon       ItemID = 462
	CookedCod    ItemID = 463
	CookedSalmon ItemID = 464
	StringItem   ItemID = 465
	Bowl         ItemID = 466
	Bow          ItemID = 467
	Book         ItemID = 468
	Saddle       ItemID = 469
)

// Ranged combat items.
const (
	Arrow   ItemID = 470
	Feather ItemID = 471
	Flint   ItemID = 472
)

// Farming material items.
const (
	Sugarcane ItemID = 480
)

// Brewing item IDs.
const (
	WaterBottle      ItemID = iota + 500 // 500
	AwkwardPotion                        // 501
	SpeedPotion                          // 502
	StrengthPotion                       // 503
	RegenPotion                          // 504
	PoisonPotion                         // 505
	NetherWart                           // 506
	BlazePowder                          // 507
	SpiderEye                            // 508
	GhastTear                            // 509
	Sugar                                // 510
	BrewingStandItem                     // 511
)
