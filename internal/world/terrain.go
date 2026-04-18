// Package world implements world management, terrain generation, and async
// chunk loading for a Minecraft-like voxel game.
package world

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/biome"
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/noise"
)

const (
	seaLevel       = 62
	caveThreshold  = 0.6
	caveMinY       = 5
	bedrockRandMax = 4
)

// oreConfig describes how a single ore type is generated.
type oreConfig struct {
	blockID   block.BlockID
	maxY      int
	threshold float64
	scale     float64
}

// colInfo holds per-column biome and computed height.
type colInfo struct {
	biome  biome.Biome
	height int
}

// TerrainGenerator produces chunks with realistic terrain using layered noise.
type TerrainGenerator struct {
	seed      int64
	heightGen *noise.OctaveNoise
	caveNoise *noise.OctaveNoise
	treeNoise *noise.OctaveNoise
	biomeMap  *biome.BiomeMap
	ores      []oreConfig
	oreNoises []*noise.OctaveNoise
}

// NewTerrainGenerator creates a TerrainGenerator with the given seed.
// It initialises noise generators for height, caves, trees, biomes, and ores.
func NewTerrainGenerator(seed int64) *TerrainGenerator {
	heightGen := noise.NewNoiseGenerator(seed)
	heightOctave := noise.NewOctaveNoise(heightGen, 6, 0.5, 2.0)

	caveGen := noise.NewNoiseGenerator(seed + 1)
	caveOctave := noise.NewOctaveNoise(caveGen, 4, 0.5, 2.0)

	treeGen := noise.NewNoiseGenerator(seed + 2)
	treeOctave := noise.NewOctaveNoise(treeGen, 2, 0.5, 2.0)

	bm := biome.NewBiomeMap(seed)

	ores := []oreConfig{
		{blockID: block.CoalOre, maxY: 80, threshold: 0.55, scale: 0.08},
		{blockID: block.IronOre, maxY: 64, threshold: 0.6, scale: 0.07},
		{blockID: block.GoldOre, maxY: 32, threshold: 0.7, scale: 0.06},
		{blockID: block.DiamondOre, maxY: 16, threshold: 0.8, scale: 0.05},
	}

	oreNoises := make([]*noise.OctaveNoise, len(ores))
	for i := range ores {
		gen := noise.NewNoiseGenerator(seed + 100 + int64(i))
		oreNoises[i] = noise.NewOctaveNoise(gen, 3, 0.5, 2.0)
	}

	return &TerrainGenerator{
		seed:      seed,
		heightGen: heightOctave,
		caveNoise: caveOctave,
		treeNoise: treeOctave,
		biomeMap:  bm,
		ores:      ores,
		oreNoises: oreNoises,
	}
}

// heightAt computes the terrain height for a column using the biome's parameters.
func (tg *TerrainGenerator) heightAt(wx, wz int, b biome.Biome) int {
	const scale = 0.01
	n := tg.heightGen.Sample2D(float64(wx)*scale, float64(wz)*scale)
	return int(math.Round(b.BaseHeight + n*b.HeightAmplitude))
}

// GenerateChunk generates a full terrain chunk at the given chunk position.
func (tg *TerrainGenerator) GenerateChunk(pos mcmath.ChunkPos) *chunk.Chunk {
	c := &chunk.Chunk{Pos: pos}
	rng := rand.New(rand.NewSource(tg.seed ^ int64(pos.X)*397 ^ int64(pos.Z)*7901))

	var cols [mcmath.ChunkSize][mcmath.ChunkSize]colInfo
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := int(pos.WorldBlockX()) + lx
			wz := int(pos.WorldBlockZ()) + lz
			b := tg.biomeMap.BiomeAt(wx, wz)
			h := tg.heightAt(wx, wz, b)
			cols[lx][lz] = colInfo{biome: b, height: h}
		}
	}

	// Pass 1: base terrain with biome-specific surface and subsurface blocks.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			ci := cols[lx][lz]
			h := ci.height
			for y := 0; y <= h && y < mcmath.ChunkHeight; y++ {
				var id block.BlockID
				switch {
				case y == 0:
					id = block.Bedrock
				case y <= bedrockRandMax && rng.Intn(y+1) == 0:
					id = block.Bedrock
				case y == h:
					if ci.biome.ID == biome.Mountains && y > 90 {
						id = block.Stone
					} else {
						id = ci.biome.SurfaceBlock
					}
				case y >= h-3:
					id = ci.biome.SubsurfaceBlock
				default:
					id = block.Stone
				}
				c.SetBlock(lx, y, lz, id)
			}
		}
	}

	// Pass 2: water at sea level and sand at shorelines.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			ci := cols[lx][lz]
			h := ci.height
			for y := h + 1; y <= seaLevel; y++ {
				if y >= 0 && y < mcmath.ChunkHeight {
					c.SetBlock(lx, y, lz, block.Water)
				}
			}
			if ci.biome.ID != biome.Desert && ci.biome.ID != biome.Ocean {
				if h <= seaLevel && h >= seaLevel-2 {
					for y := h; y >= h-3 && y >= 0; y-- {
						bid := c.GetBlock(lx, y, lz)
						if bid == block.Grass || bid == block.Dirt {
							c.SetBlock(lx, y, lz, block.Sand)
						}
					}
				}
			}
		}
	}

	// Pass 2.5: snow layer on Taiga surface blocks.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			ci := cols[lx][lz]
			if ci.biome.ID != biome.Taiga {
				continue
			}
			h := ci.height
			if h > seaLevel && h+1 < mcmath.ChunkHeight {
				c.SetBlock(lx, h+1, lz, block.Snow)
			}
		}
	}

	// Pass 3: caves.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			h := cols[lx][lz].height
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			for y := caveMinY; y < h && y < mcmath.ChunkHeight; y++ {
				n := tg.caveNoise.Sample3D(wx*0.05, float64(y)*0.05, wz*0.05)
				if n > caveThreshold {
					c.SetBlock(lx, y, lz, block.Air)
				}
			}
		}
	}

	// Pass 4: ores.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			for i, ore := range tg.ores {
				for y := 1; y < ore.maxY && y < mcmath.ChunkHeight; y++ {
					if c.GetBlock(lx, y, lz) != block.Stone {
						continue
					}
					n := tg.oreNoises[i].Sample3D(wx*ore.scale, float64(y)*ore.scale, wz*ore.scale)
					if n > ore.threshold {
						c.SetBlock(lx, y, lz, ore.blockID)
					}
				}
			}
		}
	}

	// Pass 5: trees with biome-specific density.
	tg.generateTrees(c, pos, cols, rng)

	return c
}

// generateTrees places trees on suitable surface blocks using biome-specific
// density and tree types.
func (tg *TerrainGenerator) generateTrees(c *chunk.Chunk, pos mcmath.ChunkPos, cols [mcmath.ChunkSize][mcmath.ChunkSize]colInfo, rng *rand.Rand) {
	for lx := 2; lx < mcmath.ChunkSize-2; lx++ {
		for lz := 2; lz < mcmath.ChunkSize-2; lz++ {
			ci := cols[lx][lz]
			if ci.biome.TreeDensity <= 0 {
				continue
			}
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			n := tg.treeNoise.Sample2D(wx*0.1, wz*0.1)
			if n < 0.4 {
				continue
			}
			if rng.Float64() > ci.biome.TreeDensity*5 {
				continue
			}
			h := ci.height
			if h <= seaLevel || h >= mcmath.ChunkHeight-10 {
				continue
			}
			if c.GetBlock(lx, h, lz) != ci.biome.SurfaceBlock {
				continue
			}
			tg.placeTree(c, lx, h+1, lz, ci.biome.Tree, rng)
		}
	}
}

// treeBlocks returns the log and leaf block IDs for the given tree type.
func treeBlocks(tt biome.TreeType) (log, leaf block.BlockID) {
	switch tt {
	case biome.TreeSpruce:
		return block.SpruceLog, block.SpruceLeaves
	case biome.TreeJungle:
		return block.JungleLog, block.JungleLeaves
	case biome.TreeBirch:
		return block.BirchLog, block.BirchLeaves
	default:
		return block.OakLog, block.OakLeaves
	}
}

// placeTree places a tree at the given local position.
// The trunk is 5-7 blocks tall; leaves form a 5x5x3 canopy at the top.
func (tg *TerrainGenerator) placeTree(c *chunk.Chunk, lx, baseY, lz int, tt biome.TreeType, rng *rand.Rand) {
	logID, leafID := treeBlocks(tt)
	trunkHeight := 5 + rng.Intn(3)
	topY := baseY + trunkHeight - 1

	if topY+3 >= mcmath.ChunkHeight {
		return
	}

	for y := baseY; y <= topY; y++ {
		c.SetBlock(lx, y, lz, logID)
	}

	leafStartY := topY - 1
	for dy := 0; dy < 3; dy++ {
		for dx := -2; dx <= 2; dx++ {
			for dz := -2; dz <= 2; dz++ {
				nx := lx + dx
				nz := lz + dz
				ny := leafStartY + dy
				if nx < 0 || nx >= mcmath.ChunkSize || nz < 0 || nz >= mcmath.ChunkSize {
					continue
				}
				if ny >= mcmath.ChunkHeight {
					continue
				}
				if dy == 2 && mcmath.Abs(dx) == 2 && mcmath.Abs(dz) == 2 {
					continue
				}
				if c.GetBlock(nx, ny, nz) == logID {
					continue
				}
				c.SetBlock(nx, ny, nz, leafID)
			}
		}
	}
}
