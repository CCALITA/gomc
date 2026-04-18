package block

// properties maps every known BlockID to its BlockProperties.
var properties = map[BlockID]BlockProperties{
	Air: {
		Name: "air", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 0, LightFilter: 0,
	},
	Stone: {
		Name: "stone", Solid: true, Transparent: false,
		Hardness: 1.5, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	Dirt: {
		Name: "dirt", Solid: true, Transparent: false,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 15,
	},
	Grass: {
		Name: "grass", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	Sand: {
		Name: "sand", Solid: true, Transparent: false,
		Hardness: 0.5, BlastResistance: 0.5,
		LightEmission: 0, LightFilter: 15,
	},
	Gravel: {
		Name: "gravel", Solid: true, Transparent: false,
		Hardness: 0.6, BlastResistance: 0.6,
		LightEmission: 0, LightFilter: 15,
	},
	OakLog: {
		Name: "oak_log", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 2,
		LightEmission: 0, LightFilter: 15,
	},
	OakLeaves: {
		Name: "oak_leaves", Solid: true, Transparent: true,
		Hardness: 0.2, BlastResistance: 0.2,
		LightEmission: 0, LightFilter: 1,
	},
	OakPlanks: {
		Name: "oak_planks", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	Cobblestone: {
		Name: "cobblestone", Solid: true, Transparent: false,
		Hardness: 2, BlastResistance: 6,
		LightEmission: 0, LightFilter: 15,
	},
	Glass: {
		Name: "glass", Solid: true, Transparent: true,
		Hardness: 0.3, BlastResistance: 0.3,
		LightEmission: 0, LightFilter: 0,
	},
	Water: {
		Name: "water", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 0, LightFilter: 2,
	},
	Lava: {
		Name: "lava", Solid: false, Transparent: true,
		Hardness: 100, BlastResistance: 100,
		LightEmission: 15, LightFilter: 0,
	},
	IronOre: {
		Name: "iron_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	CoalOre: {
		Name: "coal_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	DiamondOre: {
		Name: "diamond_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	GoldOre: {
		Name: "gold_ore", Solid: true, Transparent: false,
		Hardness: 3, BlastResistance: 3,
		LightEmission: 0, LightFilter: 15,
	},
	Bedrock: {
		Name: "bedrock", Solid: true, Transparent: false,
		Hardness: -1, BlastResistance: 3600000,
		LightEmission: 0, LightFilter: 15,
	},
	CraftingTable: {
		Name: "crafting_table", Solid: true, Transparent: false,
		Hardness: 2.5, BlastResistance: 2.5,
		LightEmission: 0, LightFilter: 15,
	},
	Furnace: {
		Name: "furnace", Solid: true, Transparent: false,
		Hardness: 3.5, BlastResistance: 3.5,
		LightEmission: 0, LightFilter: 15,
	},
	Chest: {
		Name: "chest", Solid: true, Transparent: false,
		Hardness: 2.5, BlastResistance: 2.5,
		LightEmission: 0, LightFilter: 15,
	},
	Torch: {
		Name: "torch", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 14, LightFilter: 0,
	},
	Obsidian: {
		Name: "obsidian", Solid: true, Transparent: false,
		Hardness: 50, BlastResistance: 1200,
		LightEmission: 0, LightFilter: 15,
	},
	Sandstone: {
		Name: "sandstone", Solid: true, Transparent: false,
		Hardness: 0.8, BlastResistance: 0.8,
		LightEmission: 0, LightFilter: 15,
	},
	Fire: {
		Name: "fire", Solid: false, Transparent: true,
		Hardness: 0, BlastResistance: 0,
		LightEmission: 15, LightFilter: 0,
	},
}

// GetProperties returns the BlockProperties for the given block ID.
// If the ID is unknown, a zero-value BlockProperties is returned.
func GetProperties(id BlockID) BlockProperties {
	return properties[id]
}

// IsSolid reports whether the block with the given ID is solid.
func IsSolid(id BlockID) bool {
	return properties[id].Solid
}

// IsTransparent reports whether the block with the given ID is transparent.
func IsTransparent(id BlockID) bool {
	return properties[id].Transparent
}

// IsLightSource reports whether the block with the given ID emits light.
func IsLightSource(id BlockID) bool {
	return properties[id].LightEmission > 0
}

// IsLava reports whether the block with the given ID is lava.
func IsLava(id BlockID) bool {
	return id == Lava
}

// IsWater reports whether the block with the given ID is water.
func IsWater(id BlockID) bool {
	return id == Water
}

// IsFire reports whether the block with the given ID is fire.
func IsFire(id BlockID) bool {
	return id == Fire
}
