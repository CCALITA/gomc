// Package biome implements a Whittaker-style biome classification system
// that maps temperature and rainfall noise to distinct terrain biomes.
package biome

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/noise"
)

// BiomeID is the numeric identifier for a biome type.
type BiomeID uint8

// Predefined biome IDs.
const (
	Plains    BiomeID = 0
	Forest    BiomeID = 1
	Desert    BiomeID = 2
	Taiga     BiomeID = 3
	Jungle    BiomeID = 4
	Swamp     BiomeID = 5
	Mountains BiomeID = 6
	Ocean     BiomeID = 7
)

// TreeType identifies the tree variant placed in a biome.
type TreeType uint8

// Predefined tree types.
const (
	TreeNone   TreeType = 0
	TreeOak    TreeType = 1
	TreeSpruce TreeType = 2
	TreeJungle TreeType = 3
	TreeBirch  TreeType = 4
)

// Biome describes the terrain characteristics for a biome type.
type Biome struct {
	ID              BiomeID
	Name            string
	Temperature     float64
	Rainfall        float64
	SurfaceBlock    uint16
	SubsurfaceBlock uint16
	Tree            TreeType
	TreeDensity     float64
	HeightAmplitude float64
	BaseHeight      float64
}

// biomeRegistry holds all predefined biomes indexed by their ID.
var biomeRegistry = [8]Biome{
	{
		ID: Plains, Name: "Plains",
		Temperature: 0.8, Rainfall: 0.4,
		SurfaceBlock: block.Grass, SubsurfaceBlock: block.Dirt,
		Tree: TreeOak, TreeDensity: 0.02,
		HeightAmplitude: 16, BaseHeight: 64,
	},
	{
		ID: Forest, Name: "Forest",
		Temperature: 0.7, Rainfall: 0.8,
		SurfaceBlock: block.Grass, SubsurfaceBlock: block.Dirt,
		Tree: TreeOak, TreeDensity: 0.15,
		HeightAmplitude: 20, BaseHeight: 66,
	},
	{
		ID: Desert, Name: "Desert",
		Temperature: 2.0, Rainfall: 0.0,
		SurfaceBlock: block.Sand, SubsurfaceBlock: block.Sand,
		Tree: TreeNone, TreeDensity: 0.0,
		HeightAmplitude: 8, BaseHeight: 64,
	},
	{
		ID: Taiga, Name: "Taiga",
		Temperature: 0.05, Rainfall: 0.8,
		SurfaceBlock: block.Grass, SubsurfaceBlock: block.Dirt,
		Tree: TreeSpruce, TreeDensity: 0.12,
		HeightAmplitude: 24, BaseHeight: 66,
	},
	{
		ID: Jungle, Name: "Jungle",
		Temperature: 1.2, Rainfall: 0.9,
		SurfaceBlock: block.Grass, SubsurfaceBlock: block.Dirt,
		Tree: TreeJungle, TreeDensity: 0.2,
		HeightAmplitude: 24, BaseHeight: 64,
	},
	{
		ID: Swamp, Name: "Swamp",
		Temperature: 0.8, Rainfall: 0.9,
		SurfaceBlock: block.Grass, SubsurfaceBlock: block.Dirt,
		Tree: TreeOak, TreeDensity: 0.06,
		HeightAmplitude: 6, BaseHeight: 60,
	},
	{
		ID: Mountains, Name: "Mountains",
		Temperature: 0.2, Rainfall: 0.3,
		SurfaceBlock: block.Stone, SubsurfaceBlock: block.Stone,
		Tree: TreeSpruce, TreeDensity: 0.02,
		HeightAmplitude: 64, BaseHeight: 80,
	},
	{
		ID: Ocean, Name: "Ocean",
		Temperature: 0.5, Rainfall: 0.5,
		SurfaceBlock: block.Sand, SubsurfaceBlock: block.Sand,
		Tree: TreeNone, TreeDensity: 0.0,
		HeightAmplitude: 8, BaseHeight: 40,
	},
}

// ByID returns the Biome for the given BiomeID.
// Returns Plains if the ID is out of range.
func ByID(id BiomeID) Biome {
	if int(id) >= len(biomeRegistry) {
		return biomeRegistry[Plains]
	}
	return biomeRegistry[id]
}

// noiseScale controls the spatial frequency of the biome noise.
// Larger values produce smaller biomes; smaller values produce larger ones.
const noiseScale = 0.005

// BiomeMap uses two noise generators (temperature and rainfall) to select
// biomes via a Whittaker-style classification.
type BiomeMap struct {
	tempNoise     *noise.OctaveNoise
	rainfallNoise *noise.OctaveNoise
}

// NewBiomeMap creates a BiomeMap with deterministic noise from the given seed.
func NewBiomeMap(seed int64) *BiomeMap {
	tempGen := noise.NewNoiseGenerator(seed + 200)
	tempOctave := noise.NewOctaveNoise(tempGen, 4, 0.5, 2.0)

	rainGen := noise.NewNoiseGenerator(seed + 201)
	rainOctave := noise.NewOctaveNoise(rainGen, 4, 0.5, 2.0)

	return &BiomeMap{
		tempNoise:     tempOctave,
		rainfallNoise: rainOctave,
	}
}

// BiomeAt returns the biome at the given world block coordinates.
// It samples temperature and rainfall noise (each normalised to [0,1]),
// then uses a Whittaker-style lookup to select a biome.
func (bm *BiomeMap) BiomeAt(worldX, worldZ int) Biome {
	temp := (bm.tempNoise.Sample2D(float64(worldX)*noiseScale, float64(worldZ)*noiseScale) + 1.0) / 2.0
	rain := (bm.rainfallNoise.Sample2D(float64(worldX)*noiseScale, float64(worldZ)*noiseScale) + 1.0) / 2.0
	return classify(temp, rain)
}

// classify selects a biome from normalised temperature [0,1] and rainfall [0,1]
// using a simplified Whittaker diagram.
func classify(temp, rain float64) Biome {
	switch {
	case temp < 0.2:
		if rain > 0.5 {
			return biomeRegistry[Taiga]
		}
		return biomeRegistry[Mountains]

	case temp < 0.4:
		if rain > 0.7 {
			return biomeRegistry[Taiga]
		}
		if rain < 0.3 {
			return biomeRegistry[Mountains]
		}
		return biomeRegistry[Forest]

	case temp < 0.6:
		if rain > 0.8 {
			return biomeRegistry[Swamp]
		}
		if rain > 0.5 {
			return biomeRegistry[Forest]
		}
		if rain < 0.2 {
			return biomeRegistry[Ocean]
		}
		return biomeRegistry[Plains]

	case temp < 0.8:
		if rain > 0.7 {
			return biomeRegistry[Jungle]
		}
		if rain > 0.4 {
			return biomeRegistry[Plains]
		}
		if rain < 0.2 {
			return biomeRegistry[Ocean]
		}
		return biomeRegistry[Plains]

	default:
		if rain > 0.6 {
			return biomeRegistry[Jungle]
		}
		if rain < 0.3 {
			return biomeRegistry[Desert]
		}
		return biomeRegistry[Plains]
	}
}
